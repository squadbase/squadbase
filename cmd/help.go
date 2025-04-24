package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/squadbase/squadbase/internal/ui"
	"github.com/urfave/cli/v2"
)

func HelpCommand() *cli.Command {
	return &cli.Command{
		Name:      "help",
		Aliases:   []string{"h"},
		Usage:     "Display help information about Squadbase CLI",
		Action:    showHelp,
		ArgsUsage: "[command]",
	}
}

func showHelp(c *cli.Context) error {
	if c.Args().Len() > 0 {
		commandName := c.Args().First()
		return showCommandHelp(c, commandName)
	}
	return ShowHelp(c)
}

func ShowHelp(c *cli.Context) error {
	w := c.App.Writer
	ui.PrintTitle("Command Line Interface")

	versionInfo := make(map[string]string)
	versionInfo["Version"] = c.App.Version
	versionInfo["Description"] = "A command-line tool for creating projects from templates"

	ui.PrintSummaryBox("Squadbase CLI", versionInfo)
	fmt.Fprintln(w, "")

	commandsInfo := make(map[string]string)
	commandsInfo["create [PROJECT_NAME]"] = "Create a new project from a template"
	commandsInfo["init [DIRECTORY]"] = "Initialize an existing directory with squadbase.yml"
	commandsInfo["help [COMMAND]"] = "Show help information"

	ui.PrintSummaryBox("💻 Available Commands", commandsInfo)
	fmt.Fprintln(w, "")

	usageExamples := make(map[string]string)
	usageExamples["Create a new project"] = "squad create [PROJECT_NAME]"
	usageExamples["Initialize a directory"] = "squad init [DIRECTORY]"
	usageExamples["Show version"] = "squad --version"
	usageExamples["Get help"] = "squad help"

	ui.PrintSummaryBox("🚀 Basic Usage", usageExamples)
	fmt.Fprintln(w, "")

	templatesInfo := make(map[string]string)
	templatesInfo["morph"] = "A Squadbase-based project template"
	templatesInfo["nextjs"] = "A Next.js project template"
	templatesInfo["streamlit"] = "A Streamlit project template"

	ui.PrintSummaryBox("🧩 Available Templates", templatesInfo)
	fmt.Fprintln(w, "")

	return nil
}

func showCommandHelp(c *cli.Context, commandName string) error {
	w := c.App.Writer
	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	switch commandName {
	case "create":
		fmt.Fprintf(w, "\n%s\n\n", green("CREATE COMMAND"))
		fmt.Fprintf(w, "%s\n\n", bold("squad create [PROJECT_NAME]"))
		fmt.Fprintln(w, "Create a new project from a template with the specified name.")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, bold("Arguments:"))
		fmt.Fprintln(w, "  PROJECT_NAME: (Optional) The name of the project to create. If not provided, you will be prompted for it.")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, bold("Example:"))
		fmt.Fprintf(w, "  %s\n", blue("# Create a new project with a specific name"))
		fmt.Fprintln(w, "  squad create my-awesome-project")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "  %s\n", blue("# Create a new project with an interactive prompt"))
		fmt.Fprintln(w, "  squad create")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, bold("Process:"))
		fmt.Fprintln(w, "  1. Specify a project name (or be prompted for one)")
		fmt.Fprintln(w, "  2. Select a template using arrow keys")
		fmt.Fprintln(w, "  3. The project will be created with the selected template")
		fmt.Fprintln(w, "  4. Option to initialize git repository")
		fmt.Fprintln(w, "")

	case "init":
		fmt.Fprintf(w, "\n%s\n\n", green("INIT COMMAND"))
		fmt.Fprintf(w, "%s\n\n", bold("squad init [DIRECTORY]"))
		fmt.Fprintln(w, "Initialize an existing directory with squadbase.yml configuration.")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, bold("Arguments:"))
		fmt.Fprintln(w, "  DIRECTORY: (Optional) The directory to initialize. If not provided, the current directory will be used.")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, bold("Example:"))
		fmt.Fprintf(w, "  %s\n", blue("# Initialize the current directory"))
		fmt.Fprintln(w, "  squad init")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "  %s\n", blue("# Initialize a specific directory"))
		fmt.Fprintln(w, "  squad init /path/to/my-project")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, bold("Process:"))
		fmt.Fprintln(w, "  1. Specify a directory (or use current directory)")
		fmt.Fprintln(w, "  2. Select a template")
		fmt.Fprintln(w, "  3. Configure runtime and package manager settings")
		fmt.Fprintln(w, "  4. Select deployment provider")
		fmt.Fprintln(w, "  5. A squadbase.yml file will be created in the specified directory")
		fmt.Fprintln(w, "")

	case "help":
		fmt.Fprintf(w, "\n%s\n\n", green("HELP COMMAND"))
		fmt.Fprintf(w, "%s\n\n", bold("squad help [COMMAND]"))
		fmt.Fprintln(w, "Show help information about Squadbase CLI or a specific command.")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, bold("Arguments:"))
		fmt.Fprintln(w, "  COMMAND: (Optional) The command to show help for. If not provided, general help is shown.")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, bold("Example:"))
		fmt.Fprintf(w, "  %s\n", blue("# Show general help"))
		fmt.Fprintln(w, "  squad help")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "  %s\n", blue("# Show help for a specific command"))
		fmt.Fprintln(w, "  squad help create")
		fmt.Fprintln(w, "")

	default:
		fmt.Fprintf(w, "\nError: Command '%s' not found.\n", commandName)
		fmt.Fprintln(w, "Run 'squad help' to see available commands.")
	}

	return nil
}
