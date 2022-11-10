package delete

import (
	"github.com/urfave/cli/v2"
)

var flags = []cli.Flag{}

var Command = &cli.Command{
	Name:  "delete",
	Usage: "Delete a VM on a public cloud",
	Flags: flags,
}
