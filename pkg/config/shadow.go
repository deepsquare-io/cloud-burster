package config

import "github.com/squarefactory/cloud-burster/validate"

type Shadow struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Zone     string `yaml:"zone"`
	SSHKey   string `yaml:"sshkey"`
}

func (c *Shadow) Validate() error {
	return validate.I.Struct(c)
}
