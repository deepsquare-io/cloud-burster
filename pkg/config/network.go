package config

import "github.com/squarefactory/cloud-burster/validate"

type Network struct {
	Name       string `yaml:"name" validate:"required"`
	SubnetCIDR string `yaml:"subnetCIDR" validate:"required,cidr"`
}

func (c *Network) Validate() error {
	return validate.I.Struct(c)
}
