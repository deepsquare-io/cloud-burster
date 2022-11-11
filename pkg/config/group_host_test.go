//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type GroupHostTestSuite struct {
	suite.Suite
}

func (suite *GroupHostTestSuite) TestValidate() {
	tests := []struct {
		input    string
		expected *config.GroupHost
		isError  bool
		title    string
	}{
		{
			input: `namePattern: 'cn-[1-10]'
ipCIDR: 172.24.0.0/20
ipOffset: 5
template:
  imageName: image
  diskSize: 50
  flavorName: flavor`,
			expected: &config.GroupHost{
				NamePattern: "cn-[1-10]",
				IPCidr:      "172.24.0.0/20",
				IPOffset:    5,
				HostTemplate: config.Host{
					DiskSize:   50,
					FlavorName: "flavor",
					ImageName:  "image",
				},
			},
			title: "Positive test",
		},
		{
			input: `namePattern: ''
ipCIDR: 172.24.0.0/20
ipOffset: 5
template:
  imageName: image
  diskSize: 50
  flavorName: flavor`,
			isError: true,
			expected: &config.GroupHost{
				NamePattern: "",
				IPCidr:      "172.24.0.0/20",
				IPOffset:    5,
				HostTemplate: config.Host{
					DiskSize:   50,
					FlavorName: "flavor",
					ImageName:  "image",
				},
			},
			title: "namePattern required",
		},
		{
			input: `namePattern: 'name'
ipCIDR: ''
ipOffset: 5
template:
  imageName: image
  diskSize: 50
  flavorName: flavor`,
			isError: true,
			expected: &config.GroupHost{
				NamePattern: "name",
				IPCidr:      "",
				IPOffset:    5,
				HostTemplate: config.Host{
					DiskSize:   50,
					FlavorName: "flavor",
					ImageName:  "image",
				},
			},
			title: "ipCIDR required",
		},
		{
			input: `namePattern: 'name'
ipCIDR: '172.24.0.0'
ipOffset: 5
template:
  imageName: image
  diskSize: 50
  flavorName: flavor`,
			isError: true,
			expected: &config.GroupHost{
				NamePattern: "name",
				IPOffset:    5,
				IPCidr:      "172.24.0.0",
				HostTemplate: config.Host{
					DiskSize:   50,
					FlavorName: "flavor",
					ImageName:  "image",
				},
			},
			title: "ipCIDR valid",
		},
		{
			input: `namePattern: 'name'
ipCIDR: '172.24.0.0/24'
ipOffset: 5
template:
  imageName: ''
  diskSize: 50
  flavorName: flavor`,
			isError: true,
			expected: &config.GroupHost{
				NamePattern: "name",
				IPCidr:      "172.24.0.0/24",
				IPOffset:    5,
				HostTemplate: config.Host{
					DiskSize:   50,
					FlavorName: "flavor",
					ImageName:  "",
				},
			},
			title: "hostTemplate valid",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Arrange
			config := &config.GroupHost{}
			err := yaml.Unmarshal([]byte(tt.input), config)
			suite.NoError(err)

			// Act
			err = config.Validate()

			// Assert
			if tt.isError {
				logger.I.Debug("expected error", zap.Error(err))
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
			suite.Equal(tt.expected, config)
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
				logger.I.Debug("expected error", zap.Error(err))
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
