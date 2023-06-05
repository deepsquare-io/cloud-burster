//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
)

var cleanShadow = config.Shadow{
	Username: "username",
	Password: "password",
	Zone:     "zone",
	SSHKey:   "sshkey",
}

type ShadowTestSuite struct {
	suite.Suite
}

func (suite *ShadowTestSuite) TestValidate() {
	tests := []struct {
		input         *config.Shadow
		isError       bool
		errorContains []string
		title         string
	}{
		{
			input: &cleanShadow,
			title: "Positive test",
		},
		{
			input: &config.Shadow{},
			title: "Positive test: Empty fields",
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

func TestShadowTestSuite(t *testing.T) {
	suite.Run(t, &ShadowTestSuite{})
}
