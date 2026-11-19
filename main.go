package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/Dopesonic/catalog-service/cmd"
	"github.com/Dopesonic/catalog-service/internal/app/constant"
	msentry "github.com/Dopesonic/catalog-service/internal/app/monitor/sentry"
)

func main() {
	app := &cli.App{
		Name:    constant.AppName,
		Version: constant.Version,
		Usage:   "Catalog management service",
		Commands: []*cli.Command{
			cmd.Migrate(),
			cmd.WebServer(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-json",
				Usage: "Human-readable log format",
			},
		},
	}

	defer msentry.Flush()

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
