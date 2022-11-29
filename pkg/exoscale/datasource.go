package exoscale

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/middlewares"
	"github.com/squarefactory/cloud-burster/utils/try"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type DataSource struct {
	client *egoscale.Client
	zone   egoscale.Zone
}

func New(
	endpoint string,
	apiKey string,
	apiSecret string,
	zoneName string,
) *DataSource {
	client := egoscale.NewClient(
		endpoint,
		apiKey,
		apiSecret,
	)
	client.HTTPClient.Transport = &middlewares.RoundTripper{
		RoundTripper: http.DefaultTransport,
	}
	zone := func() egoscale.Zone {
		req := &egoscale.ListZones{}
		resp, err := client.Request(req)
		if err != nil {
			return egoscale.Zone{}
		}
		zones := resp.(*egoscale.ListZonesResponse)
		for _, zone := range zones.Zone {
			if zone.Name == zoneName {
				return zone
			}
		}
		logger.I.Fatal("zone not found", zap.String("zone", zoneName))
		return egoscale.Zone{}
	}()
	return &DataSource{
		client: client,
		zone:   zone,
	}
}

// FindImageID retrieves the image UUID from name
func (s *DataSource) FindImageID(ctx context.Context, name string) (string, error) {
	logger.I.Debug("FindImageID called", zap.String("name", name))

	for _, filter := range []string{"self", "featured", "community"} {
		req := &egoscale.ListTemplates{
			Name:           name,
			TemplateFilter: filter,
			ZoneID:         s.zone.ID,
		}
		resp, err := s.client.RequestWithContext(ctx, req)
		if err != nil {
			logger.I.Warn(
				"FindImageID failed with template filter",
				zap.String("name", name),
				zap.String("filter", filter),
			)
			continue
		}
		templates := resp.(*egoscale.ListTemplatesResponse)
		for _, template := range templates.Template {
			if template.Name == name {
				logger.I.Debug("FindImageID returned", zap.String("image", template.ID.String()))
				return template.ID.String(), nil
			}
		}
	}

	return "", errors.New("didn't find an image")
}

// FindFlavorID retrieves the flavor UUID from name
func (s *DataSource) FindFlavorID(ctx context.Context, name string) (string, error) {
	logger.I.Debug("FindFlavorID called", zap.String("name", name))
	req := &egoscale.ListServiceOfferings{
		Name: name,
	}
	resp, err := s.client.RequestWithContext(ctx, req)
	if err != nil {
		return "", err
	}
	serviceOfferings := resp.(*egoscale.ListServiceOfferingsResponse)
	for _, so := range serviceOfferings.ServiceOffering {
		if so.Name == name {
			logger.I.Debug("FindFlavorID returned", zap.String("flavor", so.ID.String()))
			return so.ID.String(), nil
		}
	}
	return "", errors.New("didn't find a flavor")
}

// FindNetworkID retrieves the network UUID from name
func (s *DataSource) FindNetworkID(ctx context.Context, name string) (string, error) {
	logger.I.Debug("FindNetworkID called", zap.String("name", name))
	req := &egoscale.ListNetworks{
		ZoneID: s.zone.ID,
	}
	resp, err := s.client.RequestWithContext(ctx, req)
	if err != nil {
		return "", err
	}
	networks := resp.(*egoscale.ListNetworksResponse)
	for _, so := range networks.Network {
		if so.Name == name {
			logger.I.Debug("FindNetworkID returned", zap.String("network", so.ID.String()))
			return so.ID.String(), nil
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
	startVM := true
	userDataB64 := base64.StdEncoding.EncodeToString(userData)
	req := &egoscale.DeployVirtualMachine{
		Name:              host.Name,
		TemplateID:        egoscale.MustParseUUID(imageID),
		ServiceOfferingID: egoscale.MustParseUUID(flavorID),
		RootDiskSize:      int64(host.DiskSize),
		ZoneID:            s.zone.ID,
		UserData:          userDataB64,
		StartVM:           &startVM,
		NetworkIDs: []egoscale.UUID{
			*egoscale.MustParseUUID(networkID),
		},
	}
	resp, err := s.client.RequestWithContext(ctx, req)
	if err != nil {
		return err
	}
	vm := resp.(*egoscale.VirtualMachine)
	logger.I.Info("spawned a server", zap.Any("server", vm))
	return nil
}

func (s *DataSource) FindServer(
	ctx context.Context,
	name string,
) (egoscale.VirtualMachine, error) {
	logger.I.Debug("FindServer called", zap.String("name", name))
	req := &egoscale.ListVirtualMachines{
		Name:   name,
		ZoneID: s.zone.ID,
	}
	resp, err := s.client.RequestWithContext(ctx, req)
	if err != nil {
		return egoscale.VirtualMachine{}, err
	}
	vms := resp.(*egoscale.ListVirtualMachinesResponse)
	for _, vm := range vms.VirtualMachine {
		if vm.Name == name {
			logger.I.Debug("FindServer returned", zap.Any("server", vm))
			return vm, nil
		}
	}
	return egoscale.VirtualMachine{}, errors.New("didn't find a server")
}

func (s *DataSource) Delete(
	ctx context.Context,
	name string,
) error {
	logger.I.Warn("Delete called", zap.String("name", name))
	server, err := try.Do(func() (egoscale.VirtualMachine, error) {
		vm, err := s.FindServer(ctx, name)
		if err != nil {
			return vm, err
		}

		state := egoscale.VirtualMachineState(vm.State)
		if state != egoscale.VirtualMachineRunning &&
			state != egoscale.VirtualMachineShutdowned &&
			state != egoscale.VirtualMachineStopped {
			logger.I.Debug("the server isn't stable yet", zap.Any("server", vm))
			return vm, errors.New("state isn't stable yet")
		}
		if state == egoscale.VirtualMachineDestroyed {
			logger.I.Warn("Somehow the server was already deleted", zap.Any("server", vm))
			return vm, nil
		}

		if err := s.client.DeleteWithContext(ctx, vm); err != nil {
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
