package shadow

import (
	"bytes"
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
	requestStorage = "https://api.shdw-ws.fr/api/block_device/request"
	requestNode    = "https://api.shdw-ws.fr/api/vm/request"
	listNode       = "https://api.shdw-ws.fr/api/vm/list"
	killNode       = "https://api.shdw-ws.fr/api/vm/kill"
	releaseStorage = "https://api.shdw-ws.fr/api/block_device/release"
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

// Create a shadow instance and returns its public IP
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

	time.Sleep(5 * time.Second)

	// Create VM
	NodeUUID, err := s.CreateVM(ctx, host, cloud, StorageUUID)
	if err != nil {
		logger.I.Error("failed to create vm", zap.Error(err))
		return err
	}

	// Fetch public IP for provisioning
	VM, err := try.Do(func() (VM, error) {
		requestBody := fmt.Sprintf(`{
			"filters": {
				"uuid": "%s"
			}
		}`, NodeUUID)

		req, err := http.NewRequestWithContext(
			ctx,
			"POST",
			listNode,
			strings.NewReader(requestBody),
		)
		if err != nil {
			return VM{}, err
		}

		req.Header.Set(
			"Authorization",
			"Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)),
		)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return VM{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return VM{}, errors.New("failed to find public ip")
		}

		var response ListResponse
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

	fmt.Println(VM)
	// Generate config
	userData, err := GenerateCloudConfig(&CloudConfigOpts{
		PostScripts: cloud.PostScripts,
		Hostname:    host.Name,
	})
	if err != nil {
		return err
	}

	if err := s.executePostcript(ctx, VM, userData); err != nil {
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

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		requestStorage,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed to create block device: api responded with %d status code",
			resp.StatusCode,
		)
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

	req, err := http.NewRequestWithContext(ctx, "POST", requestNode, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.I.Error("failed to create VM", zap.Any("string", resp.Body))
		return "", fmt.Errorf("failed to create VM: api responded with %d error code", resp.StatusCode)
	}

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

	// list to get block devices uuid
	requestBody := fmt.Sprintf(`{
		"filters": {
			"uuid": "%s"
		}
	}`, NodeUUID)

	req, err := http.NewRequestWithContext(ctx, "POST", listNode, strings.NewReader(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to list vms")
	}

	var response ListResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	storageUUID := response.VMs[0].BlockDevices[0].UUID

	// kill VM
	requestBody = fmt.Sprintf(`{
		"dry_run": false,
		"vm": {
			"uuid": "%s"
		}
	}`, NodeUUID)

	req, err = http.NewRequestWithContext(ctx, "POST", killNode, strings.NewReader(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)),
	)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to kill VM")
	}

	time.Sleep(10 * time.Second)
	// release storage
	requestBody = fmt.Sprintf(`{
		"dry_run": false,
		"block_device": {
			"uuid": "%s"
		}
	}`, storageUUID)

	req, err = http.NewRequestWithContext(
		ctx,
		"POST",
		releaseStorage,
		strings.NewReader(requestBody),
	)
	if err != nil {
		return err
	}

	req.Header.Set(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)),
	)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to release storage")
	}

	logger.I.Warn("deleted a server", zap.Any("uuid", NodeUUID))
	return nil
}

func (s *DataSource) executePostcript(ctx context.Context, instance VM, userData []byte) error {
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
	c, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer c.Close()
	conn, chans, reqs, err := ssh.NewClientConn(c, addr, config)
	if err != nil {
		return err
	}
	client := ssh.NewClient(conn, chans, reqs)

	// Create a new SSH session
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Create a temporary bash script file
	out, err := session.CombinedOutput(cloudConfigTemplate)
	if err != nil {
		logger.I.Error("postscripts failed", zap.Error(err), zap.String("out", string(out)))
	}
	return err
}
