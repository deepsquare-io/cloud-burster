package create

import (
	"errors"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/cloud"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/utils/generators"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var flags = []cli.Flag{}

var Command = &cli.Command{
	Name:      "create",
	Usage:     "Spawn a VM on a public cloud.",
	Flags:     flags,
	ArgsUsage: "<hostnames>",
	Action: func(cCtx *cli.Context) error {
		if cCtx.NArg() < 1 {
			return errors.New("not enough arguments")
		}
		arg := cCtx.Args().Get(0)
		hostnames := generators.SplitCommaOutsideOfBrackets(arg)

		// Parse config
		conf, err := config.ParseFile(cCtx.String("config.path"))
		if err != nil {
			return err
		}

		for _, hostname := range hostnames {
			// Search host and cloud by hostname
			host, cl, err := conf.SearchHostByHostName(hostname)
			if err != nil {
				return err
			}

			// Instanciate the corresponding cloud
			cloudWorker, err := cloud.Create(cl)
			if err != nil {
				return err
			}

			if err := cloudWorker.Create(host, &cl.Network, &cl.CloudConfigTemplateOpts); err != nil {
				logger.I.Error(
					"couldn't create the host",
					zap.Any("host", host),
					zap.Any("network", cl.Network),
					zap.Any("cloudConfig", cl.CloudConfigTemplateOpts),
				)
				return err
			}
		}

		return nil
	},
}
