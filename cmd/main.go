package main

import (
	"os"

	"github.com/squarefactory/cloud-burster/cmd/create"
	"github.com/squarefactory/cloud-burster/cmd/delete"
	"github.com/squarefactory/cloud-burster/cmd/generate"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:  "config.path",
		Value: "/etc/cloud-burster/config.yaml",
		EnvVars: []string{
			"CONFIG_PATH",
		},
		Aliases: []string{"c"},
		Action: func(ctx *cli.Context, s string) error {
			info, err := os.Stat(s)
			if err != nil {
				return err
			}
			perms := info.Mode().Perm()
			if perms&0o077 != 0 {
				logger.I.Fatal(
					"incorrect permisisons for config file, must be user-only",
					zap.String("config.path", s),
					zap.Stringer("permissions", perms),
				)
			}
			return nil
		},
	},
}

var app = &cli.App{
	Name:  "cloud-burster",
	Usage: "Burst into the cloud.",
	Flags: flags,
	Commands: []*cli.Command{
		create.Command,
		delete.Command,
		generate.Command,
	},
	Suggest: true,
}

func main() {
	if err := app.Run(os.Args); err != nil {
		logger.I.Fatal("app crashed", zap.Error(err))
	}
}
