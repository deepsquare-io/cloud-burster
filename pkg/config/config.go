package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/validate"
	"go.uber.org/zap"

	"gopkg.in/yaml.v3"
)

const APIVersion = "cloud-burster.squarefactory.io/v1alpha1"

// equalAPIVersion is a validator to check API version compatibility
func equalAPIVersion(fl validator.FieldLevel) bool {
	return fl.Field().String() == APIVersion
}

func init() {
	if err := validate.I.RegisterValidation("equalAPI", equalAPIVersion); err != nil {
		logger.I.Fatal("couldn't register validation", zap.Any("validator", equalAPIVersion))
	}
}

type Config struct {
	APIVersion   string   `yaml:"apiVersion" validate:"equalAPI"`
	Clouds       []Cloud  `yaml:"clouds" validate:"dive"`
	SuffixSearch []string `yaml:"suffixSearch"`
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

	return foundHost, foundCloud, nil
}

// GenerateHosts generates hosts for the DNS server
func (c *Config) GenerateHosts() (string, error) {
	var sb strings.Builder

	// Search in all clouds
	for _, cloud := range c.Clouds {
		logger.I.Debug("found cloud", zap.Any("cloud", cloud))
		// Search in all hosts
		for _, host := range cloud.Hosts {
			logger.I.Debug("found host", zap.Any("host", host))
			sb.WriteString(fmt.Sprintf("%s %s\n", host.IP, host.Name))
		}

		// Search in all groups host
		for _, groupsHost := range cloud.GroupsHost {
			logger.I.Debug("found groupsHost", zap.Any("groupsHost", groupsHost))
			hosts, err := groupsHost.GenerateHosts()
			if err != nil {
				return "", err
			}
			for _, host := range hosts {
				logger.I.Debug("found host", zap.Any("host", host))
				sb.WriteString(fmt.Sprintf("%s %s\n", host.IP, host.Name))
			}
		}
	}

	return sb.String(), nil
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
