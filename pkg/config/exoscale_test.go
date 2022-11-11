//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var cleanExoscale = config.Exoscale{
	ComputeEndpoint: "https://api.exoscale.com/compute/",
	APIKey:          "key",
	APISecret:       "secret",
	Zone:            "zone",
}

type ExoscaleTestSuite struct {
	suite.Suite
}

func (suite *ExoscaleTestSuite) TestValidate() {
	tests := []struct {
		input   *config.Exoscale
		isError bool
		title   string
	}{
		{
			isError: false,
			input:   &cleanExoscale,
			title:   "Positive test",
		},
		{
			isError: false,
			input:   &config.Exoscale{},
			title:   "Positive test: Empty fields",
		},
		{
			isError: true,
			input: &config.Exoscale{
				ComputeEndpoint: "aaa",
				APIKey:          cleanExoscale.APIKey,
				APISecret:       cleanExoscale.APISecret,
				Zone:            cleanExoscale.Zone,
			},
			title: "Valid URL",
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

func TestExoscaleTestSuite(t *testing.T) {
	suite.Run(t, &ExoscaleTestSuite{})
}
