//go:build unit

package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
)

var cleanOpenstack = config.Openstack{
	IdentityEndpoint: "https://auth.cloud.ovh.net/",
	UserName:         "user",
	Password:         "password",
	TenantID:         "tenantID",
	TenantName:       "tenantName",
	DomainID:         "default",
	Region:           "GRA9",
}

type OpenstackTestSuite struct {
	suite.Suite
}

func (suite *OpenstackTestSuite) TestValidate() {
	tests := []struct {
		input         *config.Openstack
		isError       bool
		errorContains []string
		title         string
	}{
		{
			input: &cleanOpenstack,
			title: "Positive test",
		},
		{
			input: &config.Openstack{},
			title: "Positive test: Empty fields",
		},
		{
			isError: true,
			errorContains: []string{
				"url",
				"IdentityEndpoint",
			},
			input: &config.Openstack{
				IdentityEndpoint: "aaa",
				UserName:         cleanOpenstack.UserName,
				Password:         cleanOpenstack.Password,
				TenantID:         cleanOpenstack.TenantID,
				TenantName:       cleanOpenstack.TenantName,
				DomainID:         cleanOpenstack.DomainID,
				Region:           cleanOpenstack.Region,
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

func TestOpenstackTestSuite(t *testing.T) {
	suite.Run(t, &OpenstackTestSuite{})
}
