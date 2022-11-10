package config

import (
	"errors"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/utils/cidr"
	"github.com/squarefactory/cloud-burster/utils/generators"
	"github.com/squarefactory/cloud-burster/validate"
	"go.uber.org/zap"
)

type GroupHost struct {
	// NamePattern overrides the host name
	NamePattern string `yaml:"namePattern" validate:"required"`
	// IPcidr overrides the IP. Based on NamePattern, each host will have an IP allocated.
	IPcidr string `yaml:"ipCIDR" validate:"required,cidr"`
	// HostTemplate defines helps to define a Host
	HostTemplate Host `yaml:"template"`
}

func (g *GroupHost) GenerateHosts() ([]Host, error) {
	var out []Host

	// Generates names based on Name Pattern
	names := generators.ExpandBrackets(g.NamePattern)

	// Generates IPs
	ipAddresses := cidr.Hosts(g.IPcidr)

	if len(names) > len(ipAddresses) {
		logger.I.Error(
			"not enough IP addresses in CIDR",
			zap.String("namePattern", g.NamePattern),
			zap.Int("len(namePattern)", len(names)),
			zap.String("ipCIDR", g.IPcidr),
			zap.Int("len(ipAddresses)", len(ipAddresses)),
		)
		return []Host{}, errors.New("not enough IP addresses in CIDR")
	}

	// Map the names into host
	for idx, name := range names {
		host := Host{
			Name:       name,
			DiskSize:   g.HostTemplate.DiskSize,
			FlavorName: g.HostTemplate.FlavorName,
			ImageName:  g.HostTemplate.ImageName,
			IP:         ipAddresses[idx],
		}
		out = append(out, host)
	}

	return out, nil
}

func (g *GroupHost) Validate() error {
	return validate.I.Struct(g)
}
