package config

type SSH struct {
	AuthorizedKeys []string
}

type Network struct {
	Name string
	SubnetCIDR string
}

type Host struct {
	Name string
	OS string
	DiskSize int
	FlavorName string
	ImageName string
	IP string
	Netmask string
	DNS string
	Search string
}
