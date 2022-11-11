package delete

import (
	"context"
	"errors"
	"sync"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/cloud"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/utils/generators"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var Command = &cli.Command{
	Name:      "delete",
	Usage:     "Delete a VM on a public cloud.",
	ArgsUsage: "<hostname>",
	Action: func(cCtx *cli.Context) error {
		ctx := cCtx.Context
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

		logger.I.Info("Deleting...", zap.Any("hostnames", hostnames))

		// Parse config
		conf, err := config.ParseFile(cCtx.String("config.path"))
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		errChan := make(chan error)

		for _, hostname := range hostnames {
			wg.Add(1)

			go func(ctx *context.Context, hostname string, wg *sync.WaitGroup, errChan chan<- error) {
				defer wg.Done()
				// Search host and cloud by hostname
				host, cl, err := conf.SearchHostByHostName(hostname)
				if err != nil {
					errChan <- err
					return
				}

				// Instanciate the corresponding cloud
				cloudWorker, err := cloud.Create(cl)
				if err != nil {
					errChan <- err
					return
				}

				if err := cloudWorker.Delete(host.Name); err != nil {
					logger.I.Error(
						"couldn't create the host",
						zap.Any("host", host),
					)
					errChan <- err
					return
				}
			}(&ctx, hostname, &wg, errChan)
		}

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			if err != nil {
				return err
			}
		}

		logger.I.Info("Delete command successful.")

		return nil
	},
}
