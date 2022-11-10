package cloud

import (
	"errors"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/openstack"
	"go.uber.org/zap"
)

type DataSource interface {
	Create(
		host *config.Host,
		network *config.Network,
		cloudConfigOpts *config.CloudConfigTemplateOpts,
	) error
	Delete(name string) error
}

func Create(conf *config.Cloud) (DataSource, error) {
	if conf.Openstack.Enabled {
		return openstack.New(
			conf.Openstack.IdentityEndpoint,
			conf.Openstack.UserName,
			conf.Openstack.Password,
			conf.Openstack.TenantID,
			conf.Openstack.TenantName,
			conf.Openstack.Region,
			conf.Openstack.DomainID,
		), nil
	}

	logger.I.Error(
		"no cloud associated with the configuration",
		zap.Any("cloud configuration", conf),
	)

	return nil, errors.New("no cloud associated with the configuration")
}
