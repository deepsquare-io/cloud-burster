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

type HostTestSuite struct {
	suite.Suite
}

func (suite *HostTestSuite) TestValidate() {
	tests := []struct {
		input    string
		expected *config.Host
		isError  bool
		title    string
	}{
		{
			input: `name: host
diskSize: 50
flavorName: flavor
imageName: image
ip: 172.24.0.0`,
			expected: &config.Host{
				Name:       "host",
				DiskSize:   50,
				FlavorName: "flavor",
				ImageName:  "image",
				IP:         "172.24.0.0",
			},
			title: "Positive test",
		},
		{
			input: `diskSize: 50
flavorName: flavor
imageName: image`,
			expected: &config.Host{
				Name:       "",
				DiskSize:   50,
				FlavorName: "flavor",
				ImageName:  "image",
				IP:         "",
			},
			title: "Positive test without optional fields",
		},
		{
			input: `flavorName: flavor
imageName: image`,
			isError: true,
			expected: &config.Host{
				Name:       "",
				DiskSize:   0,
				FlavorName: "flavor",
				ImageName:  "image",
				IP:         "",
			},
			title: "Required disk size",
		},
		{
			input: `diskSize: 50
imageName: image`,
			isError: true,
			expected: &config.Host{
				Name:       "",
				DiskSize:   50,
				FlavorName: "",
				ImageName:  "image",
				IP:         "",
			},
			title: "Required flavor",
		},
		{
			input: `diskSize: 50
flavorName: flavor`,
			isError: true,
			expected: &config.Host{
				Name:       "",
				DiskSize:   50,
				FlavorName: "flavor",
				ImageName:  "",
				IP:         "",
			},
			title: "Required image",
		},
		{
			input: `diskSize: 50
imageName: image
flavorName: flavor
ip: ip`,
			isError: true,
			expected: &config.Host{
				Name:       "",
				DiskSize:   50,
				FlavorName: "flavor",
				ImageName:  "image",
				IP:         "ip",
			},
			title: "Valid IP",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Arrange
			config := &config.Host{}
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

func TestHostTestSuite(t *testing.T) {
	suite.Run(t, &HostTestSuite{})
}
