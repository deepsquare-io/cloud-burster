package exoscale

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	egoscalev2 "github.com/exoscale/egoscale/v2"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/utils/ptr"
	"github.com/squarefactory/cloud-burster/utils/try"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type DataSource struct {
	client *egoscalev2.Client
	zone   string
}

func New(
	apiKey string,
	apiSecret string,
	zoneName string,
) *DataSource {
	client, err := egoscalev2.NewClient(
		apiKey,
		apiSecret,
	)
	if err != nil {
		panic(err)
	}
	return &DataSource{
		client: client,
		zone:   zoneName,
	}
}

// FindImageID retrieves the image UUID from name
func (s *DataSource) FindImageID(ctx context.Context, name string) (string, error) {
	logger.I.Debug("FindImageID called", zap.String("name", name))

	for _, filter := range []string{"private", "public"} {
		templates, err := s.client.ListTemplates(
			ctx,
			s.zone,
			egoscalev2.ListTemplatesWithVisibility(filter),
		)
		if err != nil {
			logger.I.Error("ListTemplates failed", zap.String("filter", filter))
		}
		for _, template := range templates {
			if *template.Name == name {
				logger.I.Debug("FindImageID returned", zap.String("image", *template.ID))
				return *template.ID, nil
			}
		}
	}

	return "", errors.New("didn't find an image")
}

// FindFlavorID retrieves the flavor UUID from name
func (s *DataSource) FindFlavorID(ctx context.Context, name string) (string, error) {
	logger.I.Debug("FindFlavorID called", zap.String("name", name))
	types, err := s.client.ListInstanceTypes(ctx, s.zone)
	if err != nil {
		return "", err
	}
	family, size, _ := strings.Cut(strings.ToLower(name), "-")
	for _, so := range types {
		// For standard, compare size with the whole name
		if *so.Family == "standard" && *so.Size == strings.ToLower(name) {
			logger.I.Debug("FindFlavorID returned", zap.String("flavor", *so.ID))
			return *so.ID, nil
		}
		if *so.Family == family && *so.Size == size {
			logger.I.Debug("FindFlavorID returned", zap.String("flavor", *so.ID))
			return *so.ID, nil
		}
	}
	return "", errors.New("didn't find a flavor")
}

// FindNetworkID retrieves the network UUID from name
func (s *DataSource) FindNetworkID(ctx context.Context, name string) (string, error) {
	logger.I.Debug("FindNetworkID called", zap.String("name", name))
	networks, err := s.client.ListPrivateNetworks(ctx, s.zone)
	if err != nil {
		return "", err
	}
	for _, so := range networks {
		if *so.Name == name {
			logger.I.Debug("FindNetworkID returned", zap.String("network", *so.ID))
			return *so.ID, nil
		}
	}
	return "", errors.New("didn't find a network")
}

// Create an instance
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
	imageID, err := s.FindImageID(ctx, host.ImageName)
	if err != nil {
		return err
	}
	flavorID, err := s.FindFlavorID(ctx, host.FlavorName)
	if err != nil {
		return err
	}
	networkID, err := s.FindNetworkID(ctx, cloud.Network.Name)
	if err != nil {
		return err
	}
	_, net, err := net.ParseCIDR(cloud.Network.SubnetCIDR)
	if err != nil {
		return err
	}
	mask, _ := net.Mask.Size()

	var customConfig []byte
	if len(cloud.CustomConfig) == 0 {
		customConfig = []byte{}
	} else {
		customConfig, err = yaml.Marshal(cloud.CustomConfig)
		if err != nil {
			return err
		}
	}

	userData, err := GenerateCloudConfig(&CloudConfigOpts{
		AuthorizedKeys:    cloud.AuthorizedKeys,
		PostScripts:       cloud.PostScripts,
		DNS:               cloud.Network.DNS,
		Search:            cloud.Network.Search,
		AddressCIDR:       fmt.Sprintf("%s/%d", host.IP, mask),
		Gateway:           cloud.Network.Gateway,
		CustomCloudConfig: string(customConfig),
	})
	if err != nil {
		return err
	}
	userDataB64 := base64.StdEncoding.EncodeToString(userData)
	instance, err := s.client.CreateInstance(ctx, s.zone, &egoscalev2.Instance{
		Name:           &host.Name,
		TemplateID:     &imageID,
		InstanceTypeID: &flavorID,
		DiskSize:       ptr.Ref(int64(host.DiskSize)),
		Zone:           &s.zone,
		UserData:       &userDataB64,
		PrivateNetworkIDs: &[]string{
			networkID,
		},
	})
	if err != nil {
		return err
	}
	logger.I.Info("spawned a server", zap.Any("server", instance))
	return nil
}

func (s *DataSource) FindServer(
	ctx context.Context,
	name string,
) (*egoscalev2.Instance, error) {
	logger.I.Debug("FindServer called", zap.String("name", name))
	instances, err := s.client.ListInstances(ctx, s.zone)
	if err != nil {
		return nil, err
	}
	for _, vm := range instances {
		if *vm.Name == name {
			logger.I.Debug("FindServer returned", zap.Any("server", vm))
			return vm, nil
		}
	}
	return nil, errors.New("didn't find a server")
}

func (s *DataSource) Delete(
	ctx context.Context,
	name string,
) error {
	logger.I.Warn("Delete called", zap.String("name", name))
	server, err := try.Do(func() (*egoscalev2.Instance, error) {
		vm, err := s.FindServer(ctx, name)
		if err != nil {
			return vm, err
		}
		if *vm.State != "running" &&
			*vm.State != "stopped" {
			logger.I.Debug("the server isn't stable yet", zap.Any("server", vm))
			return vm, errors.New("state isn't stable yet")
		}
		if *vm.State == "destroyed" {
			logger.I.Warn("Somehow the server was already deleted", zap.Any("server", vm))
			return vm, nil
		}

		if err := s.client.DeleteInstance(ctx, s.zone, vm); err != nil {
			return vm, err
		}

		return vm, nil
	}, 10, 5*time.Second)
	if err != nil {
		return err
	}

	logger.I.Warn("deleted a server", zap.Any("server", server))
	return nil
}
