package shadow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/middlewares"
	"github.com/squarefactory/cloud-burster/utils/try"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type DataSource struct {
	http.Client
	username string
	password string
	zone     string
	sshKey   string
}

const (
	requestStorage       = "https://api.shdw-ws.fr/api/block_device/request"
	listStorage          = "https://api.shdw-ws.fr/api/block_device/list"
	requestNode          = "https://api.shdw-ws.fr/api/vm/request"
	listNode             = "https://api.shdw-ws.fr/api/vm/list"
	killNode             = "https://api.shdw-ws.fr/api/vm/kill"
	releaseStorage       = "https://api.shdw-ws.fr/api/block_device/release"
	BlockDeviceAllocated = 2
)

func New(
	username string,
	password string,
	zone string,
	sshKey string,
) *DataSource {
	s := &DataSource{
		username: username,
		password: password,
		zone:     zone,
		sshKey:   sshKey,
	}
	s.Client.Transport = &middlewares.RoundTripper{
		RoundTripper: http.DefaultTransport,
	}
	return s
}

// Create a shadow instance
func (s *DataSource) Create(
	ctx context.Context,
	host *config.Host,
	cloud *config.Cloud,
) error {
	logger.I.Debug(
		"Create called",
		zap.Any("host", host),
		zap.Any("cloud", cloud),
	)

	// Create block device
	logger.I.Info("creating a block device")
	StorageUUID, err := s.CreateBlockDevice(ctx, host)
	if err != nil {
		logger.I.Error("failed to create block device", zap.Error(err))
		return err
	}

	// Wait for block device to be allocated
	_, err = try.Do(func() (string, error) {

		Filter := struct {
			UUID string `json:"uuid"`
		}{UUID: StorageUUID}

		requestBody := struct {
			Filters interface{} `json:"filters"`
		}{Filters: Filter}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return "", err
		}

		resp, err := s.InterrogateAPI(ctx, listStorage, jsonBody)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			logger.I.Error(
				"shadow API returned non-ok code",
				zap.Int("status code", resp.StatusCode),
				zap.String("body", string(body)),
			)
			return "", fmt.Errorf("shadow API returned non-ok code: %d", resp.StatusCode)
		}

		var response BlockDeviceListResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return "", err
		}

		if response.BlockDevices[0].Status != BlockDeviceAllocated {
			return "", errors.New("block device has not been allocated yet")
		}

		return "", nil
	}, 10, 2*time.Second)

	if err != nil {
		logger.I.Error("failed to find block device status", zap.Error(err))
		return err
	}

	logger.I.Info("block device allocated, creating vm", zap.String("device uuid", StorageUUID))

	// Create VM
	NodeUUID, err := s.CreateVM(ctx, host, cloud, StorageUUID)
	if err != nil {
		logger.I.Error("failed to create vm", zap.Error(err))
		return err
	}

	// Fetch public IP for provisioning
	VM, err := try.Do(func() (VM, error) {

		Filter := struct {
			UUID string `json:"uuid"`
		}{
			UUID: NodeUUID,
		}
		requestBody := struct {
			Filters interface{} `json:"filters"`
		}{
			Filters: Filter,
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return VM{}, err
		}

		resp, err := s.InterrogateAPI(ctx, listNode, jsonBody)
		if err != nil {
			return VM{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			logger.I.Error(
				"shadow API returned non-ok code",
				zap.Int("status code", resp.StatusCode),
				zap.String("body", string(body)),
			)
			return VM{}, fmt.Errorf("shadow API returned non-ok code: %d", resp.StatusCode)
		}

		var response VMListResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return VM{}, err
		}

		if response.VMs[0].VMPublicIPv4 == "" && response.VMs[0].VMPublicSSHPort == 0 {
			return VM{}, errors.New("instance has not been assigned an ip yet")
		}

		return response.VMs[0], nil
	}, 10, 10*time.Second)

	if err != nil {
		logger.I.Error("failed to find public IP", zap.Error(err))
		return err
	}

	logger.I.Info(
		"instance has been assigned an ip, generating config",
		zap.String("ip", VM.VMPublicIPv4),
		zap.Int("port", VM.VMPublicSSHPort),
	)

	// Generate config
	userData, err := GenerateCloudConfig(&CloudConfigOpts{
		PostScripts: cloud.PostScripts,
		Hostname:    host.Name,
	})
	if err != nil {
		return err
	}

	logger.I.Info("generated config, spamming ssh")
	if err := s.ExecutePostcript(ctx, host, VM, userData); err != nil {
		logger.I.Error("failed to execute postcript", zap.Error(err))
		return err
	}

	logger.I.Info("spawned a server", zap.Any("vm", VM))
	return nil
}

// CreateBlockDevice creates a storage volume and returns its UUID
func (s *DataSource) CreateBlockDevice(ctx context.Context, host *config.Host) (string, error) {

	blockDevice := struct {
		DatacenterLabel string `json:"datacenter_label"`
		Size            int    `json:"size_gib"`
	}{
		DatacenterLabel: s.zone,
		Size:            host.DiskSize,
	}

	requestBody := struct {
		DryRun      bool        `json:"dry_run"`
		BlockDevice interface{} `json:"block_device"`
	}{
		DryRun:      false,
		BlockDevice: blockDevice,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	resp, err := s.InterrogateAPI(ctx, requestStorage, jsonBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		logger.I.Error(
			"shadow API returned non-ok code",
			zap.Int("status code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return "", fmt.Errorf("shadow API returned non-ok code: %d", resp.StatusCode)
	}

	var response struct {
		BlockDevice struct {
			Cost int    `json:"cost"`
			Size string `json:"gib_size"`
			UUID string `json:"uuid"`
		} `json:"block_device"`
		DryRun bool `json:"dry_run"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.BlockDevice.UUID, nil
}

// CreateVM spawns a VM attached to a storage volume and returns its UUID
func (s *DataSource) CreateVM(
	ctx context.Context,
	host *config.Host,
	cloud *config.Cloud,
	blockDeviceUUID string,
) (string, error) {
	type RequestBodyVM struct {
		SKU           string      `json:"sku"`
		RAM           int         `json:"ram"`
		GPU           int         `json:"gpu"`
		Image         string      `json:"image"`
		BlockDevice   interface{} `json:"block_devices"`
		Interruptible bool        `json:"interruptible"`
	}

	type RequestBody struct {
		DryRun bool          `json:"dry_run"`
		VM     RequestBodyVM `json:"vm"`
	}

	type BlockDevice struct {
		UUID string `json:"uuid"`
	}

	Device := BlockDevice{UUID: blockDeviceUUID}

	url, err := url.Parse(host.ImageName)
	if err != nil {
		return "", fmt.Errorf("url failed to parse: %w", err)
	}
	q := url.Query()
	q.Add("hostname", host.Name)
	url.RawQuery = q.Encode()
	fmt.Println(url)

	// TODO: do not hardcode resources
	requestBody := RequestBody{
		DryRun: false,
		VM: RequestBodyVM{
			SKU:           host.FlavorName,
			RAM:           128,
			GPU:           1,
			Image:         url.String(),
			BlockDevice:   []BlockDevice{Device},
			Interruptible: true,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	resp, err := s.InterrogateAPI(ctx, requestNode, jsonBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		logger.I.Error(
			"shadow API returned non-ok code",
			zap.Int("status code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		if err := s.DeleteBlockDevice(ctx, DeleteBlockDeviceRequest{blockDeviceUUID}); err != nil {
			logger.I.Error(err.Error())
		}
		return "", fmt.Errorf("shadow API returned non-ok code: %d", resp.StatusCode)
	}

	var response struct {
		DryRun bool `json:"dry_run"`
		VM     struct {
			UUID string `json:"uuid"`
		} `json:"vm"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		if err := s.DeleteBlockDevice(ctx, DeleteBlockDeviceRequest{blockDeviceUUID}); err != nil {
			logger.I.Error(err.Error())
		}
		return "", err
	}

	return response.VM.UUID, nil
}

func (s *DataSource) FindVM(ctx context.Context, name string) (VM, error) {
	requestBody := struct {
		Filters struct {
			UUID *string `json:"uuid,omitempty"`
		} `json:"filters"`
	}{}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return VM{}, err
	}

	resp, err := s.InterrogateAPI(ctx, listNode, jsonBody)
	if err != nil {
		return VM{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		logger.I.Error(
			"shadow API returned non-ok code",
			zap.Int("status code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return VM{}, fmt.Errorf("shadow API returned non-ok code: %d", resp.StatusCode)
	}

	// list to get block devices uuid
	var response VMListResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return VM{}, err
	}

	if len(response.VMs) <= 0 {
		return VM{}, fmt.Errorf("VM not found: %v", name)
	}
	var vm VM
	var insertTime time.Time
	for _, v := range response.VMs {
		vInsertTime, err := time.Parse(time.RFC3339, v.InsertedOn)
		if err != nil {
			logger.I.Error("failed to parse insert time", zap.Error(err))
			continue
		}
		url, err := url.Parse(v.Image)
		if err != nil {
			logger.I.Error("failed to parse url", zap.Error(err))
			continue
		}
		if url.Query().Get("hostname") == name && vInsertTime.After(insertTime) {
			vm = v
			insertTime = vInsertTime
		}
	}
	if len(vm.BlockDevices) <= 0 {
		return VM{}, fmt.Errorf("BlockDevice not found: %v", vm)
	}
	logger.I.Debug("found VM", zap.Any("vm", vm))
	return vm, nil
}

// Delete a server
func (s *DataSource) Delete(ctx context.Context, name string) error {
	logger.I.Warn("Delete called", zap.String("name", name))

	vm, err := s.FindVM(ctx, name)
	if err != nil {
		return err
	}

	// kill VM
	VM := struct {
		UUID string `json:"uuid"`
	}{
		UUID: vm.UUID,
	}

	requestBodyNode := struct {
		DryRun bool        `json:"dry_run"`
		VM     interface{} `json:"vm"`
	}{
		DryRun: false,
		VM:     VM,
	}

	jsonBody, err := json.Marshal(requestBodyNode)
	if err != nil {
		return err
	}

	resp, err := s.InterrogateAPI(ctx, killNode, jsonBody)
	if err != nil {
		logger.I.Error("failed to kill node", zap.Error(err))
	} else {
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			logger.I.Error(
				"shadow API returned non-ok code",
				zap.Int("status code", resp.StatusCode),
				zap.String("body", string(body)),
			)
			logger.I.Error("shadow API returned non-ok code", zap.Int("status", resp.StatusCode))
		}
	}

	// release storage
	time.Sleep(10 * time.Second)

	Block := struct {
		UUID string `json:"uuid"`
	}{
		UUID: vm.BlockDevices[0].UUID,
	}

	if err := s.DeleteBlockDevice(ctx, Block); err != nil {
		return err
	}

	logger.I.Warn("Deleted a server", zap.Any("name", name))
	return nil

}

type DeleteBlockDeviceRequest struct {
	UUID string `json:"uuid"`
}

func (s *DataSource) DeleteBlockDevice(ctx context.Context, block DeleteBlockDeviceRequest) error {
	requestBodyDev := struct {
		DryRun bool        `json:"dry_run"`
		Device interface{} `json:"block_device"`
	}{
		DryRun: false,
		Device: block,
	}

	jsonBody, err := json.Marshal(requestBodyDev)
	if err != nil {
		return err
	}

	resp, err := s.InterrogateAPI(ctx, releaseStorage, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		logger.I.Error(
			"shadow API returned non-ok code",
			zap.Int("status code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return fmt.Errorf("shadow API returned non-ok code: %d", resp.StatusCode)
	}

	return nil
}

func (s *DataSource) ExecutePostcript(
	ctx context.Context,
	host *config.Host,
	instance VM,
	userData []byte,
) error {
	// Parse the private key
	pk, err := base64.StdEncoding.DecodeString(s.sshKey)
	if err != nil {
		return err
	}
	signer, err := ssh.ParsePrivateKey(pk)
	if err != nil {
		return err
	}

	// SSH client configuration
	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	// Connect to the SSH server
	d := net.Dialer{Timeout: config.Timeout}
	addr := fmt.Sprintf("%s:%d", instance.VMPublicIPv4, instance.VMPublicSSHPort)

	out, err := try.Do(func() ([]byte, error) {
		// Healthcheck VM
		health, err := s.FindVM(ctx, host.Name)
		if err != nil {
			return nil, err
		}
		// Is Terminated ?
		if health.Status == 3 {
			logger.I.Warn("VM seems terminated", zap.Error(err))
		}
		c, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, err
		}
		defer c.Close()
		conn, chans, reqs, err := ssh.NewClientConn(c, addr, config)
		if err != nil {
			return nil, err
		}
		client := ssh.NewClient(conn, chans, reqs)

		// Create a new SSH session
		session, err := client.NewSession()
		if err != nil {
			return nil, err
		}
		defer session.Close()

		logger.I.Info("ssh connection successful, executing postcripts")
		// Create a temporary bash script file
		out, err := session.CombinedOutput(string(userData))
		if err != nil {
			logger.I.Error("postscripts failed", zap.Error(err), zap.String("out", string(out)))
			return nil, err
		}

		return out, nil
	}, 20, 20*time.Second)

	if err != nil {
		logger.I.Error("failed to execute postcripts", zap.Error(err))
		return err
	}

	logger.I.Info("successfully executed postcript", zap.Any("out", out))
	return nil
}

func (s *DataSource) InterrogateAPI(
	ctx context.Context,
	endpoint string,
	jsonBody []byte,
) (*http.Response, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		endpoint,
		strings.NewReader(string(jsonBody)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)),
	)
	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	return s.Do(req)
}
