package cloud

import (
	"context"
	"errors"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/exoscale"
	"github.com/squarefactory/cloud-burster/pkg/openstack"
	"github.com/squarefactory/cloud-burster/pkg/shadow"
	"go.uber.org/zap"
)

type DataSource interface {
	Create(
		ctx context.Context,
		host *config.Host,
		cloud *config.Cloud,
	) error
	Delete(
		ctx context.Context,
		name string,
	) error
}

func New(conf *config.Cloud) (DataSource, error) {
	switch conf.Type {
	case "openstack":
		return openstack.New(
			conf.Openstack.IdentityEndpoint,
			conf.Openstack.UserName,
			conf.Openstack.Password,
			conf.Openstack.TenantID,
			conf.Openstack.TenantName,
			conf.Openstack.Region,
			conf.Openstack.DomainID,
		), nil
	case "exoscale":
		return exoscale.New(
			conf.Exoscale.ComputeEndpoint,
			conf.Exoscale.APIKey,
			conf.Exoscale.APISecret,
			conf.Exoscale.Zone,
		), nil
	case "shadow":
		return shadow.New(
			conf.Shadow.Username,
			conf.Shadow.Password,
			conf.Shadow.Zone,
			conf.Shadow.SSHKey,
		), nil
	}

	logger.I.Error(
		"no cloud associated with the configuration",
		zap.Any("cloud configuration", conf),
	)

	return nil, errors.New("no cloud associated with the configuration")
}
