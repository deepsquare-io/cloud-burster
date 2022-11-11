package exoscale

import (
	"errors"
	"net/http"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/middlewares"
	"github.com/squarefactory/cloud-burster/utils/try"
	"go.uber.org/zap"
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
		egoscale.WithoutV2Client(),
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

func (s *DataSource) Create(
	host *config.Host,
	network *config.Network,
	cloudConfigOpts *config.CloudConfigTemplateOpts,
) error {
	// TODO: implements
	return nil
}

func (s *DataSource) FindServer(name string) (egoscale.VirtualMachine, error) {
	logger.I.Debug("FindServer called", zap.String("name", name))
	req := &egoscale.ListVirtualMachines{
		Name:   name,
		ZoneID: s.zone.ID,
	}
	resp, err := s.client.Request(req)
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
	return egoscale.VirtualMachine{}, errors.New("server not found")
}

func (s *DataSource) Delete(name string) error {
	logger.I.Warn("Delete called", zap.String("name", name))
	server, err := try.Do(func() (egoscale.VirtualMachine, error) {
		return s.FindServer(name)
	}, 3, 5*time.Second)
	if err != nil {
		return err
	}

	err = s.client.Delete(server)
	if err != nil {
		return err
	}

	logger.I.Warn("deleted a server", zap.Any("server", server))
	return nil
}
