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

type OpenstackTestSuite struct {
	suite.Suite
}

func (suite *OpenstackTestSuite) TestValidate() {
	tests := []struct {
		input    string
		expected *config.Openstack
		isError  bool
		title    string
	}{
		{
			input: `enabled: true
identityEndpoint: 'https://auth.cloud.ovh.net/'
user: user-79q6gZ9jD2Mw
password: ''
region: GRA9
tenantID: 9adc45ea0a4e4d84a5acff1d829613e0
tenantName: '6246671714361170'
domainID: default`,
			isError: false,
			expected: &config.Openstack{
				Enabled:          true,
				IdentityEndpoint: "https://auth.cloud.ovh.net/",
				User:             "user-79q6gZ9jD2Mw",
				Password:         "",
				TenantID:         "9adc45ea0a4e4d84a5acff1d829613e0",
				TenantName:       "6246671714361170",
				DomainID:         "default",
				Region:           "GRA9",
			},
			title: "Positive test",
		},
		{
			input: `enabled: false
identityEndpoint: ''
user: ''
password: ''
region: ''
tenantID: ''
tenantName: ''
domainID: ''`,
			isError: false,
			expected: &config.Openstack{
				Enabled:          false,
				IdentityEndpoint: "",
				User:             "",
				Password:         "",
				TenantID:         "",
				TenantName:       "",
				DomainID:         "",
				Region:           "",
			},
			title: "Positive test",
		},
		{
			input: `enabled: true
identityEndpoint: 'aaa'
user: user-79q6gZ9jD2Mw
password: ''
region: GRA9
tenantID: 9adc45ea0a4e4d84a5acff1d829613e0
tenantName: '6246671714361170'
domainID: default`,
			isError: true,
			expected: &config.Openstack{
				Enabled:          true,
				IdentityEndpoint: "aaa",
				User:             "user-79q6gZ9jD2Mw",
				Password:         "",
				TenantID:         "9adc45ea0a4e4d84a5acff1d829613e0",
				TenantName:       "6246671714361170",
				DomainID:         "default",
				Region:           "GRA9",
			},
			title: "Valid URL",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Arrange
			config := &config.Openstack{}
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

func TestOpenstackTestSuite(t *testing.T) {
	suite.Run(t, &OpenstackTestSuite{})
}
