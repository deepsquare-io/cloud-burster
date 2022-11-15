//go:build unit

package config_test

import (
	"path/filepath"
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/utils/projectpath"
	"github.com/stretchr/testify/suite"
)

var cleanConfig = config.Config{
	APIVersion: config.APIVersion,
	Clouds: []config.Cloud{
		cleanOpenstackCloud,
		cleanExoscaleCloud,
	},
}

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestValidate() {
	tests := []struct {
		input         *config.Config
		isError       bool
		errorContains []string
		title         string
	}{
		{
			input: &cleanConfig,
			title: "Positive test",
		},
		{
			isError: true,
			errorContains: []string{
				"APIVersion",
				"equalAPI",
			},
			input: &config.Config{
				APIVersion: "aaa",
			},
			title: "Wrong API version",
		},
		{
			isError: true,
			errorContains: []string{
				"Cloud",
			},
			input: &config.Config{
				APIVersion: config.APIVersion,
				Clouds: []config.Cloud{
					{},
				},
			},
			title: "Valid cloud",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Act
			err := tt.input.Validate()

			// Assert
			if tt.isError {
				suite.Error(err)
				for _, contain := range tt.errorContains {
					suite.ErrorContains(err, contain)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *ConfigTestSuite) TestParseFile() {
	// Arrange
	expected := &cleanConfig

	// Act
	config, err := config.ParseFile(filepath.Join(projectpath.Root, "config.yaml.example"))
	suite.NoError(err)
	err = config.Validate()

	// Assert
	suite.NoError(err)
	suite.Equal(expected, config)
}

func (suite *ConfigTestSuite) TestSearchHostByHostName() {
	// Arrange
	conf := cleanConfig
	err := conf.Validate()
	suite.NoError(err)

	tests := []struct {
		input         string
		expectedHost  *config.Host
		expectedCloud *config.Cloud
		title         string
	}{
		{
			input:         "cn-s-5.example.com",
			expectedCloud: &cleanConfig.Clouds[0],
			expectedHost: &config.Host{
				Name:       "cn-s-5.example.com",
				DiskSize:   cleanConfig.Clouds[0].GroupsHost[0].HostTemplate.DiskSize,
				FlavorName: cleanConfig.Clouds[0].GroupsHost[0].HostTemplate.FlavorName,
				ImageName:  cleanConfig.Clouds[0].GroupsHost[0].HostTemplate.ImageName,
				IP:         "172.28.1.5",
			},
			title: "Search in groups host",
		},
		{
			input:         cleanConfig.Clouds[0].Hosts[0].Name,
			expectedCloud: &cleanConfig.Clouds[0],
			expectedHost:  &cleanConfig.Clouds[0].Hosts[0],
			title:         "Search in hosts",
		},
		{
			input:         "aaa",
			expectedCloud: nil,
			expectedHost:  nil,
			title:         "Not found",
		},
	}

	// Act
	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Act
			host, cloud, err := conf.SearchHostByHostName(tt.input)

			// Assert
			suite.NoError(err)
			suite.Equal(tt.expectedHost, host)
			suite.Equal(tt.expectedCloud, cloud)
		})
	}
}

func (suite *ConfigTestSuite) TestGenerateHosts() {
	// Arrange
	conf := &config.Config{
		APIVersion: config.APIVersion,
		Clouds: []config.Cloud{
			{
				AuthorizedKeys: cleanConfig.Clouds[0].AuthorizedKeys,
				PostScripts:    cleanConfig.Clouds[0].PostScripts,
				Network:        cleanConfig.Clouds[0].Network,
				Openstack:      cleanConfig.Clouds[0].Openstack,
				Exoscale:       cleanConfig.Clouds[0].Exoscale,
				Type:           cleanConfig.Clouds[0].Type,
				GroupsHost: []config.GroupHost{
					{
						NamePattern: "cn-[1-5]",
						IPCidr:      "172.28.0.0/20",
						IPOffset:    256,
						HostTemplate: config.Host{
							DiskSize:   50,
							FlavorName: "flavor",
							ImageName:  "image",
						},
					},
				},
				Hosts: []config.Host{
					{
						Name:       "cn-test",
						DiskSize:   50,
						FlavorName: "flavor",
						ImageName:  "image",
						IP:         "10.10.10.10",
					},
				},
			},
		},
	}
	err := conf.Validate()
	expected := `10.10.10.10 cn-test
172.28.1.1 cn-1
172.28.1.2 cn-2
172.28.1.3 cn-3
172.28.1.4 cn-4
172.28.1.5 cn-5
`
	suite.NoError(err)

	// Act
	actual, err := conf.GenerateHosts()

	// Assert
	suite.NoError(err)
	suite.Equal(expected, actual)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, &ConfigTestSuite{})
}
