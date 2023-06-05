package config

import "github.com/squarefactory/cloud-burster/validate"

type Network struct {
	Name       string `yaml:"name"             validate:"required"`
	SubnetCIDR string `yaml:"subnetCIDR"       validate:"required,cidr"`
	DNS        string `yaml:"dns"              validate:"required,ip"`
	Search     string `yaml:"search,omitempty" validate:"omitempty"`
	Gateway    string `yaml:"gateway"          validate:"required,ip"`
}

func (c *Network) Validate() error {
	return validate.I.Struct(c)
}
