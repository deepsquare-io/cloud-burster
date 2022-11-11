package config

import "github.com/squarefactory/cloud-burster/validate"

type Exoscale struct {
	ComputeEndpoint string `yaml:"computeEndpoint" validate:"omitempty,url"`
	APIKey          string `yaml:"apiKey"`
	APISecret       string `yaml:"apiSecret"`
	Zone            string `yaml:"zone"`
}

func (c *Exoscale) Validate() error {
	return validate.I.Struct(c)
}
