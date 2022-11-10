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

type NetworkTestSuite struct {
	suite.Suite
}

func (suite *NetworkTestSuite) TestValidate() {
	tests := []struct {
		input    string
		expected *config.Network
		isError  bool
		title    string
	}{
		{
			input: `name: 'net'
subnetCIDR: '172.24.0.0/20'`,
			expected: &config.Network{
				Name:       "net",
				SubnetCIDR: "172.24.0.0/20",
			},
			title: "Positive test",
		},
		{
			input: `name: ''
subnetCIDR: '172.24.0.0/20'`,
			isError: true,
			expected: &config.Network{
				Name:       "",
				SubnetCIDR: "172.24.0.0/20",
			},
			title: "Required name",
		},
		{
			input: `name: 'net'
subnetCIDR: ''`,
			isError: true,
			expected: &config.Network{
				Name:       "net",
				SubnetCIDR: "",
			},
			title: "Required CIDR",
		},
		{
			input: `name: 'net'
subnetCIDR: 'aaa'`,
			isError: true,
			expected: &config.Network{
				Name:       "net",
				SubnetCIDR: "aaa",
			},
			title: "Valid CIDR",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Arrange
			config := &config.Network{}
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

func TestNetworkTestSuite(t *testing.T) {
	suite.Run(t, &NetworkTestSuite{})
}
