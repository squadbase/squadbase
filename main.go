package main

import (
	"os"

	"github.com/squadbase/squadbase/cmd"
	"github.com/squadbase/squadbase/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "squad",
		Usage:                "A command-line tool for creating projects from templates",
		Version:              version.GetFullVersion(),
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			cmd.InitCommand(),
			cmd.CreateCommand(),
			cmd.HelpCommand(),
		},
		Action: func(c *cli.Context) error {
			return cmd.ShowHelp(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
