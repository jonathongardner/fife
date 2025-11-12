package cli

import (
	"context"
	"fmt"

	"github.com/jonathongardner/fife/app"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func App() *cli.Command {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Println(app.Version)
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "print the version",
	}

	return &cli.Command{
		Name:    "fife",
		Version: app.Version,
		Usage:   "App for creating reverse proxy",
		Commands: []*cli.Command{
			reverseProx,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "logging level",
			},
			&cli.BoolFlag{
				Name:  "no-json",
				Usage: "log plain text",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Bool("verbose") {
				logrus.SetLevel(logrus.DebugLevel)
				logrus.Debug("Setting to debug...")
			}
			if !cmd.Bool("no-json") {
				logrus.SetFormatter(&logrus.JSONFormatter{})
			}
			return ctx, nil
		},
	}
}
