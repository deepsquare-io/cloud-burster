package config

import "github.com/squarefactory/cloud-burster/validate"

type Cloud struct {
	AuthorizedKeys []string        `yaml:"authorizedKeys"`
	PostScripts    PostScriptsOpts `yaml:"postScripts,omitempty"  validate:"omitempty"`
	Type           string          `yaml:"type"                   validate:"required"`
	Network        `yaml:"network" validate:"required"`
	GroupsHost     []GroupHost                 `yaml:"groupsHost,omitempty"   validate:"omitempty,dive"`
	Hosts          []Host                      `yaml:"hosts,omitempty"        validate:"omitempty,dive"`
	CustomConfig   map[interface{}]interface{} `yaml:"customConfig,omitempty" validate:"omitempty"`
	*Openstack     `yaml:"openstack,omitempty" validate:"required_if=Type openstack,excluded_unless=Type openstack"`
	*Exoscale      `yaml:"exoscale,omitempty" validate:"required_if=Type exoscale,excluded_unless=Type exoscale"`
	*Shadow        `yaml:"shadow,omitempty" validate:"required_if=Type shadow,excluded_unless=Type shadow"`
}

type PostScriptsOpts struct {
	Git GitOpts `yaml:"git,omitempty" validate:"omitempty"`
}

type GitOpts struct {
	Key string `yaml:"key,omitempty" validate:"omitempty"`
	URL string `yaml:"url,omitempty" validate:"omitempty"`
	Ref string `yaml:"ref,omitempty" validate:"omitempty"`
}

func (c *Cloud) Validate() error {
	return validate.I.Struct(c)
}
