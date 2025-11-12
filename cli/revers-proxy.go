package cli

import (
	"context"
	"fmt"

	"github.com/jonathongardner/fife/server"
	"github.com/urfave/cli/v3"
)

var reverseProx = &cli.Command{
	Name:    "reverse-proxy",
	Aliases: []string{"rp"},
	Usage:   "Start a reverse proxy",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Path to config file",
			Value:   "./.fife.yaml",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cfg, err := server.LoadConfig(cmd.String("config"))
		if err != nil {
			return fmt.Errorf("error loading config %w", err)
		}

		if err := server.NewReverseProxy(ctx, cfg); err != nil {
			return fmt.Errorf("error with server %w", err)
		}

		return nil
	},
}
