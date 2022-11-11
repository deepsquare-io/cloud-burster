package config

import "github.com/squarefactory/cloud-burster/validate"

type Cloud struct {
	Network                 `yaml:"network" validate:"required"`
	GroupsHost              []GroupHost `yaml:"groupsHost,omitempty" validate:"omitempty,dive"`
	Hosts                   []Host      `yaml:"hosts,omitempty" validate:"omitempty,dive"`
	Type                    string      `yaml:"type" validate:"required"`
	CloudConfigTemplateOpts `yaml:"cloudConfig"`
	*Openstack              `yaml:"openstack,omitempty" validate:"required_if=Type openstack,excluded_unless=Type openstack"`
	*Exoscale               `yaml:"exoscale,omitempty" validate:"required_if=Type exoscale,excluded_unless=Type exoscale"`
}

func (c *Cloud) Validate() error {
	return validate.I.Struct(c)
}
