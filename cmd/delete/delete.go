package delete

import (
	"github.com/squarefactory/cloud-burster/cmd/delete/ovh"
	"github.com/urfave/cli/v2"
)

var flags = []cli.Flag{}

var Command = &cli.Command{
	Name:  "delete",
	Usage: "Delete a VM on a public cloud",
	Flags: flags,
	Subcommands: []*cli.Command{
		ovh.Command,
	},
}
