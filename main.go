package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonathongardner/fife/cli"
	"github.com/sirupsen/logrus"
)

func main() {
	app := cli.App()
	ctx := listenforCtrlC(context.Background())

	if err := app.Run(ctx, os.Args); err != nil {
		logrus.WithError(err).Fatal("Failed to run app")
	}
}

func listenforCtrlC(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// ctx := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigChan {
			logrus.Warnf("Recieved signal: %v", sig)
			cancel()
		}
	}()

	return ctx
}
