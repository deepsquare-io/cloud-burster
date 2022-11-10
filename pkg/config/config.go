package config

import (
	"errors"
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

func (c *Config) SearchHostByHostName(hostname string) (*Host, *Cloud, error) {
	var found bool
	var foundCloud *Cloud
	var foundHost *Host

	// Search the corresponding cloud
	for _, cloud := range c.Clouds {
		// Search the corresponding host
		for _, host := range cloud.Hosts {
			if hostname == host.Name {
				found = true
				foundHost = &host
				foundCloud = &cloud
				break
			}
		}

		if found {
			break
		}

		// Not found yet, search the corresponding group host
		for _, groupsHost := range cloud.GroupsHost {
			hosts, err := groupsHost.GenerateHosts()
			if err != nil {
				return nil, nil, err
			}
			for _, host := range hosts {
				if hostname == host.Name {
					found = true
					foundHost = &host
					foundCloud = &cloud
					break
				}
			}

			if found {
				break
			}
		}

		if found {
			break
		}
	}

	if foundHost == nil || foundCloud == nil {
		return nil, nil, errors.New("host not found")
	}

	return foundHost, foundCloud, nil
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
