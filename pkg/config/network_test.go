//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var cleanNetwork = config.Network{
	Name:       "net",
	SubnetCIDR: "172.28.0.0/20",
}

type NetworkTestSuite struct {
	suite.Suite
}

func (suite *NetworkTestSuite) TestValidate() {
	tests := []struct {
		input   *config.Network
		isError bool
		title   string
	}{
		{
			input: &cleanNetwork,
			title: "Positive test",
		},
		{
			isError: true,
			input: &config.Network{
				SubnetCIDR: cleanNetwork.SubnetCIDR,
			},
			title: "Required name",
		},
		{
			isError: true,
			input: &config.Network{
				Name: cleanNetwork.Name,
			},
			title: "Required CIDR",
		},
		{
			isError: true,
			input: &config.Network{
				Name:       cleanNetwork.Name,
				SubnetCIDR: "aaa",
			},
			title: "Valid CIDR",
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

func TestNetworkTestSuite(t *testing.T) {
	suite.Run(t, &NetworkTestSuite{})
}
