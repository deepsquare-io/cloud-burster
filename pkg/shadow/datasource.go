package shadow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/utils/try"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type DataSource struct {
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
	return &DataSource{
		username: username,
		password: password,
		zone:     zone,
		sshKey:   sshKey,
	}
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

	logger.I.Info("instance has been assigned an ip, generating config", zap.String("ip", VM.VMPublicIPv4))

	// Generate config
	userData, err := GenerateCloudConfig(&CloudConfigOpts{
		PostScripts: cloud.PostScripts,
		Hostname:    host.Name,
	})
	if err != nil {
		return err
	}

	if err := s.ExecutePostcript(ctx, VM, userData); err != nil {
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

	type BlockDevice struct {
		UUID string `json:"uuid"`
	}

	Device := BlockDevice{UUID: blockDeviceUUID}

	Instance := struct {
		SKU         string      `json:"sku"`
		RAM         int         `json:"ram"`
		GPU         int         `json:"gpu"`
		Image       string      `json:"image"`
		BlockDevice interface{} `json:"block_devices"`
	}{
		SKU:         host.FlavorName,
		RAM:         112,
		GPU:         1,
		Image:       host.ImageName,
		BlockDevice: []BlockDevice{Device},
	}

	requestBody := struct {
		DryRun bool        `json:"dry_run"`
		VM     interface{} `json:"vm"`
	}{
		DryRun: false,
		VM:     Instance,
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

	var response struct {
		DryRun bool `json:"dry_run"`
		VM     struct {
			UUID string `json:"uuid"`
		} `json:"vm"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.VM.UUID, nil
}

// Delete a server
func (s *DataSource) Delete(ctx context.Context, NodeUUID string) error {

	logger.I.Warn("Delete called", zap.String("uuid", NodeUUID))

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
		return err
	}

	resp, err := s.InterrogateAPI(ctx, listNode, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// list to get block devices uuid

	var response VMListResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	storageUUID := response.VMs[0].BlockDevices[0].UUID

	// kill VM
	VM := struct {
		UUID string `json:"uuid"`
	}{
		UUID: NodeUUID,
	}

	requestBodyNode := struct {
		DryRun bool        `json:"dry_run"`
		VM     interface{} `json:"vm"`
	}{
		DryRun: false,
		VM:     VM,
	}

	jsonBody, err = json.Marshal(requestBodyNode)
	if err != nil {
		return err
	}

	resp, err = s.InterrogateAPI(ctx, killNode, jsonBody)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// release storage
	time.Sleep(10 * time.Second)

	Block := struct {
		UUID string `json:"uuid"`
	}{
		UUID: storageUUID,
	}

	requestBodyDev := struct {
		DryRun bool        `json:"dry_run"`
		Device interface{} `json:"block_device"`
	}{
		DryRun: false,
		Device: Block,
	}

	jsonBody, err = json.Marshal(requestBodyDev)
	if err != nil {
		return err
	}

	resp, err = s.InterrogateAPI(ctx, releaseStorage, jsonBody)
	if err != nil {
		return err
	}
	resp.Body.Close()

	logger.I.Warn("Deleted a server", zap.Any("uuid", NodeUUID))
	return nil
}

func (s *DataSource) ExecutePostcript(ctx context.Context, instance VM, userData []byte) error {
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

	session, err := try.Do(func() (*ssh.Session, error) {
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

		return session, nil
	}, 10, 2*time.Second)

	if err != nil {
		logger.I.Error("failed to establish ssh session", zap.Error(err))
	}

	defer session.Close()
	// Create a temporary bash script file
	out, err := session.CombinedOutput(cloudConfigTemplate)
	if err != nil {
		logger.I.Error("postscripts failed", zap.Error(err), zap.String("out", string(out)))
	}
	return err
}

func (s *DataSource) InterrogateAPI(ctx context.Context, endpoint string, jsonBody []byte) (*http.Response, error) {

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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to interrogate api: responded with %d status code", resp.StatusCode)
	}

	return resp, nil
}
