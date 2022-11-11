package config

import "github.com/squarefactory/cloud-burster/validate"

type Openstack struct {
	IdentityEndpoint string `yaml:"identityEndpoint" validate:"omitempty,url"`
	UserName         string `yaml:"username"`
	Password         string `yaml:"password"`
	TenantID         string `yaml:"tenantID"`
	TenantName       string `yaml:"tenantName"`
	DomainID         string `yaml:"domainID"`
	Region           string `yaml:"region"`
}

func (c *Openstack) Validate() error {
	return validate.I.Struct(c)
}
