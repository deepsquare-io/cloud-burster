package validate

import (
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "validate",
	Action: func(cCtx *cli.Context) error {
		// Parse config
		conf, err := config.ParseFile(cCtx.String("config.path"))
		if err != nil {
			return err
		}
		if err := conf.Validate(); err != nil {
			return err
		}

		logger.I.Info("Config is valid.")

		return nil
	},
}
