package main

import (
	"log"
	"os"

	"github.com/GabrielBG99/goget/commands"
	"github.com/urfave/cli/v2"
)

var Version = "dev"

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "Show goget1s version",
	}

	app := &cli.App{
		Name:    "goget",
		Usage:   "A download accelerator written in Go",
		Version: Version,
		Commands: []*cli.Command{
			commands.Single(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
