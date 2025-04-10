package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"slices"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/squadbase/squadbase/internal/project"
	"github.com/squadbase/squadbase/internal/templates"
	"github.com/squadbase/squadbase/internal/ui"
	"github.com/urfave/cli/v2"
)

func InitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Aliases:   []string{"i"},
		Usage:     "Initialize an existing directory with squadbase.yml configuration",
		ArgsUsage: "[DIRECTORY]",
		Action:    initAction,
	}
}

func initAction(c *cli.Context) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	ui.PrintStep(1, 4, "Directory Selection")

	directory := c.Args().First()
	if directory == "" {
		var err error
		directory, err = os.Getwd()
		if err != nil {
			fmt.Printf("%s Error: Failed to get current directory\n", color.RedString("ERROR:"))
			return err
		}

		fmt.Printf("\nConfiguring current directory: %s\n", blue(directory))

		var confirmDir bool
		prompt := &survey.Confirm{
			Message: "Is this the correct directory to initialize?",
			Default: true,
		}
		err = survey.AskOne(prompt, &confirmDir)
		if err != nil {
			return fmt.Errorf("initialization cancelled")
		}

		if !confirmDir {
			dirPrompt := &survey.Input{
				Message: "Please specify the directory path to initialize:",
			}
			err = survey.AskOne(dirPrompt, &directory)
			if err != nil {
				return fmt.Errorf("initialization cancelled")
			}
		}
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Printf("%s Directory '%s' does not exist\n", color.RedString("ERROR:"), directory)
		return err
	}

	templates, err := templates.GetAvailableTemplates(false)
	if err != nil {
		fmt.Printf("%s Failed to get available templates: %v\n", color.RedString("ERROR:"), err)
		return err
	} else if len(templates) == 0 {
		fmt.Printf("%s No templates available\n", color.RedString("ERROR:"))
		return fmt.Errorf("no templates available")
	}

	templateNames := make([]string, 0, len(templates))
	for _, template := range templates {
		templateNames = append(templateNames, template.Name)
	}

	fmt.Printf("\n%s %s\n", cyan("Select a framework for your project:"), "🧩")
	fmt.Printf("  (Use %s arrows to navigate, %s to select)\n", yellow("↑/↓"), yellow("Enter"))

	var templateName string
	templatePrompt := &survey.Select{
		Message: "",
		Options: templateNames,
	}
	err = survey.AskOne(templatePrompt, &templateName)
	if err != nil {
		return fmt.Errorf("initialization cancelled")
	}

	var languageVersion string
	var packageManager string
	var deploymentProvider string

	if templateName == "squadbase" || templateName == "streamlit" {
		fmt.Printf("\n%s %s\n", cyan("Python Configuration"), "🐍")

		currentPyVersion := project.GetCurrentPythonVersion()
		supportedVersions := []string{"3.9", "3.10", "3.11", "3.12"}

		defaultVersion := "3.10"
		if slices.Contains(supportedVersions, currentPyVersion) {
			defaultVersion = currentPyVersion
		}

		versionPrompt := &survey.Select{
			Message: "Select Python version (supported: 3.9-3.12):",
			Options: supportedVersions,
			Default: defaultVersion,
		}
		err = survey.AskOne(versionPrompt, &languageVersion)
		if err != nil {
			return fmt.Errorf("initialization cancelled")
		}

		pmPrompt := &survey.Select{
			Message: "Select package manager:",
			Options: []string{"poetry", "uv", "pip"},
			Default: "poetry",
		}
		err = survey.AskOne(pmPrompt, &packageManager)
		if err != nil {
			return fmt.Errorf("initialization cancelled")
		}
	} else if templateName == "nextjs" {
		fmt.Printf("\n%s %s\n", cyan("Node.js Configuration"), "📦")

		currentNodeVersion := project.GetCurrentNodeVersion()
		supportedNodeVersions := []string{"16", "18", "20"}

		defaultNodeVersion := "18"
		if slices.Contains(supportedNodeVersions, currentNodeVersion) {
			defaultNodeVersion = currentNodeVersion
		}

		versionPrompt := &survey.Select{
			Message: "Select Node.js version (supported: 16, 18, 20):",
			Options: supportedNodeVersions,
			Default: defaultNodeVersion,
		}
		err = survey.AskOne(versionPrompt, &languageVersion)
		if err != nil {
			return fmt.Errorf("initialization cancelled")
		}

		pmPrompt := &survey.Select{
			Message: "Select package manager:",
			Options: []string{"npm", "yarn", "pnpm"},
			Default: "npm",
		}
		err = survey.AskOne(pmPrompt, &packageManager)
		if err != nil {
			return fmt.Errorf("initialization cancelled")
		}
	}

	fmt.Printf("\n%s %s\n", cyan("Deployment Configuration"), "🚀")

	deploymentOptions := []string{"aws", "gcp"}
	deploymentDefault := "aws"

	if templateName == "streamlit" || templateName == "nextjs" {
		deploymentOptions = []string{"gcp"}
		deploymentDefault = "gcp"
	}

	dpPrompt := &survey.Select{
		Message: "Select deployment provider:",
		Options: deploymentOptions,
		Default: deploymentDefault,
	}
	err = survey.AskOne(dpPrompt, &deploymentProvider)
	if err != nil {
		return fmt.Errorf("initialization cancelled")
	}

	fmt.Printf("\n%s\n", cyan("Configuration Summary:"))
	fmt.Printf("  %-20s %s\n", "Framework:", green(templateName))

	if languageVersion != "" {
		if templateName == "squadbase" || templateName == "streamlit" {
			fmt.Printf("  %-20s %s\n", "Python Version:", green(languageVersion))
		} else {
			fmt.Printf("  %-20s %s\n", "Node.js Version:", green(languageVersion))
		}
	}

	if packageManager != "" {
		fmt.Printf("  %-20s %s\n", "Package Manager:", green(packageManager))
	}

	fmt.Printf("  %-20s %s\n", "Deployment Provider:", green(deploymentProvider))

	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Apply these settings to create squadbase.yml?",
		Default: true,
	}
	err = survey.AskOne(confirmPrompt, &confirm)
	if err != nil {
		return fmt.Errorf("initialization cancelled")
	}

	if !confirm {
		fmt.Println("Configuration cancelled by user.")
		return nil
	}

	fmt.Printf("%s Creating squadbase.yml...\n", cyan("INFO:"))
	time.Sleep(500 * time.Millisecond)

	err = project.CreateSquadbaseYml(directory, templateName, languageVersion, packageManager, deploymentProvider)
	if err != nil {
		fmt.Printf("%s Failed to create squadbase.yml: %v\n", color.RedString("ERROR:"), err)
		return err
	}

	absPath, err := filepath.Abs(directory)
	if err != nil {
		absPath = directory
	}

	fmt.Printf("\n%s Successfully created squadbase.yml in %s\n", green("✅"), absPath)
	fmt.Printf("\n%s\n", green("🎉 Project successfully initialized!"))
	fmt.Printf("Your %s project configuration is ready in %s/squadbase.yml\n",
		templateName, blue(absPath))

	return nil
}
