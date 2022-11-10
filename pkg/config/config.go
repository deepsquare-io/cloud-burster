package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/squarefactory/cloud-burster/validate"

	"gopkg.in/yaml.v3"
)

const APIVersion = "cloud-burster.squarefactory.io/v1alpha1"

// equalAPIVersion is a validator to check API version compatibility
func equalAPIVersion(fl validator.FieldLevel) bool {
	return fl.Field().String() == APIVersion
}

func init() {
	validate.I.RegisterValidation("equalAPI", equalAPIVersion)
}

type Config struct {
	APIVersion string  `yaml:"apiVersion" validate:"equalAPI"`
	Clouds     []Cloud `yaml:"clouds" validate:"dive"`
}

func (c *Config) Validate() error {
	return validate.I.Struct(c)
}

func ParseFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
