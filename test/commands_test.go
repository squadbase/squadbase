package test

import (
	"bytes"
	"testing"

	"github.com/squadbase/squadbase/cmd"
	"github.com/urfave/cli/v2"
)

func setupApp() *cli.App {
	app := &cli.App{
		Name:    "squad",
		Usage:   "A command-line tool for creating projects from templates",
		Version: "0.1.0-test",
		Commands: []*cli.Command{
			cmd.InitCommand(),
			cmd.CreateCommand(),
			cmd.HelpCommand(),
		},
		Action: func(c *cli.Context) error {
			return cmd.ShowHelp(c)
		},
	}
	return app
}

func TestNoCommand(t *testing.T) {
	app := setupApp()
	var buf bytes.Buffer
	app.Writer = &buf

	err := app.Run([]string{"squad"})
	if err != nil {
		t.Errorf("Error running with no command: %v", err)
	}

	output := buf.String()
	expectedStrings := []string{
		"Squadbase CLI",
		"AVAILABLE COMMANDS:",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Errorf("Expected output to contain '%s', but it doesn't: %s", expected, output)
		}
	}
}
