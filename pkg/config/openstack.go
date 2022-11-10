package config

import "github.com/squarefactory/cloud-burster/validate"

type Openstack struct {
	Enabled          bool   `yaml:"enabled"`
	IdentityEndpoint string `yaml:"identityEndpoint" validate:"omitempty,url"`
	User             string `yaml:"user" validate:"omitempty"`
	Password         string `yaml:"password" validate:"omitempty"`
	TenantID         string `yaml:"tenantID" validate:"omitempty"`
	TenantName       string `yaml:"tenantName" validate:"omitempty"`
	DomainID         string `yaml:"domainID" validate:"omitempty"`
	Region           string `yaml:"region" validate:"omitempty"`
}

func (c *Openstack) Validate() error {
	return validate.I.Struct(c)
}
