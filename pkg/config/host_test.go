//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
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
		input   *config.Host
		isError bool
		title   string
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
			input: &config.Host{
				FlavorName: cleanHost.FlavorName,
				ImageName:  cleanHost.ImageName,
			},
			title: "Required disk size",
		},
		{
			isError: true,
			input: &config.Host{
				DiskSize:  cleanHost.DiskSize,
				ImageName: cleanHost.ImageName,
			},
			title: "Required flavor",
		},
		{
			isError: true,
			input: &config.Host{
				DiskSize:   cleanHost.DiskSize,
				FlavorName: cleanHost.FlavorName,
			},
			title: "Required image",
		},
		{
			isError: true,
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
				logger.I.Info("expected error", zap.Error(err))
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func TestHostTestSuite(t *testing.T) {
	suite.Run(t, &HostTestSuite{})
}
