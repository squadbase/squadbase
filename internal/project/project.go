package project

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/squadbase/squadbase/internal/templates"
	"github.com/squadbase/squadbase/internal/ui"
)

type Config struct {
	Language           string
	Version            string
	PackageManager     string
	AuthorName         string
	AuthorEmail        string
	DeploymentProvider string
}

func CreateProject(projectName string, templateName string, config *Config) error {
	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	availableTemplates, err := templates.GetAvailableTemplates(false)
	if err != nil {
		return fmt.Errorf("failed to get available templates: %w", err)
	}

	templateExists := false
	for _, tmpl := range availableTemplates {
		if tmpl.Name == templateName {
			templateExists = true
			break
		}
	}

	if !templateExists {
		return fmt.Errorf("template %s not found", templateName)
	}

	err = templates.CopyTemplateFiles(templateName, projectName)
	if err != nil {
		return fmt.Errorf("failed to copy template files: %w", err)
	}

	if config != nil {
		switch templateName {
		case "morph":
			switch config.PackageManager {
			case "poetry":
				err = createPoetryPyprojectToml(projectName, templateName, config.AuthorName, config.AuthorEmail)
			case "uv":
				err = createUvPyprojectToml(projectName, templateName, config.AuthorName, config.AuthorEmail)
			case "pip":
				err = createRequirementsTxt(projectName, templateName)
			}

			if !isCommandAvailable("npm") {
				currentDir, _ := os.Getwd()
				err = os.Chdir(projectName)
				if err == nil {
					npmInstall := exec.Command("npm", "install")
					_ = npmInstall.Run()

					shadcnCmd := exec.Command("npx", "shadcn@latest", "add", "--yes", "https://morph-components.vercel.app/r/morph-components.json")
					_ = shadcnCmd.Run()

					_ = os.Chdir(currentDir)
				}
			} else {
				ui.PrintWarningBox("🚧 Warning", "npm is not available.\nAfter installing npm, run following commands in your project directory.:\n- `npm install`.\n- `npx shadcn@latest add --yes https://morph-components.vercel.app/r/morph-components.json`")
			}
		case "streamlit":
			switch config.PackageManager {
			case "poetry":
				err = createPoetryPyprojectToml(projectName, templateName, config.AuthorName, config.AuthorEmail)
			case "uv":
				err = createUvPyprojectToml(projectName, templateName, config.AuthorName, config.AuthorEmail)
			case "pip":
				err = createRequirementsTxt(projectName, templateName)
			}
		case "nextjs":
			err = updatePackageJson(projectName, config.Version, config.AuthorName, config.AuthorEmail)
		}

		if err != nil {
			return fmt.Errorf("failed to customize project: %w", err)
		}
	}

	err = CreateSquadbaseYml(projectName, templateName, config.Version, config.PackageManager, config.DeploymentProvider)
	if err != nil {
		return fmt.Errorf("failed to create squadbase.yml: %w", err)
	}

	return nil
}

func CreateSquadbaseYml(
	projectPath string,
	templateName string,
	languageVersion string,
	packageManager string,
	deploymentProvider string,
) error {
	language := "python"
	if templateName == "nextjs" {
		language = "nodejs"
	}

	comment := ""
	if language == "python" {
		comment = " # Supported: python3.9, python3.10, python3.11, python3.12"
	}

	packageManagerComment := ""
	if packageManager != "" {
		if language == "python" {
			packageManagerComment = " # Supported: poetry, uv, pip"
		}
	}

	content := fmt.Sprintf(`version: '1'
# Build Settings
build:
    # These settings are required when use_custom_dockerfile is false
    # They define the environment in which the project will be built
    runtime: %s%s%s
    framework: %s
    package_manager: %s%s
    # entrypoint: .
    # context: .
    # build_args:
    #   - ARG_NAME=value
    #   - ANOTHER_ARG=value

# Deployment Settings
deployment:
    provider: %s
    # These settings are used only when you want to customize the deployment settings
    # aws:
    #     region: ap-northeast-1
    #     memory: 1024
    #     timeout: 30
    #     provisioned_concurrency: 0
    #     ephemeral_storage: 512MB
    # gcp:
    #     region: us-central1
    #     memory: 1024
    #     cpu: 1
    #     concurrency: 80
    #     timeout: 60
    #     min_instances: 0
    #     ephemeral_storage: 100Mi
`,
		language, languageVersion, comment,
		templateName,
		packageManager, packageManagerComment,
		deploymentProvider,
	)

	filePath := filepath.Join(projectPath, "squadbase.yml")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func InitializeGit(projectPath string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	err = os.Chdir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}

	gitInit := exec.Command("git", "init")
	err = gitInit.Run()
	if err != nil {
		os.Chdir(currentDir)
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	gitBranch := exec.Command("git", "branch", "-M", "main")
	err = gitBranch.Run()
	if err != nil {
		os.Chdir(currentDir)
		return fmt.Errorf("failed to rename main branch: %w", err)
	}

	gitAdd := exec.Command("git", "add", ".")
	err = gitAdd.Run()
	if err != nil {
		os.Chdir(currentDir)
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	gitCommit := exec.Command("git", "commit", "-m", "Initial commit from Squadbase CLI")
	err = gitCommit.Run()
	if err != nil {
		os.Chdir(currentDir)
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	err = os.Chdir(currentDir)
	if err != nil {
		return fmt.Errorf("failed to change back to original directory: %w", err)
	}

	return nil
}

func GetGitUserInfo() (string, string) {
	defaultName := "Your Name"
	defaultEmail := "your.email@example.com"

	nameCmd := exec.Command("git", "config", "--get", "user.name")
	nameOutput, err := nameCmd.Output()
	name := defaultName
	if err == nil && len(nameOutput) > 0 {
		name = strings.TrimSpace(string(nameOutput))
	}

	emailCmd := exec.Command("git", "config", "--get", "user.email")
	emailOutput, err := emailCmd.Output()
	email := defaultEmail
	if err == nil && len(emailOutput) > 0 {
		email = strings.TrimSpace(string(emailOutput))
	}

	return name, email
}

func GetCurrentPythonVersion() string {
	pythonCmd := exec.Command("python3", "--version")
	output, err := pythonCmd.Output()

	if err != nil {
		pythonCmd = exec.Command("python", "--version")
		output, err = pythonCmd.Output()
		if err != nil {
			return "3.10" // default fallback
		}
	}

	versionStr := string(output)
	re := regexp.MustCompile(`Python (\d+\.\d+)\.\d+`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) > 1 {
		return matches[1]
	}

	return "3.10" // default fallback
}

func GetCurrentNodeVersion() string {
	nodeCmd := exec.Command("node", "--version")
	output, err := nodeCmd.Output()

	if err != nil {
		return "18" // default fallback
	}

	versionStr := string(output)
	re := regexp.MustCompile(`v(\d+)\.`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) > 1 {
		return matches[1]
	}

	return "18"
}

func createPoetryPyprojectToml(projectPath, templateName, authorName, authorEmail string) error {
	var dependencies []string
	var packages string

	if templateName == "morph" {
		morphPackageVersion := getLatestPackageVersion("morph-data")
		morphPackage := "morph-data"
		if morphPackageVersion != "" {
			morphPackage = fmt.Sprintf("morph-data>=%s", morphPackageVersion)
		}
		dependencies = []string{morphPackage}
		packages = `{ include = "src"}`
	} else if templateName == "streamlit" {
		dependencies = []string{"streamlit>=1.20.0"}
		packages = `{ include = "app"}`
	}

	content := fmt.Sprintf(`[tool.poetry]
name = "%s"
version = "0.1.0"
description = "A %s project"
authors = ["%s <%s>"]
packages = [%s]

[tool.poetry.dependencies]
python = ">=3.9,<3.13"
`, filepath.Base(projectPath), templateName, authorName, authorEmail, packages)

	for _, dep := range dependencies {
		parts := strings.SplitN(dep, ">=", 2)
		pkgName := parts[0]
		if len(parts) > 1 {
			content += fmt.Sprintf("%s = \">=%s\"\n", pkgName, parts[1])
		} else {
			content += fmt.Sprintf("%s = \"*\"\n", pkgName)
		}
	}

	content += `
[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"
`

	filePath := filepath.Join(projectPath, "pyproject.toml")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func createUvPyprojectToml(projectPath, templateName, authorName, authorEmail string) error {
	var dependencies []string
	var packages string

	if templateName == "morph" {
		morphPackageVersion := getLatestPackageVersion("morph-data")
		morphPackage := "morph-data"
		if morphPackageVersion != "" {
			morphPackage = fmt.Sprintf("morph-data>=%s", morphPackageVersion)
		}
		dependencies = []string{morphPackage}
		packages = `["src"]`
	} else if templateName == "streamlit" {
		dependencies = []string{"streamlit>=1.20.0"}
		packages = `["app"]`
	}

	content := fmt.Sprintf(`[project]
name = "%s"
version = "0.1.0"
description = "A %s project"
authors = [{ name = "%s", email = "%s" }]
requires-python = ">=3.9,<3.13"

dependencies = [
`, filepath.Base(projectPath), templateName, authorName, authorEmail)

	for _, dep := range dependencies {
		parts := strings.SplitN(dep, ">=", 2)
		pkgName := parts[0]
		if len(parts) > 1 {
			content += fmt.Sprintf("    \"%s>=%s\",\n", pkgName, parts[1])
		} else {
			content += fmt.Sprintf("    \"%s\",\n", pkgName)
		}
	}
	content += "]\n"

	content += fmt.Sprintf(`
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.hatch.build.targets.wheel]
packages = %s

[tool.hatch.metadata]
allow-direct-references = true
`, packages)

	filePath := filepath.Join(projectPath, "pyproject.toml")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func createRequirementsTxt(projectPath, templateName string) error {
	var dependencies []string

	if templateName == "morph" {
		morphPackageVersion := getLatestPackageVersion("morph-data")
		morphPackage := "morph-data"
		if morphPackageVersion != "" {
			morphPackage = fmt.Sprintf("morph-data>=%s", morphPackageVersion)
		}
		dependencies = []string{morphPackage}
	} else if templateName == "streamlit" {
		dependencies = []string{"streamlit>=1.20.0"}
	}

	content := "# Requirements for the project\n\n"
	for _, dep := range dependencies {
		content += fmt.Sprintf("%s\n", dep)
	}

	filePath := filepath.Join(projectPath, "requirements.txt")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func updatePackageJson(projectPath, nodeVersion, authorName, authorEmail string) error {
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("%s Warning: Could not find package.json in %s\n", yellow("WARNING:"), projectPath)
		return nil
	}

	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	var packageJSON map[string]any
	err = json.Unmarshal(data, &packageJSON)
	if err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	packageJSON["name"] = filepath.Base(projectPath)

	engines := make(map[string]string)
	engines["node"] = fmt.Sprintf(">=%s.0.0", nodeVersion)
	packageJSON["engines"] = engines

	packageJSON["author"] = fmt.Sprintf("%s <%s>", authorName, authorEmail)

	updatedJSON, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated package.json: %w", err)
	}

	return os.WriteFile(packageJSONPath, updatedJSON, 0644)
}

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func getLatestPackageVersion(packageName string) string {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var result struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return ""
	}

	return result.Info.Version
}
