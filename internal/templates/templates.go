package templates

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Template struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

type TemplateList []Template

type TemplateJSON struct {
	Templates TemplateList `json:"templates"`
}

const (
	GitHubRepoURL    = "https://github.com/squadbase/squadbase-template"
	GitHubRepoBranch = "main"
)

var (
	templateCache = TemplateList{}
)

func GetAvailableTemplates(forceRefresh bool) (TemplateList, error) {
	if !forceRefresh && templateCache != nil && len(templateCache) > 0 {
		return templateCache, nil
	}

	tempDir, err := os.MkdirTemp("", "squadbase_template_")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	zipPath := filepath.Join(tempDir, "repo.zip")
	downloadURL := fmt.Sprintf("%s/archive/refs/heads/%s.zip", GitHubRepoURL, GitHubRepoBranch)

	err = downloadFile(downloadURL, zipPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to download template repository: %w", err)
	}

	err = unzip(zipPath, tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to extract zip file: %w", err)
	}

	repoDir := ""
	files, err := os.ReadDir(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to read temporary directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() && file.Name() != "__MACOSX" {
			repoDir = filepath.Join(tempDir, file.Name())
			break
		}
	}

	if repoDir == "" {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("no directories found in the downloaded zip")
	}

	var templates TemplateList

	templateJSONPath := filepath.Join(repoDir, "template.json")
	if _, err := os.Stat(templateJSONPath); err == nil {
		data, err := os.ReadFile(templateJSONPath)
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("failed to read template.json: %w", err)
		}

		var templateData TemplateJSON
		err = json.Unmarshal(data, &templateData)
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("failed to parse template.json: %w", err)
		}
		templates = templateData.Templates
	} else {
		files, err := os.ReadDir(repoDir)
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("failed to read repository directory: %w", err)
		}

		for _, file := range files {
			if file.IsDir() && !strings.HasPrefix(file.Name(), ".") && file.Name() != "__pycache__" {
				templates = append(templates, Template{
					Name:        file.Name(),
					Description: fmt.Sprintf("A %s project template", file.Name()),
					Path:        file.Name(),
				})
			}
		}
	}

	var validTemplates TemplateList
	for _, templateInfo := range templates {
		templatePath := templateInfo.Path
		if templatePath == "" {
			templatePath = templateInfo.Name
		}

		templateDir := filepath.Join(repoDir, templatePath)
		if strings.HasPrefix(templatePath, "./") {
			trimmedPath := strings.TrimPrefix(templatePath, "./")
			templateDir = filepath.Join(repoDir, trimmedPath)
		}

		if info, err := os.Stat(templateDir); err == nil && info.IsDir() {
			validTemplates = append(validTemplates, templateInfo)
		}
	}

	templateCache = validTemplates

	return validTemplates, nil
}

func ListTemplateFiles(templateName string) ([]string, error) {
	tempDir, err := os.MkdirTemp("", "squadbase_template_")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	templatePath, err := downloadTemplateFromGitHub(templateName, tempDir)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(templatePath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list template files: %w", err)
	}

	return files, nil
}

func CopyTemplateFiles(templateName string, destination string) error {
	tempDir, err := os.MkdirTemp("", "squadbase_template_")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	templatePath, err := downloadTemplateFromGitHub(templateName, tempDir)
	if err != nil {
		return err
	}

	return filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		destPath := filepath.Join(destination, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		err = copyFile(path, destPath)
		if err != nil {
			return fmt.Errorf("failed to copy file %s: %w", relPath, err)
		}

		return nil
	})
}

var templateRepoDir string

func downloadTemplateFromGitHub(templateName string, tempDir string) (string, error) {
	if templateRepoDir != "" {
		templatePath := filepath.Join(templateRepoDir, templateName)
		if _, err := os.Stat(templatePath); err == nil {
			return templatePath, nil
		}
	}

	zipPath := filepath.Join(tempDir, "repo.zip")
	downloadURL := fmt.Sprintf("%s/archive/refs/heads/%s.zip", GitHubRepoURL, GitHubRepoBranch)

	err := downloadFile(downloadURL, zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to download template repository: %w", err)
	}

	err = unzip(zipPath, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to extract zip file: %w", err)
	}

	repoDir := ""
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to read temporary directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() && file.Name() != "__MACOSX" {
			repoDir = filepath.Join(tempDir, file.Name())
			break
		}
	}

	if repoDir == "" {
		return "", fmt.Errorf("no directories found in the downloaded zip")
	}

	templateRepoDir = repoDir

	templateDir := filepath.Join(repoDir, templateName)
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		availableTemplates, err := GetAvailableTemplates(false)
		if err != nil {
			return "", fmt.Errorf("failed to get available templates: %w", err)
		}

		var templatePath string
		for _, tmpl := range availableTemplates {
			if tmpl.Name == templateName {
				if tmpl.Path != "" {
					templatePath = tmpl.Path
				} else {
					templatePath = templateName
				}
				break
			}
		}

		if templatePath == "" {
			templatePath = templateName
		}

		templateDir = filepath.Join(repoDir, templatePath)
		if strings.HasPrefix(templatePath, "./") {
			trimmedPath := strings.TrimPrefix(templatePath, "./")
			templateDir = filepath.Join(repoDir, trimmedPath)
		}

		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			return "", fmt.Errorf("template '%s' not found in GitHub repository at path %s", templateName, templateDir)
		}
	}

	return templateDir, nil
}

func downloadFile(url string, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}
