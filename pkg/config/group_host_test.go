//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var cleanGroupHost = config.GroupHost{
	NamePattern: "cn-s-[1-50].example.com",
	IPCidr:      "172.28.0.0/20",
	IPOffset:    256,
	HostTemplate: config.Host{
		DiskSize:   50,
		FlavorName: "d2-2",
		ImageName:  "Rocky Linux 9",
	},
}

type GroupHostTestSuite struct {
	suite.Suite
}

func (suite *GroupHostTestSuite) TestValidate() {
	tests := []struct {
		input   *config.GroupHost
		isError bool
		title   string
	}{
		{
			input: &cleanGroupHost,
			title: "Positive test",
		},
		{
			isError: true,
			input: &config.GroupHost{
				NamePattern:  "",
				IPCidr:       cleanGroupHost.IPCidr,
				IPOffset:     cleanGroupHost.IPOffset,
				HostTemplate: cleanGroupHost.HostTemplate,
			},
			title: "namePattern required",
		},
		{
			isError: true,
			input: &config.GroupHost{
				NamePattern:  cleanGroupHost.NamePattern,
				IPCidr:       "",
				IPOffset:     cleanGroupHost.IPOffset,
				HostTemplate: cleanGroupHost.HostTemplate,
			},
			title: "ipCIDR required",
		},
		{
			isError: true,
			input: &config.GroupHost{
				NamePattern:  cleanGroupHost.NamePattern,
				IPCidr:       "172.24.0.0",
				IPOffset:     cleanGroupHost.IPOffset,
				HostTemplate: cleanGroupHost.HostTemplate,
			},
			title: "ipCIDR valid",
		},
		{
			isError: true,
			input: &config.GroupHost{
				NamePattern:  cleanGroupHost.NamePattern,
				IPCidr:       cleanGroupHost.IPCidr,
				IPOffset:     cleanGroupHost.IPOffset,
				HostTemplate: config.Host{},
			},
			title: "hostTemplate valid",
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

func (suite *GroupHostTestSuite) TestGenerateHosts() {
	hostTemplate := config.Host{
		DiskSize:   50,
		FlavorName: "flavor",
		ImageName:  "image",
	}
	tests := []struct {
		input    config.GroupHost
		expected []config.Host
		isError  bool
		title    string
	}{
		{
			input: config.GroupHost{
				NamePattern:  "cn[1-5]",
				IPCidr:       "172.20.0.0/20",
				IPOffset:     5,
				HostTemplate: hostTemplate,
			},
			isError: false,
			expected: []config.Host{
				{
					Name:       "cn1",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.6",
				},
				{
					Name:       "cn2",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.7",
				},
				{
					Name:       "cn3",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.8",
				},
				{
					Name:       "cn4",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.9",
				},
				{
					Name:       "cn5",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.10",
				},
			},
			title: "Positive test",
		},
		{
			input: config.GroupHost{
				NamePattern:  "cn[1-2000]",
				IPCidr:       "172.20.0.0/24",
				HostTemplate: hostTemplate,
			},
			isError:  true,
			expected: []config.Host{},
			title:    "Not enough IP",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Act
			actual, err := tt.input.GenerateHosts()

			// Assert
			if tt.isError {
				logger.I.Info("expected error", zap.Error(err))
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
			suite.Equal(tt.expected, actual)
		})
	}
}

func TestGroupHostTestSuite(t *testing.T) {
	suite.Run(t, &GroupHostTestSuite{})
}
