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

type CloudConfigTemplateOptsTestSuite struct {
	suite.Suite
}

func (suite *CloudConfigTemplateOptsTestSuite) TestValidate() {
	tests := []struct {
		input    string
		expected *config.CloudConfigTemplateOpts
		isError  bool
		title    string
	}{
		{
			input: `authorizedKeys:
  - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4
dns: 1.1.1.1
search: example.com
postScripts:
  git:
    key: key
    url: git@github.com:SquareFactory/compute-configs.git
    ref: main`,
			isError: false,
			expected: &config.CloudConfigTemplateOpts{
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
			title: "Positive test",
		},
		{
			input: `authorizedKeys:
  - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4
dns: ''
search: example.com
postScripts:
  git:
    key: key
    url: git@github.com:SquareFactory/compute-configs.git
    ref: main`,
			isError: true,
			expected: &config.CloudConfigTemplateOpts{
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
			title: "DNS required",
		},
		{
			input: `authorizedKeys:
  - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4
dns: 'aaa'
search: example.com
postScripts:
  git:
    key: key
    url: git@github.com:SquareFactory/compute-configs.git
    ref: main`,
			isError: true,
			expected: &config.CloudConfigTemplateOpts{
				AuthorizedKeys: []string{
					"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4",
				},
				DNS:    "aaa",
				Search: "example.com",
				PostScripts: config.PostScriptsOpts{
					Git: config.GitOpts{
						Key: "key",
						URL: "git@github.com:SquareFactory/compute-configs.git",
						Ref: "main",
					},
				},
			},
			title: "DNS valid",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Arrange
			config := &config.CloudConfigTemplateOpts{}
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

func TestCloudConfigTemplateOptsTestSuite(t *testing.T) {
	suite.Run(t, &CloudConfigTemplateOptsTestSuite{})
}
