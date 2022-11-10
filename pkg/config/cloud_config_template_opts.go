package config

import (
	"github.com/squarefactory/cloud-burster/validate"
)

type CloudConfigTemplateOpts struct {
	AuthorizedKeys []string        `yaml:"authorizedKeys"`
	DNS            string          `yaml:"dns" validate:"required,ip"`
	Search         string          `yaml:"search,omitempty" validate:"omitempty"`
	PostScripts    PostScriptsOpts `yaml:"postScripts,omitempty" validate:"omitempty"`
}

func (c *CloudConfigTemplateOpts) Validate() error {
	return validate.I.Struct(c)
}

type PostScriptsOpts struct {
	Git GitOpts `yaml:"git,omitempty" validate:"omitempty"`
}

type GitOpts struct {
	Key string `yaml:"key,omitempty" validate:"omitempty"`
	URL string `yaml:"url,omitempty" validate:"omitempty"`
	Ref string `yaml:"ref,omitempty" validate:"omitempty"`
}
