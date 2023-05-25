package shadow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

type VM struct {
	PublicIP string `json:"vm_public_ipv4"`
	SSHPort  string `json:"vm_public_sshport"`
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

		req, err := http.NewRequestWithContext(ctx, "POST", listNode, strings.NewReader(requestBody))
		if err != nil {
			return VM{}, err
		}

		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return VM{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return VM{}, errors.New("failed to find public ip")
		}
		var response struct {
			Filters bool `json:"filters"`
			VM      VM   `json:"vms"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return VM{}, err
		}

		return response.VM, nil
	}, 10, 5*time.Second)

	if err != nil {
		logger.I.Error("failed to find public IP", zap.Error(err))
		return err
	}

	// Generate config
	userData, err := GenerateCloudConfig(&CloudConfigOpts{
		PostScripts: cloud.PostScripts,
		Hostname:    host.Name,
	})
	if err != nil {
		return err
	}

	if err := s.executePostcript(VM, userData); err != nil {
		logger.I.Error("failed to execute postcript", zap.Error(err))
		return err
	}

	logger.I.Info("spawned a server", zap.Any("vm", VM))
	return nil
}

// CreateBlockDevice creates a storage volume and returns its UUID
func (s *DataSource) CreateBlockDevice(ctx context.Context, host *config.Host) (string, error) {
	requestBody := fmt.Sprintf(`{
		"dry_run": false,
		"block_device": {
			"datacenter_label": "%s",
			"size_gib": %d
		}
	}`, s.zone, host.DiskSize)

	req, err := http.NewRequestWithContext(ctx, "POST", requestStorage, strings.NewReader(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to create block device")
	}

	var response struct {
		BlockDevice struct {
			Cost string `json:"cost"`
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
func (s *DataSource) CreateVM(ctx context.Context, host *config.Host, cloud *config.Cloud, blockDeviceUUID string) (string, error) {
	requestBody := fmt.Sprintf(`{
		"dry_run": false,
		"pubkeys": [
			%s
		],
		"vm": {
			"sku": "%s",
			"ram": 112,
			"gpu": 1,
			"image": "%s",
			"block_devices": [
				{
					"uuid": "%s"
				}
			],
			"vnc": true
		}
	// }`, strings.Join(formatPubKeys(cloud.AuthorizedKeys), ",\n"), host.FlavorName, host.ImageName, blockDeviceUUID)

	req, err := http.NewRequestWithContext(ctx, "POST", requestNode, strings.NewReader(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to create VM")
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

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to list vms")
	}

	var response struct {
		Filters bool `json:"filters"`
		VM      struct {
			BlockDevice struct {
				StorageUUID string `json:"uuid"`
			} `json:"block_devices"`
		} `json:"vms"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	storageUUID := response.VM.BlockDevice.StorageUUID

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

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to kill VM")
	}

	// release storage
	requestBody = fmt.Sprintf(`{
		"block_device": {
			"uuid": "%s"
		},
		"dry_run": false
	}`, storageUUID)

	req, err = http.NewRequestWithContext(ctx, "POST", releaseStorage, strings.NewReader(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.username+":"+s.password)))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to kill VM")
	}

	logger.I.Warn("deleted a server", zap.Any("uuid", NodeUUID))
	return nil
}

func (s *DataSource) executePostcript(Instance VM, userData []byte) error {
	// Parse the private keys
	signer, err := ssh.ParsePrivateKey([]byte(s.sshKey))
	if err != nil {
		return err
	}

	// SSH client configuration
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	// Connect to the SSH server
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", Instance.PublicIP, Instance.SSHPort), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create a new SSH session
	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Create a temporary bash script file
	script := "/tmp/postscript.sh"
	err = session.Run(fmt.Sprintf("echo '%s' > %s && chmod +x %s && %s", string(userData), script, script, script))
	if err != nil {
		return err
	}
	return nil
}

// Format AuthorizedKeys to insert in http request body
func formatPubKeys(pubkeys []string) []string {
	formattedKeys := make([]string, len(pubkeys))
	for i, key := range pubkeys {
		formattedKeys[i] = fmt.Sprintf(`"%s"`, key)
	}
	return formattedKeys
}
