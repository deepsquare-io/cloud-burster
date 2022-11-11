//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
)

var cleanExoscale = config.Exoscale{
	ComputeEndpoint: "https://api.exoscale.com/v1",
	APIKey:          "key",
	APISecret:       "secret",
	Zone:            "zone",
}

type ExoscaleTestSuite struct {
	suite.Suite
}

func (suite *ExoscaleTestSuite) TestValidate() {
	tests := []struct {
		input         *config.Exoscale
		isError       bool
		errorContains []string
		title         string
	}{
		{
			input: &cleanExoscale,
			title: "Positive test",
		},
		{
			input: &config.Exoscale{},
			title: "Positive test: Empty fields",
		},
		{
			isError: true,
			errorContains: []string{
				"url",
				"ComputeEndpoint",
			},
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

func TestExoscaleTestSuite(t *testing.T) {
	suite.Run(t, &ExoscaleTestSuite{})
}
