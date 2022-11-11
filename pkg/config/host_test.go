//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
)

var cleanHost = config.Host{
	Name:       "host",
	DiskSize:   50,
	FlavorName: "d2-2",
	ImageName:  "Rocky Linux 9",
	IP:         "172.28.16.254",
}

type HostTestSuite struct {
	suite.Suite
}

func (suite *HostTestSuite) TestValidate() {
	tests := []struct {
		input         *config.Host
		isError       bool
		errorContains []string
		title         string
	}{
		{
			input: &cleanHost,
			title: "Positive test",
		},
		{
			input: &config.Host{
				DiskSize:   cleanHost.DiskSize,
				FlavorName: cleanHost.FlavorName,
				ImageName:  cleanHost.ImageName,
			},
			title: "Positive test without optional fields",
		},
		{
			isError: true,
			errorContains: []string{
				"required",
				"DiskSize",
			},
			input: &config.Host{
				FlavorName: cleanHost.FlavorName,
				ImageName:  cleanHost.ImageName,
			},
			title: "Required disk size",
		},
		{
			isError: true,
			errorContains: []string{
				"required",
				"FlavorName",
			},
			input: &config.Host{
				DiskSize:  cleanHost.DiskSize,
				ImageName: cleanHost.ImageName,
			},
			title: "Required flavor",
		},
		{
			isError: true,
			errorContains: []string{
				"required",
				"ImageName",
			},
			input: &config.Host{
				DiskSize:   cleanHost.DiskSize,
				FlavorName: cleanHost.FlavorName,
			},
			title: "Required image",
		},
		{
			isError: true,
			errorContains: []string{
				"ip",
				"IP",
			},
			input: &config.Host{
				DiskSize:   cleanHost.DiskSize,
				FlavorName: cleanHost.FlavorName,
				ImageName:  cleanHost.ImageName,
				IP:         "ip",
			},
			title: "Valid IP",
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

func TestHostTestSuite(t *testing.T) {
	suite.Run(t, &HostTestSuite{})
}
