//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var cleanOpenstackCloud = config.Cloud{
	Network: cleanNetwork,
	Hosts: []config.Host{
		cleanHost,
	},
	GroupsHost: []config.GroupHost{
		cleanGroupHost,
	},
	Type:                    "openstack",
	CloudConfigTemplateOpts: cleanCloudConfigTemplateOpts,
	Openstack:               &cleanOpenstack,
}

var cleanExoscaleCloud = config.Cloud{
	Network: cleanNetwork,
	Hosts: []config.Host{
		cleanHost,
	},
	GroupsHost: []config.GroupHost{
		cleanGroupHost,
	},
	Type:                    "exoscale",
	CloudConfigTemplateOpts: cleanCloudConfigTemplateOpts,
	Exoscale:                &cleanExoscale,
}

type CloudTestSuite struct {
	suite.Suite
}

func (suite *CloudTestSuite) TestValidate() {
	tests := []struct {
		input   *config.Cloud
		isError bool
		title   string
	}{
		{
			isError: false,
			input:   &cleanOpenstackCloud,
		},
		{
			isError: true,
			input: &config.Cloud{
				GroupsHost:              cleanOpenstackCloud.GroupsHost,
				CloudConfigTemplateOpts: cleanOpenstackCloud.CloudConfigTemplateOpts,
				Openstack:               cleanOpenstackCloud.Openstack,
				Type:                    cleanOpenstackCloud.Type,
			},
			title: "Network required/valid",
		},
		{
			isError: true,
			input: &config.Cloud{
				Network: cleanOpenstackCloud.Network,
				GroupsHost: []config.GroupHost{
					{},
				},
				CloudConfigTemplateOpts: cleanOpenstackCloud.CloudConfigTemplateOpts,
				Openstack:               cleanOpenstackCloud.Openstack,
				Type:                    cleanOpenstackCloud.Type,
			},
			title: "GroupsHost valid",
		},
		{
			isError: true,
			input: &config.Cloud{
				Network:                 cleanOpenstackCloud.Network,
				GroupsHost:              cleanOpenstackCloud.GroupsHost,
				CloudConfigTemplateOpts: config.CloudConfigTemplateOpts{},
				Openstack:               cleanOpenstackCloud.Openstack,
				Type:                    cleanOpenstackCloud.Type,
			},
			title: "cloudConfig valid",
		},
		{
			isError: true,
			input: &config.Cloud{
				Network:                 cleanOpenstackCloud.Network,
				GroupsHost:              cleanOpenstackCloud.GroupsHost,
				CloudConfigTemplateOpts: cleanOpenstackCloud.CloudConfigTemplateOpts,
				Type:                    cleanOpenstackCloud.Type,
				Openstack: &config.Openstack{
					IdentityEndpoint: "aaa",
				},
			},
			title: "openstack valid",
		},
		{
			isError: true,
			input: &config.Cloud{
				Network:                 cleanOpenstackCloud.Network,
				GroupsHost:              cleanOpenstackCloud.GroupsHost,
				CloudConfigTemplateOpts: cleanOpenstackCloud.CloudConfigTemplateOpts,
				Type:                    "openstack",
			},
			title: "If type == openstack, openstack is required",
		},
		{
			isError: true,
			input: &config.Cloud{
				Network:                 cleanOpenstackCloud.Network,
				GroupsHost:              cleanOpenstackCloud.GroupsHost,
				CloudConfigTemplateOpts: cleanOpenstackCloud.CloudConfigTemplateOpts,
				Type:                    "exoscale",
			},
			title: "If type == exoscale, exoscale is required",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Act
			err := tt.input.Validate()

			// Assert
			if tt.isError {
				logger.I.Info("expected error", zap.Error(err))
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func TestCloudTestSuite(t *testing.T) {
	suite.Run(t, &CloudTestSuite{})
}
