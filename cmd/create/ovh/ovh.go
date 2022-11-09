package ovh

import "github.com/urfave/cli/v2"

var flags = []cli.Flag{}

var Command = &cli.Command{
	Name:  "ovh",
	Usage: "Spawn a VM on OVH",
	Flags: flags,
}
