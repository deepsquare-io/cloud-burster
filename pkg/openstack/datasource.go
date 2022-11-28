package openstack

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/portsecurity"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/middlewares"
	"github.com/squarefactory/cloud-burster/utils/try"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type DataSource struct {
	provider      *gophercloud.ProviderClient
	computeClient *gophercloud.ServiceClient
	networkClient *gophercloud.ServiceClient
}

func New(
	endpoint string,
	username string,
	password string,
	tenantID string,
	tenantName string,
	region string,
	domainID string,
) *DataSource {
	provider, err := openstack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: endpoint,
		Username:         username,
		Password:         password,
		TenantID:         tenantID,
		TenantName:       tenantName,
		DomainID:         domainID,
	})
	provider.HTTPClient.Transport = &middlewares.RoundTripper{
		RoundTripper: http.DefaultTransport,
	}
	if err != nil {
		logger.I.Panic("failed to authenticate client", zap.Error(err))
	}
	computeClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		logger.I.Panic("couldn't instanciate computeClient", zap.Error(err))
	}
	networkClient, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		logger.I.Panic("couldn't instanciate networkClient", zap.Error(err))
	}
	return &DataSource{
		provider:      provider,
		computeClient: computeClient,
		networkClient: networkClient,
	}
}

// FindImageID retrieves the image UUID from name
func (s *DataSource) FindImageID(name string) (string, error) {
	logger.I.Debug("FindImageID called", zap.String("name", name))
	pager := images.ListDetail(s.computeClient, images.ListOpts{
		Name: name,
	})
	var result images.Image
	err := pager.EachPage(func(p pagination.Page) (bool, error) {
		list, err := images.ExtractImages(p)
		if err != nil {
			return false, err
		}

		for _, i := range list {
			if i.Name == name {
				result = i
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return "", err
	}
	if result.ID == "" {
		return "", errors.New("didn't find an image")
	}
	logger.I.Debug("FindImageID returned", zap.Any("image", result))
	return result.ID, nil
}

// FindFlavorID retrieves the flavor UUID from name
func (s *DataSource) FindFlavorID(name string) (string, error) {
	logger.I.Debug("FindFlavorID called", zap.String("name", name))
	pager := flavors.ListDetail(s.computeClient, flavors.ListOpts{})
	var result flavors.Flavor
	err := pager.EachPage(func(p pagination.Page) (bool, error) {
		list, err := flavors.ExtractFlavors(p)
		if err != nil {
			return false, err
		}

		for _, f := range list {
			if f.Name == name {
				result = f
				return false, nil
			}
		}

		return true, nil
	})

	if err != nil {
		return "", err
	}
	if result.ID == "" {
		return "", errors.New("didn't find a flavor")
	}
	logger.I.Debug("FindFlavorID returned", zap.Any("flavor", result))
	return result.ID, nil
}

// FindNetworkID retrieves the network from name
func (s *DataSource) FindNetworkID(name string) (string, error) {
	logger.I.Debug("FindNetworkID called", zap.String("name", name))
	pager := networks.List(s.networkClient, networks.ListOpts{
		Name: name,
	})
	var result networks.Network
	err := pager.EachPage(func(p pagination.Page) (bool, error) {
		list, err := networks.ExtractNetworks(p)
		if err != nil {
			return false, err
		}

		for _, net := range list {
			if net.Name == name {
				result = net
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return "", err
	}
	if result.ID == "" {
		return "", errors.New("didn't find a network")
	}
	logger.I.Debug("FindNetworkID returned", zap.Any("network", result))
	return result.ID, nil
}

// FindSubnetIDByNetwork retrieves the subnet UUID from CIDR by Network
func (s *DataSource) FindSubnetIDByNetwork(cidr string, networkID string) (string, error) {
	pager := subnets.List(s.networkClient, subnets.ListOpts{
		NetworkID: networkID,
		CIDR:      cidr,
	})

	var result subnets.Subnet
	err := pager.EachPage(func(p pagination.Page) (bool, error) {
		list, err := subnets.ExtractSubnets(p)
		if err != nil {
			return false, err
		}

		for _, subnet := range list {
			if subnet.CIDR == cidr {
				result = subnet
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return "", err
	}
	if result.ID == "" {
		return "", errors.New("didn't find a subnet")
	}
	logger.I.Debug("FindSubnetIDByNetwork returned", zap.Any("subnet", result))
	return result.ID, nil
}

// CreatePort connected to a network
func (s *DataSource) CreatePort(ip string, networkID string, subnetID string) (string, error) {
	adminStateUp := true
	portCreateOpts := ports.CreateOpts{
		NetworkID:      networkID,
		AdminStateUp:   &adminStateUp,
		SecurityGroups: &[]string{},
		FixedIPs: []ports.IP{
			{
				IPAddress: ip,
				SubnetID:  subnetID,
			},
		},
	}
	portSecurityEnabled := false
	createOpts := portsecurity.PortCreateOptsExt{
		CreateOptsBuilder:   portCreateOpts,
		PortSecurityEnabled: &portSecurityEnabled,
	}
	port, err := ports.Create(s.networkClient, createOpts).Extract()
	if err != nil {
		return "", err
	}
	logger.I.Debug("CreatePort returned", zap.Any("port", port))
	return port.ID, nil
}

// FindPortByDeviceID retrieves the port UUID attached to an instance
func (s *DataSource) FindPortByDeviceID(deviceID string) (string, error) {
	logger.I.Debug("FindPortByDeviceID called", zap.Any("deviceID", deviceID))
	pager := ports.List(s.networkClient, ports.ListOpts{
		DeviceID: deviceID,
	})

	var result ports.Port
	err := pager.EachPage(func(p pagination.Page) (bool, error) {
		list, err := ports.ExtractPorts(p)
		if err != nil {
			return false, err
		}

		for _, port := range list {
			if port.DeviceID == deviceID {
				result = port
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return "", err
	}
	if result.ID == "" {
		return "", errors.New("didn't find a port")
	}
	logger.I.Debug("FindPortByDeviceID returned", zap.Any("port", result))
	return result.ID, nil
}

func (s *DataSource) DeletePort(id string) error {
	logger.I.Warn("DeletePort called", zap.String("id", id))
	return ports.Delete(s.networkClient, id).ExtractErr()
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
	image, err := s.FindImageID(host.ImageName)
	if err != nil {
		return err
	}
	flavor, err := s.FindFlavorID(host.FlavorName)
	if err != nil {
		return err
	}
	networkID, err := s.FindNetworkID(cloud.Network.Name)
	if err != nil {
		return err
	}
	subnetID, err := s.FindSubnetIDByNetwork(cloud.Network.SubnetCIDR, networkID)
	if err != nil {
		return err
	}
	portID, err := s.CreatePort(host.IP, networkID, subnetID)
	if err != nil {
		return err
	}
	configDrive := true

	customConfig, err := yaml.Marshal(cloud.CustomConfig)
	if err != nil {
		return err
	}

	userData, err := GenerateCloudConfig(&CloudConfigOpts{
		AuthorizedKeys:    cloud.AuthorizedKeys,
		PostScripts:       cloud.PostScripts,
		DNS:               cloud.Network.DNS,
		Search:            cloud.Network.Search,
		CustomCloudConfig: string(customConfig),
	})
	if err != nil {
		return err
	}
	server, err := bootfromvolume.Create(s.computeClient, bootfromvolume.CreateOptsExt{
		CreateOptsBuilder: servers.CreateOpts{
			Name:      host.Name,
			ImageRef:  image,
			FlavorRef: flavor,
			UserData:  userData,
			Networks: []servers.Network{
				{
					Port: portID,
				},
			},
			ConfigDrive: &configDrive,
		},
		BlockDevice: []bootfromvolume.BlockDevice{
			{
				UUID:                image,
				SourceType:          "image",
				DestinationType:     "local",
				BootIndex:           0,
				DeleteOnTermination: true,
			},
			{
				SourceType:          "blank",
				DestinationType:     "volume",
				DeleteOnTermination: true,
				BootIndex:           1,
				VolumeSize:          host.DiskSize,
			},
		},
	}).Extract()
	if err != nil {
		if err := s.DeletePort(portID); err != nil {
			logger.I.Error("failed to delete port", zap.Error(err))
		}
		return err
	}
	logger.I.Info("spawned a server", zap.Any("server", server))
	return nil
}

// FindServerID retrieves the instance UUID by name
func (s *DataSource) FindServerID(name string) (string, error) {
	logger.I.Debug("FindServerID called", zap.String("name", name))
	pager := servers.List(s.computeClient, servers.ListOpts{
		Name: name,
	})

	var result servers.Server
	err := pager.EachPage(func(p pagination.Page) (bool, error) {
		list, err := servers.ExtractServers(p)
		if err != nil {
			return false, err
		}

		for _, server := range list {
			if server.Name == name {
				result = server
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return "", err
	}

	if result.ID == "" {
		return "", errors.New("didn't find a server")
	}
	logger.I.Debug("FindServerID returned", zap.Any("server", result))
	return result.ID, nil
}

func (s *DataSource) Delete(ctx context.Context, name string) error {
	logger.I.Warn("Delete called", zap.String("name", name))
	serverID, err := try.Do(func() (string, error) {
		return s.FindServerID(name)
	}, 3, 5*time.Second)
	if err != nil {
		return err
	}

	// Find associated port and delete it
	portID, err := try.Do(func() (string, error) {
		return s.FindPortByDeviceID(serverID)
	}, 10, 5*time.Second)
	if err != nil {
		logger.I.Warn("couldn't delete port of associated server",
			zap.Any("serverID", serverID),
			zap.Error(err),
		)
	} else {
		err = s.DeletePort(portID)
		if err != nil {
			return err
		}
	}

	// Finally, delete the server
	err = servers.ForceDelete(s.computeClient, serverID).ExtractErr()
	if err != nil {
		return err
	}

	logger.I.Warn("deleted a server", zap.Any("server", serverID))
	return nil
}
