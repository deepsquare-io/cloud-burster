package config

type Config struct {
	Network                 `yaml:"network"`
	GroupsHost              []GroupHost `yaml:"groupsHost"`
	Hosts                   []Host      `yaml:"hosts"`
	CloudConfigTemplateOpts `yaml:"cloudConfigOpts"`
}

type Network struct {
	Name       string `yaml:"name"`
	SubnetCIDR string `yaml:"subnetCIDR"`
}

type GroupHost struct {
	// NamePattern overrides the host name
	NamePattern string `yaml:"namePattern"`
	// IPcidr overrides the IP. Based on NamePattern, each host will have an IP allocated.
	IPcidr string `yaml:"ipCIDR"`
	// HostTemplate defines helps to define a Host
	HostTemplate Host
}

type Host struct {
	Name       string `yaml:"name"`
	DiskSize   int    `yaml:"diskSize"`
	FlavorName string `yaml:"flavorName"`
	ImageName  string `yaml:"imageName"`
	IP         string `yaml:"ip"`
}

type CloudConfigTemplateOpts struct {
	AuthorizedKeys []string        `yaml:"authorizedKeys"`
	DNS            string          `yaml:"dns"`
	Search         string          `yaml:"search"`
	PostScripts    PostScriptsOpts `yaml:"postScripts"`
}

type PostScriptsOpts struct {
	Git GitOpts `yaml:"git"`
}

type GitOpts struct {
	Key string `yaml:"key"`
	URL string `yaml:"url"`
	Ref string `yaml:"ref"`
}
