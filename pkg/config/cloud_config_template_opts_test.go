//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var cleanCloudConfigTemplateOpts = config.CloudConfigTemplateOpts{
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
}

type CloudConfigTemplateOptsTestSuite struct {
	suite.Suite
}

func (suite *CloudConfigTemplateOptsTestSuite) TestValidate() {
	tests := []struct {
		input   *config.CloudConfigTemplateOpts
		isError bool
		title   string
	}{
		{
			isError: false,
			input:   &cleanCloudConfigTemplateOpts,
			title:   "Positive test",
		},
		{
			isError: true,
			input: &config.CloudConfigTemplateOpts{
				AuthorizedKeys: cleanCloudConfigTemplateOpts.AuthorizedKeys,
				DNS:            "",
				Search:         cleanCloudConfigTemplateOpts.Search,
				PostScripts:    cleanCloudConfigTemplateOpts.PostScripts,
			},
			title: "DNS required",
		},
		{
			isError: true,
			input: &config.CloudConfigTemplateOpts{
				AuthorizedKeys: cleanCloudConfigTemplateOpts.AuthorizedKeys,
				DNS:            "aaa",
				Search:         cleanCloudConfigTemplateOpts.Search,
				PostScripts:    cleanCloudConfigTemplateOpts.PostScripts,
			},
			title: "DNS valid",
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

func TestCloudConfigTemplateOptsTestSuite(t *testing.T) {
	suite.Run(t, &CloudConfigTemplateOptsTestSuite{})
}
