package main

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/robinbraemer/devroach"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	globs := cli.NewStringSlice()
	dir := ""
	app := &cli.App{
		Name:  "devroach",
		Usage: "A CLI for starting a local in-memory CockroachDB for development and running auto-running migrations",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "migrations",
				Usage: "Glob patterns for migration files",
				Value: cli.NewStringSlice(
					"prisma/migrations/**/*.sql",
				),
				Destination: globs,
			},
			&cli.StringFlag{
				Name:        "dir",
				Usage:       "The root directory to search for migration files using the glob patterns",
				Value:       ".",
				Destination: &dir,
			},
		},
		Action: func(c *cli.Context) error {
			log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level:     slog.Level(-10),
				AddSource: true,
			}))
			ctx := logr.NewContextWithSlogLogger(c.Context, log)

			_, clean, err := devroach.NewPool(ctx, os.DirFS(dir), globs.Value()...)
			if err != nil {
				return err
			}
			defer clean()

			ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
			defer cancel()
			<-ctx.Done()
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
