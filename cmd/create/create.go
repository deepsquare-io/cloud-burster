package create

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

var flags = []cli.Flag{}

var Command = &cli.Command{
	Name:      "create",
	Usage:     "Spawn a VM on a public cloud.",
	Flags:     flags,
	ArgsUsage: "<hostnames>",
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

		// Parse config
		conf, err := config.ParseFile(cCtx.String("config.path"))
		if err != nil {
			return err
		}
		if err := conf.Validate(); err != nil {
			return err
		}

		logger.I.Info("Creating...", zap.Any("hostnames", hostnames))

		var wg sync.WaitGroup
		errChan := make(chan error)

		for _, hostname := range hostnames {
			wg.Add(1)
			go func(ctx context.Context, hostname string, wg *sync.WaitGroup, errChan chan<- error) {
				defer wg.Done()

				// Search host and cloud by hostname
				var host *config.Host
				var cl *config.Cloud
				var err error

				// Search hosts using hostname and suffix
				for _, suffix := range conf.SuffixSearch {
					host, cl, err = conf.SearchHostByHostName(hostname + suffix)
					if err != nil {
						errChan <- err
						return
					}
					if host != nil && cl != nil {
						break
					}
				}

				// If host is nil, default to search using hostname
				if host == nil && cl == nil {
					host, cl, err = conf.SearchHostByHostName(hostname)
					if err != nil {
						errChan <- err
						return
					}
				}

				// If host is still nil, crash
				if host == nil && cl == nil {
					errChan <- errors.New("hostname not found")
					return
				}

				// Instanciate the corresponding cloud
				cloudWorker, err := cloud.New(cl)
				if err != nil {
					errChan <- err
					return
				}

				if err := cloudWorker.Create(ctx, host, cl); err != nil {
					logger.I.Warn(
						"couldn't create the host",
						zap.Error(err),
						zap.Any("host", host),
						zap.Any("cloud", cl),
					)
					errChan <- err
					return
				}
			}(
				ctx,
				hostname,
				&wg,
				errChan,
			)
		}

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for e := range errChan {
			if e != nil {
				logger.I.Error("create thrown an error", zap.Error(e))
				err = e
			}
		}
		if err != nil {
			return err
		}

		logger.I.Info("Create command successful.")

		return nil
	},
}
