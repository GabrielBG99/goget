package main

import (
	"log"
	"os"

	"github.com/GabrielBG99/goget/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "goget",
		Usage: "A download accelerator written in Go",
		Commands: []*cli.Command{
			commands.Single(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
