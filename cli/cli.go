package cli

import (
	"fmt"

	"github.com/teamscanworks/breaker/api"
	"github.com/teamscanworks/breaker/breakerclient"
	"github.com/teamscanworks/breaker/config"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

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
	}
	app.Commands = []*cli.Command{
		&cli.Command{
			Name:  "api",
			Usage: "api management commands",
			Subcommands: []*cli.Command{
				&cli.Command{
					Name:  "issue-jwt",
					Usage: "encodes a new jwt for use with the api server",
					Action: func(cCtx *cli.Context) error {
						cfgPath := cCtx.String("config.path")
						cfg, err := config.LoadConfig(cfgPath)
						if err != nil {
							return err
						}
						logger, err := cfg.ZapLogger()
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
				&cli.Command{
					Name:  "start",
					Usage: "start the api server",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "dry.run",
							Usage: "do not broadcast transactions",
						},
					},
					Action: func(cCtx *cli.Context) error {
						cfgPath := cCtx.String("config.path")
						dryRun := cCtx.Bool("dry.run")
						cfg, err := config.LoadConfig(cfgPath)
						if err != nil {
							return err
						}
						apiOpts := cfg.ApiOpts(dryRun)
						logger, err := cfg.ZapLogger()
						if err != nil {
							return err
						}
						jwt := api.NewJWT(
							cfg.API.Password,
							cfg.API.IdentifierField,
							cfg.API.TokenValidityDurationSeconds,
						)
						apiServer, err := api.NewAPI(
							cCtx.Context,
							logger,
							jwt,
							apiOpts,
						)
						if err != nil {
							return err
						}
						logger.Info("TODO: enable catching unix signals to trigger api exit")
						if dryRun {
							return apiServer.Serve()
						} else {
							bc, err := breakerclient.NewBreakerClient(
								cCtx.Context,
								logger,
								&cfg.Compass,
							)
							if err != nil {
								return err
							}
							apiServer.WithBreakerClient(bc)
							return apiServer.Serve()
						}
					},
				},
			},
		},
		&cli.Command{
			Name:  "config",
			Usage: "configuration management",
			Subcommands: []*cli.Command{
				&cli.Command{
					Name:  "new",
					Usage: "generate a new configuration file",
					Action: func(cCtx *cli.Context) error {
						cfgPath := cCtx.String("config.path")
						return config.NewConfig(cfgPath)
					},
				},
				&cli.Command{
					Name:  "new-key",
					Usage: "create a new keypair",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "key.name",
							Usage: "name to refer to the keypair with",
						},
					},
					Action: func(cCtx *cli.Context) error {
						fmt.Println("todo")
						return nil
					},
				},
				&cli.Command{
					Name:  "import-key",
					Usage: "import a pre-existing keypair",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "key.name",
							Usage: "name to refer to the keypair with",
						},
					},
					Action: func(cCtx *cli.Context) error {
						fmt.Println("todo")
						return nil
					},
				},
			},
		},
	}
}
