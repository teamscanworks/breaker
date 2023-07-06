package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/teamscanworks/breaker/api"
	"github.com/teamscanworks/breaker/breakerclient"
	"github.com/teamscanworks/breaker/config"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// create, and execute the breaker-cli application
func RunCLI() {
	app := cli.NewApp()
	app.Name = "breaker-cli"
	app.Usage = "circuit breaker client library and api server"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config.path",
			Usage: "path to yaml configuration file",
			Value: "config.yaml",
		},
		&cli.BoolFlag{
			Name:  "debug.log",
			Usage: "enable debug logging",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:  "api",
			Usage: "api management commands",
			Subcommands: []*cli.Command{
				{
					Name:  "issue-jwt",
					Usage: "encodes a new jwt for use with the api server",
					Action: func(cCtx *cli.Context) error {
						cfgPath := cCtx.String("config.path")
						cfg, err := config.LoadConfig(cfgPath)
						if err != nil {
							return err
						}
						logger, err := cfg.ZapLogger(cCtx.Bool("debug.log"))
						if err != nil {
							return err
						}
						jwt := api.NewJWT(
							cfg.API.Password,
							cfg.API.IdentifierField,
							cfg.API.TokenValidityDurationSeconds,
						)
						tkn, err := jwt.Encode(cCtx.String("identifier.value"), nil)
						if err != nil {
							return err
						}
						logger.Info("issued token", zap.String("jwt.token", tkn))
						return nil
					},
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "identifier.value",
							Usage: "value to use as the identifier",
						},
					},
				},
				{
					Name:  "start",
					Usage: "start the api server",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "dry.run",
							Usage: "do not broadcast transactions",
						},
					},
					Action: func(cCtx *cli.Context) error {
						ctx, cancel := context.WithCancel(cCtx.Context)
						cfgPath := cCtx.String("config.path")
						dryRun := cCtx.Bool("dry.run")
						cfg, err := config.LoadConfig(cfgPath)
						if err != nil {
							cancel()
							return err
						}
						apiOpts := cfg.ApiOpts(dryRun)
						logger, err := cfg.ZapLogger(cCtx.Bool("debug.log"))
						if err != nil {
							cancel()
							return err
						}
						jwt := api.NewJWT(
							cfg.API.Password,
							cfg.API.IdentifierField,
							cfg.API.TokenValidityDurationSeconds,
						)
						apiServer, err := api.NewAPI(
							ctx,
							logger,
							jwt,
							apiOpts,
						)
						if err != nil {
							cancel()
							return err
						}
						quitChannel := make(chan os.Signal, 1)
						signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
						var wg sync.WaitGroup
						wg.Add(1)
						go func() {
							<-quitChannel
							logger.Info("caught exit signal")
							cancel()
							apiServer.Close()
							wg.Done()
						}()
						if dryRun {
							if err := apiServer.Serve(); err != nil {
								logger.Error("api serve encountered error", zap.Error(err))
							}
						} else {
							bc, err := breakerclient.NewBreakerClient(
								ctx,
								logger,
								&cfg.Compass,
							)
							if err != nil {
								// clear resources
								quitChannel <- os.Interrupt
								return err
							}
							apiServer.WithBreakerClient(bc)
							if err := apiServer.Serve(); err != nil {
								logger.Error("api serve encountered error", zap.Error(err))
							}
						}
						wg.Wait()
						return nil
					},
				},
			},
		},
		{
			Name:  "config",
			Usage: "configuration management",
			Subcommands: []*cli.Command{
				{
					Name:  "new",
					Usage: "generate a new configuration file",
					Action: func(cCtx *cli.Context) error {
						cfgPath := cCtx.String("config.path")
						return config.NewConfig(cfgPath)
					},
				},
				{
					Name:  "list-active-keypair",
					Usage: "print the keypair actively in use for signing transactions",
					Action: func(cCtx *cli.Context) error {
						cfgPath := cCtx.String("config.path")
						cfg, err := config.LoadConfig(cfgPath)
						if err != nil {
							return err
						}
						logger, err := cfg.ZapLogger(cCtx.Bool("debug.log"))
						if err != nil {
							return err
						}
						bc, err := breakerclient.NewBreakerClient(cCtx.Context, logger, &cfg.Compass)
						if err != nil {
							return err
						}
						kp, err := bc.GetActiveKeypair()
						if err != nil {
							return err
						}
						if kp == nil {
							logger.Warn("no active keypair")
						} else {
							logger.Info("found active keypair", zap.String("address", kp.String()))
						}
						return nil
					},
				},
				{
					Name:  "new-key",
					Usage: "create a new keypair",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "key.name",
							Usage: "name to refer to the keypair with",
						},
						&cli.BoolFlag{
							Name:  "create.mnemonic",
							Usage: "if present, create a keypair via a new mnemonic phrase",
						},
					},
					Action: func(cCtx *cli.Context) error {
						cfgPath := cCtx.String("config.path")
						cfg, err := config.LoadConfig(cfgPath)
						if err != nil {
							return err
						}
						logger, err := cfg.ZapLogger(cCtx.Bool("debug.log"))
						if err != nil {
							return err
						}
						bc, err := breakerclient.NewBreakerClient(cCtx.Context, logger, &cfg.Compass)
						if err != nil {
							return err
						}
						if cCtx.Bool("create.mnemonic") {
							logger.Info("creating mnemonic")
							mnemonic, err := bc.NewMnemonic(cCtx.String("key.name"))
							if err != nil {
								return err
							}
							fmt.Println("mnemonic ", mnemonic)
						} else {
							return fmt.Errorf("invalid options")
						}
						return nil
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
