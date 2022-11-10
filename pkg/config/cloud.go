package config

import "github.com/squarefactory/cloud-burster/validate"

type Cloud struct {
	Network                 `yaml:"network" validate:"required"`
	GroupsHost              []GroupHost `yaml:"groupsHost,omitempty" validate:"omitempty,dive"`
	Hosts                   []Host      `yaml:"hosts,omitempty" validate:"omitempty,dive"`
	CloudConfigTemplateOpts `yaml:"cloudConfig"`
	Openstack               `yaml:"openstack,omitempty" validate:"omitempty"`
}

func (c *Cloud) Validate() error {
	return validate.I.Struct(c)
}
