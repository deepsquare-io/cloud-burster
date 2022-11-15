package generate

import (
	"errors"
	"fmt"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/utils/generators"
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
				if err := conf.Validate(); err != nil {
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
		{
			Name:      "hostnames",
			Usage:     "Generate hostnames from host pattern",
			ArgsUsage: "<hostnames>",
			Action: func(cCtx *cli.Context) error {
				if cCtx.NArg() < 1 {
					return errors.New("not enough arguments")
				}

				arg := cCtx.Args().Get(0)
				hostnamesRanges := generators.SplitCommaOutsideOfBrackets(arg)

				var hostnames []string
				for _, hostnamesRange := range hostnamesRanges {
					h := generators.ExpandBrackets(hostnamesRange)
					hostnames = append(hostnames, h...)
				}

				for _, h := range hostnames {
					fmt.Println(h)
				}

				return nil
			},
		},
	},
}
