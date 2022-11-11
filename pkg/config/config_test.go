//go:build unit

package config_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/utils/projectpath"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var cleanConfig = &config.Config{
	APIVersion: config.APIVersion,
	Clouds: []config.Cloud{
		{
			Network: config.Network{
				Name:       "name",
				SubnetCIDR: "172.28.0.0/20",
			},

			Hosts: []config.Host{
				{
					Name:       "test.example.com",
					DiskSize:   100,
					FlavorName: "test-flavor",
					ImageName:  "image",
					IP:         "172.28.16.254",
				},
			},
			GroupsHost: []config.GroupHost{
				{
					NamePattern: "cn-s-[1-50].example.com",
					IPCidr:      "172.28.0.0/20",
					HostTemplate: config.Host{
						DiskSize:   50,
						FlavorName: "d2-2",
						ImageName:  "Rocky Linux 9",
					},
				},
			},
			CloudConfigTemplateOpts: config.CloudConfigTemplateOpts{
				AuthorizedKeys: []string{
					"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4",
				},
				DNS:    "1.1.1.1",
				Search: "example.com",
				PostScripts: config.PostScriptsOpts{
					Git: config.GitOpts{
						Key: "key",
						URL: "git@github.com:SquareFactory/compute-configs.git",
						Ref: "main",
					},
				},
			},
			Openstack: config.Openstack{
				Enabled:          true,
				IdentityEndpoint: "https://auth.cloud.ovh.net/",
				UserName:         "user",
				Password:         "",
				TenantID:         "tenantID",
				TenantName:       "tenantName",
				DomainID:         "default",
				Region:           "GRA9",
			},
		},
	},
}

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestValidate() {
	tests := []struct {
		input    string
		expected *config.Config
		isError  bool
		title    string
	}{
		{
			input: fmt.Sprintf(`apiVersion: '%s'
clouds:
  - network:
      name: 'name'
      subnetCIDR: '172.28.0.0/20'
    hosts:
      - name: test.example.com
        diskSize: 100
        imageName: image
        flavorName: test-flavor
        ip: "172.28.16.254"
    groupsHost:
      - namePattern: cn-s-[1-50].example.com
        ipCIDR: 172.28.0.0/20
        template:
          diskSize: 50
          flavorName: 'd2-2'
          imageName: 'Rocky Linux 9'
    cloudConfig:
      authorizedKeys:
        - 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4'
      dns: 1.1.1.1
      search: example.com
      postScripts:
        git:
          key: key
          url: git@github.com:SquareFactory/compute-configs.git
          ref: main
    openstack:
      enabled: true
      identityEndpoint: https://auth.cloud.ovh.net/
      username: user
      password: ''
      region: GRA9
      tenantID: tenantID
      tenantName: 'tenantName'
      domainID: default`, config.APIVersion),
			isError:  false,
			expected: cleanConfig,
			title:    "Positive test",
		},
		{
			input: `apiVersion: 'aaa'
clouds:
  - network:
      name: 'name'
      subnetCIDR: '172.28.0.0/20'
    groupsHost:
      - namePattern: cn-s-[1-50].example.com
        ipCIDR: 172.28.0.0/20
        template:
          diskSize: 50
          flavorName: 'd2-2'
          imageName: 'Rocky Linux 9'
    cloudConfig:
      authorizedKeys:
        - 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4'
      dns: 1.1.1.1
      search: example.com
      postScripts:
        git:
          key: key
          url: git@github.com:SquareFactory/compute-configs.git
          ref: main
    openstack:
      enabled: true
      identityEndpoint: https://auth.cloud.ovh.net/
      username: user
      password: ''
      region: GRA9
      tenantID: tenantID
      tenantName: 'tenantName'
      domainID: default`,
			isError: true,
			expected: &config.Config{
				APIVersion: "aaa",
				Clouds: []config.Cloud{
					{
						Network: config.Network{
							Name:       "name",
							SubnetCIDR: "172.28.0.0/20",
						},
						GroupsHost: []config.GroupHost{
							{
								NamePattern: "cn-s-[1-50].example.com",
								IPCidr:      "172.28.0.0/20",
								HostTemplate: config.Host{
									DiskSize:   50,
									FlavorName: "d2-2",
									ImageName:  "Rocky Linux 9",
								},
							},
						},
						CloudConfigTemplateOpts: config.CloudConfigTemplateOpts{
							AuthorizedKeys: []string{
								"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4",
							},
							DNS:    "1.1.1.1",
							Search: "example.com",
							PostScripts: config.PostScriptsOpts{
								Git: config.GitOpts{
									Key: "key",
									URL: "git@github.com:SquareFactory/compute-configs.git",
									Ref: "main",
								},
							},
						},
						Openstack: config.Openstack{
							Enabled:          true,
							IdentityEndpoint: "https://auth.cloud.ovh.net/",
							UserName:         "user",
							Password:         "",
							TenantID:         "tenantID",
							TenantName:       "tenantName",
							DomainID:         "default",
							Region:           "GRA9",
						},
					},
				},
			},
			title: "Wrong API version",
		},
		{
			input: fmt.Sprintf(`apiVersion: '%s'
clouds:
  - network:
      name: ''
      subnetCIDR: '172.28.0.0/20'
    groupsHost:
      - namePattern: cn-s-[1-50].example.com
        ipCIDR: 172.28.0.0/20
        template:
          diskSize: 50
          flavorName: 'd2-2'
          imageName: 'Rocky Linux 9'
    cloudConfig:
      authorizedKeys:
        - 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4'
      dns: 1.1.1.1
      search: example.com
      postScripts:
        git:
          key: key
          url: git@github.com:SquareFactory/compute-configs.git
          ref: main
    openstack:
      enabled: true
      identityEndpoint: https://auth.cloud.ovh.net/
      username: user
      password: ''
      region: GRA9
      tenantID: tenantID
      tenantName: 'tenantName'
      domainID: default`, config.APIVersion),
			isError: true,
			expected: &config.Config{
				APIVersion: config.APIVersion,
				Clouds: []config.Cloud{
					{
						Network: config.Network{
							Name:       "",
							SubnetCIDR: "172.28.0.0/20",
						},
						GroupsHost: []config.GroupHost{
							{
								NamePattern: "cn-s-[1-50].example.com",
								IPCidr:      "172.28.0.0/20",
								HostTemplate: config.Host{
									DiskSize:   50,
									FlavorName: "d2-2",
									ImageName:  "Rocky Linux 9",
								},
							},
						},
						CloudConfigTemplateOpts: config.CloudConfigTemplateOpts{
							AuthorizedKeys: []string{
								"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4",
							},
							DNS:    "1.1.1.1",
							Search: "example.com",
							PostScripts: config.PostScriptsOpts{
								Git: config.GitOpts{
									Key: "key",
									URL: "git@github.com:SquareFactory/compute-configs.git",
									Ref: "main",
								},
							},
						},
						Openstack: config.Openstack{
							Enabled:          true,
							IdentityEndpoint: "https://auth.cloud.ovh.net/",
							UserName:         "user",
							Password:         "",
							TenantID:         "tenantID",
							TenantName:       "tenantName",
							DomainID:         "default",
							Region:           "GRA9",
						},
					},
				},
			},
			title: "Valid cloud",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Arrange
			config := &config.Config{}
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

func (suite *ConfigTestSuite) TestParseFile() {
	// Arrange
	expected := &config.Config{
		APIVersion: config.APIVersion,
		Clouds: []config.Cloud{
			{
				Network: config.Network{
					Name:       "name",
					SubnetCIDR: "172.28.0.0/20",
				},
				GroupsHost: []config.GroupHost{
					{
						NamePattern: "cn-s-[1-50].example.com",
						IPCidr:      "172.28.0.0/20",
						HostTemplate: config.Host{
							DiskSize:   50,
							FlavorName: "d2-2",
							ImageName:  "Rocky Linux 9",
						},
					},
				},
				CloudConfigTemplateOpts: config.CloudConfigTemplateOpts{
					AuthorizedKeys: []string{
						"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4",
					},
					DNS:    "1.1.1.1",
					Search: "example.com",
					PostScripts: config.PostScriptsOpts{
						Git: config.GitOpts{
							Key: "key",
							URL: "git@github.com:SquareFactory/compute-configs.git",
							Ref: "main",
						},
					},
				},
				Openstack: config.Openstack{
					Enabled:          true,
					IdentityEndpoint: "https://auth.cloud.ovh.net/",
					UserName:         "user",
					Password:         "",
					TenantID:         "tenantID",
					TenantName:       "tenantName",
					DomainID:         "default",
					Region:           "GRA9",
				},
			},
		},
	}

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
		isError       bool
		title         string
	}{
		{
			input:         "cn-s-5.example.com",
			isError:       false,
			expectedCloud: &cleanConfig.Clouds[0],
			expectedHost: &config.Host{
				Name:       "cn-s-5.example.com",
				DiskSize:   cleanConfig.Clouds[0].GroupsHost[0].HostTemplate.DiskSize,
				FlavorName: cleanConfig.Clouds[0].GroupsHost[0].HostTemplate.FlavorName,
				ImageName:  cleanConfig.Clouds[0].GroupsHost[0].HostTemplate.ImageName,
				IP:         "172.28.0.5",
			},
			title: "Search in groups host",
		},
		{
			input:         cleanConfig.Clouds[0].Hosts[0].Name,
			isError:       false,
			expectedCloud: &cleanConfig.Clouds[0],
			expectedHost:  &cleanConfig.Clouds[0].Hosts[0],
			title:         "Search in hosts",
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
				Network:                 cleanConfig.Clouds[0].Network,
				CloudConfigTemplateOpts: cleanConfig.Clouds[0].CloudConfigTemplateOpts,
				Openstack:               cleanConfig.Clouds[0].Openstack,
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
