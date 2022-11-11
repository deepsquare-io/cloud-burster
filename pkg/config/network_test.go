//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
)

var cleanNetwork = config.Network{
	Name:       "net",
	SubnetCIDR: "172.28.0.0/20",
	DNS:        "1.1.1.1",
	Search:     "example.com",
	Gateway:    "172.28.0.2",
}

type NetworkTestSuite struct {
	suite.Suite
}

func (suite *NetworkTestSuite) TestValidate() {
	tests := []struct {
		input         *config.Network
		isError       bool
		errorContains []string
		title         string
	}{
		{
			input: &cleanNetwork,
			title: "Positive test",
		},
		{
			isError: true,
			errorContains: []string{
				"required",
				"Name",
			},
			input: &config.Network{
				SubnetCIDR: cleanNetwork.SubnetCIDR,
				DNS:        cleanNetwork.DNS,
				Search:     cleanNetwork.Search,
				Gateway:    cleanNetwork.Gateway,
			},
			title: "Required name",
		},
		{
			isError: true,
			errorContains: []string{
				"required",
				"SubnetCIDR",
			},
			input: &config.Network{
				Name:    cleanNetwork.Name,
				DNS:     cleanNetwork.DNS,
				Search:  cleanNetwork.Search,
				Gateway: cleanNetwork.Gateway,
			},
			title: "Required CIDR",
		},
		{
			isError: true,
			errorContains: []string{
				"cidr",
				"SubnetCIDR",
			},
			input: &config.Network{
				Name:       cleanNetwork.Name,
				DNS:        cleanNetwork.DNS,
				Search:     cleanNetwork.Search,
				SubnetCIDR: "aaa",
				Gateway:    cleanNetwork.Gateway,
			},
			title: "Valid CIDR",
		},
		{
			isError: true,
			errorContains: []string{
				"required",
				"DNS",
			},
			input: &config.Network{
				Name:       cleanNetwork.Name,
				SubnetCIDR: cleanNetwork.SubnetCIDR,
				Search:     cleanNetwork.Search,
				Gateway:    cleanNetwork.Gateway,
			},
			title: "Required DNS",
		},
		{
			isError: true,
			errorContains: []string{
				"ip",
				"DNS",
			},
			input: &config.Network{
				Name:       cleanNetwork.Name,
				DNS:        "aaa",
				SubnetCIDR: cleanNetwork.SubnetCIDR,
				Search:     cleanNetwork.Search,
				Gateway:    cleanNetwork.Gateway,
			},
			title: "Valid DNS",
		},
		{
			isError: true,
			errorContains: []string{
				"required",
				"Gateway",
			},
			input: &config.Network{
				Name:       cleanNetwork.Name,
				SubnetCIDR: cleanNetwork.SubnetCIDR,
				Search:     cleanNetwork.Search,
				DNS:        cleanNetwork.DNS,
			},
			title: "Required Gateway",
		},
		{
			isError: true,
			errorContains: []string{
				"ip",
				"Gateway",
			},
			input: &config.Network{
				Name:       cleanNetwork.Name,
				DNS:        cleanNetwork.DNS,
				SubnetCIDR: cleanNetwork.SubnetCIDR,
				Search:     cleanNetwork.Search,
				Gateway:    "aaa",
			},
			title: "Valid Gateway",
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

func TestNetworkTestSuite(t *testing.T) {
	suite.Run(t, &NetworkTestSuite{})
}
