package generate

import (
	"fmt"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "generate",
	Usage: "Generate config files.",
	Subcommands: []*cli.Command{
		{
			Name:  "hosts",
			Usage: "Generate hosts file for a DNS server",
			Action: func(cCtx *cli.Context) error {
				// Parse config
				conf, err := config.ParseFile(cCtx.String("config.path"))
				if err != nil {
					return err
				}

				hosts, err := conf.GenerateHosts()
				if err != nil {
					return err
				}
				fmt.Print(hosts)
				return nil
			},
		},
	},
}
