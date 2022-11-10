package delete

import (
	"errors"

	"github.com/squarefactory/cloud-burster/pkg/cloud"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:      "delete",
	Usage:     "Delete a VM on a public cloud.",
	ArgsUsage: "<hostname>",
	Action: func(cCtx *cli.Context) error {
		if cCtx.NArg() < 1 {
			return errors.New("not enough arguments")
		}
		hostname := cCtx.Args().Get(0)

		// Parse config
		conf, err := config.ParseFile(cCtx.String("config.path"))
		if err != nil {
			return err
		}

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

		return cloudWorker.Delete(host.Name)
	},
}
