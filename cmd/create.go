package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"slices"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/squadbase/squadbase/internal/project"
	"github.com/squadbase/squadbase/internal/templates"
	"github.com/squadbase/squadbase/internal/ui"
	"github.com/urfave/cli/v2"
)

func CreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Aliases:   []string{"c"},
		Usage:     "Create a new project from a template",
		ArgsUsage: "[PROJECT_NAME]",
		Action:    createAction,
	}
}

func createAction(c *cli.Context) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()

	ui.PrintStep(1, 6, "Project Setup")

	projectName := c.Args().First()
	if projectName == "" {
		namePrompt := &survey.Input{
			Message: "What do you want to name your project? 📝",
		}
		err := survey.AskOne(namePrompt, &projectName)
		if err != nil {
			ui.PrintError("Project creation cancelled")
			return fmt.Errorf("project creation cancelled")
		}
	}

	absPath, err := filepath.Abs(projectName)
	if err != nil {
		absPath = projectName
	}
	_, err = os.Stat(absPath)
	if !os.IsNotExist(err) {
		ui.PrintError(fmt.Sprintf("Directory %s already exists. Please choose a different name.", absPath))
		return fmt.Errorf("directory already exists")
	}

	ui.PrintInfo("Fetching available templates...")
	spinner := ui.ShowSpinner("Loading templates")

	availableTemplates, err := templates.GetAvailableTemplates(false)
	if err != nil {
		spinner.Fail("Failed to load templates")
		ui.PrintError(fmt.Sprintf("Failed to get available templates: %v", err))
		return err
	} else if len(availableTemplates) == 0 {
		spinner.Fail("No templates found")
		ui.PrintError("No templates available for project creation.")
		return fmt.Errorf("no templates available")
	}
	spinner.Success("Templates loaded successfully")

	templateNames := make([]string, 0, len(availableTemplates))
	for _, template := range availableTemplates {
		templateNames = append(templateNames, template.Name)
	}

	ui.PrintStep(2, 6, "Template Selection")

	fmt.Println()
	title := "Select a framework for your project: 🧩"
	instructions := "  (Use ↑/↓ arrows to navigate, Enter to select)"

	fmt.Println(ui.GetPrimaryText(title))
	fmt.Println(ui.GetSecondaryText(instructions))

	var templateName string
	templatePrompt := &survey.Select{
		Message: "",
		Options: templateNames,
	}
	err = survey.AskOne(templatePrompt, &templateName)
	if err != nil {
		ui.PrintError("Template selection cancelled")
		return fmt.Errorf("project creation cancelled")
	}

	var selectedTemplate templates.Template
	for _, template := range availableTemplates {
		if template.Name == templateName {
			selectedTemplate = template
			break
		}
	}

	templateInfo := make(map[string]string)
	templateInfo["Template"] = templateName
	templateInfo["Description"] = selectedTemplate.Description

	templateFiles, err := templates.ListTemplateFiles(templateName)
	if err == nil && len(templateFiles) > 0 {
		count := min(len(templateFiles), 5)

		filesList := ""
		for i := 0; i < count; i++ {
			if i == count-1 && len(templateFiles) > count {
				filesList += "..."
			} else {
				filesList += templateFiles[i]
				if i < count-1 {
					filesList += ", "
				}
			}
		}

		templateInfo["Key Files"] = filesList
	}

	ui.PrintSummaryBox("✨ Template Information", templateInfo)

	authorName, authorEmail := project.GetGitUserInfo()

	config := &project.Config{
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
	}

	if templateName == "squadbase" || templateName == "streamlit" {
		ui.PrintStep(3, 6, "Python Configuration")
		fmt.Println(ui.GetAccentText("\n🐍 Python Configuration"))

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
		var pythonVersion string
		err = survey.AskOne(versionPrompt, &pythonVersion)
		if err != nil {
			ui.PrintError("Python configuration cancelled")
			return fmt.Errorf("project creation cancelled")
		}

		pmPrompt := &survey.Select{
			Message: "Select package manager:",
			Options: []string{"poetry", "uv", "pip"},
			Default: "poetry",
		}
		var packageManager string
		err = survey.AskOne(pmPrompt, &packageManager)
		if err != nil {
			ui.PrintError("Package manager selection cancelled")
			return fmt.Errorf("project creation cancelled")
		}

		config.Language = "python"
		config.Version = pythonVersion
		config.PackageManager = packageManager

		fmt.Printf("\n%s\n", cyan("Configuration Summary:"))
		fmt.Printf("  %-20s %s\n", "Python Version:", green(pythonVersion))
		fmt.Printf("  %-20s %s\n", "Package Manager:", green(packageManager))
		fmt.Printf("  %-20s %s\n", "Author:", green(fmt.Sprintf("%s <%s>", authorName, authorEmail)))

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
		var nodeVersion string
		err = survey.AskOne(versionPrompt, &nodeVersion)
		if err != nil {
			return fmt.Errorf("project creation cancelled")
		}

		pmPrompt := &survey.Select{
			Message: "Select package manager:",
			Options: []string{"npm", "yarn", "pnpm"},
			Default: "npm",
		}
		var jsPackageManager string
		err = survey.AskOne(pmPrompt, &jsPackageManager)
		if err != nil {
			return fmt.Errorf("project creation cancelled")
		}

		config.Language = "nodejs"
		config.Version = nodeVersion
		config.PackageManager = jsPackageManager

		fmt.Printf("\n%s\n", cyan("Configuration Summary:"))
		fmt.Printf("  %-20s %s\n", "Node.js Version:", green(nodeVersion))
		fmt.Printf("  %-20s %s\n", "Package Manager:", green(jsPackageManager))
		fmt.Printf("  %-20s %s\n", "Author:", green(fmt.Sprintf("%s <%s>", authorName, authorEmail)))
	}

	ui.PrintStep(4, 5, "Deployment Configuration")
	fmt.Println(ui.GetAccentText("\n🚀 Deployment Configuration"))

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
	var deploymentProvider string
	err = survey.AskOne(dpPrompt, &deploymentProvider)
	if err != nil {
		ui.PrintError("Deployment provider selection cancelled")
		return fmt.Errorf("project creation cancelled")
	}

	config.DeploymentProvider = deploymentProvider

	ui.PrintStep(5, 6, "Project Creation")
	ui.PrintInfo(fmt.Sprintf("Creating %s project...", templateName))

	spinner = ui.ShowSpinner("Downloading template and creating project structure")

	err = project.CreateProject(projectName, templateName, config)
	if err != nil {
		spinner.Fail("Project creation failed")
		ui.PrintError(fmt.Sprintf("Failed to create project: %v", err))
		return err
	}
	spinner.Success("Project structure created")

	ui.PrintSuccess(fmt.Sprintf("Successfully created \"%s\" project in %s", templateName, absPath))

	ui.PrintStep(6, 6, "Version Control")
	fmt.Println(ui.GetAccentText("\n📦 Git Setup"))

	var useGit bool
	gitPrompt := &survey.Confirm{
		Message: "Do you want to use git for version control?",
		Default: true,
	}
	err = survey.AskOne(gitPrompt, &useGit)
	if err != nil {
		ui.PrintWarning("Project created but git initialization was cancelled")
		return fmt.Errorf("project created but git initialization was cancelled")
	}

	if useGit {
		ui.PrintInfo("Initializing git repository...")

		spinner := ui.ShowSpinner("Setting up git repository")

		err = project.InitializeGit(absPath)
		if err != nil {
			spinner.Fail("Git initialization failed")
			ui.PrintWarning(fmt.Sprintf("Failed to initialize git: %v", err))
			ui.PrintInfo("Project created but git initialization failed.")
		} else {
			spinner.Success("Git repository initialized")
			ui.PrintSuccess("Created initial commit")
		}
	}

	successBox := make(map[string]string)
	successBox["Project"] = templateName
	successBox["Location"] = absPath
	if config.Language == "python" {
		successBox["Python Version"] = config.Version
		successBox["Package Manager"] = config.PackageManager
	} else if config.Language == "nodejs" {
		successBox["Node.js Version"] = config.Version
		successBox["Package Manager"] = config.PackageManager
	}
	successBox["Deployment Provider"] = config.DeploymentProvider
	successBox["Git Initialized"] = fmt.Sprintf("%v", useGit)

	fmt.Println()
	ui.PrintSummaryBox("🎉 Project Successfully Configured!", successBox)
	fmt.Println(ui.GetPrimaryText("\nYour project is ready to go! 🚀"))

	return nil
}
