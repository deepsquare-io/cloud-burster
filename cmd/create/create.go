package create

import (
	"github.com/urfave/cli/v2"
)

var flags = []cli.Flag{}

var Command = &cli.Command{
	Name:  "create",
	Usage: "Spawn a VM on a public cloud",
	Flags: flags,
}
