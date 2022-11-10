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

type CloudTestSuite struct {
	suite.Suite
}

func (suite *CloudTestSuite) TestValidate() {
	tests := []struct {
		input    string
		expected *config.Cloud
		isError  bool
		title    string
	}{
		{
			input: `network:
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
  username: user-79q6gZ9jD2Mw
  password: ''
  region: GRA9
  tenantID: 9adc45ea0a4e4d84a5acff1d829613e0
  tenantName: '6246671714361170'
  domainID: default`,
			isError: false,
			expected: &config.Cloud{
				Network: config.Network{
					Name:       "name",
					SubnetCIDR: "172.28.0.0/20",
				},
				GroupsHost: []config.GroupHost{
					{
						NamePattern: "cn-s-[1-50].example.com",
						IPcidr:      "172.28.0.0/20",
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
					UserName:         "user-79q6gZ9jD2Mw",
					Password:         "",
					TenantID:         "9adc45ea0a4e4d84a5acff1d829613e0",
					TenantName:       "6246671714361170",
					DomainID:         "default",
					Region:           "GRA9",
				},
			},
			title: "Positive test",
		},
		{
			input: `groupsHost:
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
  username: user-79q6gZ9jD2Mw
  password: ''
  region: GRA9
  tenantID: 9adc45ea0a4e4d84a5acff1d829613e0
  tenantName: '6246671714361170'
  domainID: default`,
			isError: true,
			expected: &config.Cloud{
				GroupsHost: []config.GroupHost{
					{
						NamePattern: "cn-s-[1-50].example.com",
						IPcidr:      "172.28.0.0/20",
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
					UserName:         "user-79q6gZ9jD2Mw",
					Password:         "",
					TenantID:         "9adc45ea0a4e4d84a5acff1d829613e0",
					TenantName:       "6246671714361170",
					DomainID:         "default",
					Region:           "GRA9",
				},
			},
			title: "Network required/valid",
		},
		{
			input: `network:
  name: 'name'
  subnetCIDR: '172.28.0.0/20'
groupsHost:
  - namePattern: ''
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
  username: user-79q6gZ9jD2Mw
  password: ''
  region: GRA9
  tenantID: 9adc45ea0a4e4d84a5acff1d829613e0
  tenantName: '6246671714361170'
  domainID: default`,
			isError: true,
			expected: &config.Cloud{
				Network: config.Network{
					Name:       "name",
					SubnetCIDR: "172.28.0.0/20",
				},
				GroupsHost: []config.GroupHost{
					{
						NamePattern: "",
						IPcidr:      "172.28.0.0/20",
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
					UserName:         "user-79q6gZ9jD2Mw",
					Password:         "",
					TenantID:         "9adc45ea0a4e4d84a5acff1d829613e0",
					TenantName:       "6246671714361170",
					DomainID:         "default",
					Region:           "GRA9",
				},
			},
			title: "GrouHost valid",
		},
		{
			input: `network:
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
  dns: ''
  search: example.com
  postScripts:
    git:
      key: key
      url: git@github.com:SquareFactory/compute-configs.git
      ref: main
openstack:
  enabled: true
  identityEndpoint: https://auth.cloud.ovh.net/
  username: user-79q6gZ9jD2Mw
  password: ''
  region: GRA9
  tenantID: 9adc45ea0a4e4d84a5acff1d829613e0
  tenantName: '6246671714361170'
  domainID: default`,
			isError: true,
			expected: &config.Cloud{
				Network: config.Network{
					Name:       "name",
					SubnetCIDR: "172.28.0.0/20",
				},
				GroupsHost: []config.GroupHost{
					{
						NamePattern: "cn-s-[1-50].example.com",
						IPcidr:      "172.28.0.0/20",
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
					DNS:    "",
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
					UserName:         "user-79q6gZ9jD2Mw",
					Password:         "",
					TenantID:         "9adc45ea0a4e4d84a5acff1d829613e0",
					TenantName:       "6246671714361170",
					DomainID:         "default",
					Region:           "GRA9",
				},
			},
			title: "cloudConfig valid",
		},
		{
			input: `network:
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
  identityEndpoint: aaa
  username: user-79q6gZ9jD2Mw
  password: ''
  region: GRA9
  tenantID: 9adc45ea0a4e4d84a5acff1d829613e0
  tenantName: '6246671714361170'
  domainID: default`,
			isError: true,
			expected: &config.Cloud{
				Network: config.Network{
					Name:       "name",
					SubnetCIDR: "172.28.0.0/20",
				},
				GroupsHost: []config.GroupHost{
					{
						NamePattern: "cn-s-[1-50].example.com",
						IPcidr:      "172.28.0.0/20",
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
					IdentityEndpoint: "aaa",
					UserName:         "user-79q6gZ9jD2Mw",
					Password:         "",
					TenantID:         "9adc45ea0a4e4d84a5acff1d829613e0",
					TenantName:       "6246671714361170",
					DomainID:         "default",
					Region:           "GRA9",
				},
			},
			title: "openstack valid",
		},
		{
			input: `network:
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
      ref: main`,
			isError: false,
			expected: &config.Cloud{
				Network: config.Network{
					Name:       "name",
					SubnetCIDR: "172.28.0.0/20",
				},
				GroupsHost: []config.GroupHost{
					{
						NamePattern: "cn-s-[1-50].example.com",
						IPcidr:      "172.28.0.0/20",
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
			},
			title: "Positive test, allow empty Openstack",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Arrange
			config := &config.Cloud{}
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

func TestCloudTestSuite(t *testing.T) {
	suite.Run(t, &CloudTestSuite{})
}
