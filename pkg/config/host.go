package config

import "github.com/squarefactory/cloud-burster/validate"

type Host struct {
	Name       string `yaml:"name,omitempty" validate:"omitempty"`
	DiskSize   int    `yaml:"diskSize"       validate:"required"`
	RAM        int    `yaml:"ram,omitempty" validate:"omitempty"`
	GPU        int    `yaml:"gpu,omitempty" validate:"omitempty"`
	FlavorName string `yaml:"flavorName"     validate:"required"`
	ImageName  string `yaml:"imageName"      validate:"required"`
	IP         string `yaml:"ip,omitempty"   validate:"omitempty,ip"`
}

func (c *Host) Validate() error {
	return validate.I.Struct(c)
}
