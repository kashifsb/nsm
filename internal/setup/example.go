package setup

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kashifsb/nsm/pkg/logger"
	"github.com/kashifsb/nsm/pkg/utils"
)

// go:embed templates/*
var templates embed.FS

type ExampleManager struct {
	frameworks map[string]FrameworkConfig
}

type FrameworkConfig struct {
	Name        string
	Description string
	Language    string
	Templates   []string
	PostCreate  func(projectPath string) error
	Commands    map[string]string
}

type ProjectTemplate struct {
	Name        string
	Description string
	Domain      string
	ProjectName string
	Language    string
	Framework   string
	Port        int
	HTTPSPort   int
	Author      string
	Email       string
	Year        string
}

func NewExampleManager() *ExampleManager {
	return &ExampleManager{
		frameworks: map[string]FrameworkConfig{
			"react-vite-typescript": {
				Name:        "React + Vite + TypeScript",
				Description: "Modern React application with Vite build tool and TypeScript",
				Language:    "TypeScript",
				Templates:   []string{"react-vite-ts"},
				PostCreate:  setupReactViteProject,
				Commands: map[string]string{
					"dev":     "npm run dev",
					"build":   "npm run build",
					"preview": "npm run preview",
					"lint":    "npm run lint",
				},
			},
			"go": {
				Name:        "Go Web Server",
				Description: "Simple Go web server with HTTP routing",
				Language:    "Go",
				Templates:   []string{"go-web"},
				PostCreate:  setupGoProject,
				Commands: map[string]string{
					"dev":   "go run .",
					"build": "go build -o bin/server .",
					"test":  "go test ./...",
				},
			},
			"rust": {
				Name:        "Rust Web Server",
				Description: "Rust web server using Axum framework",
				Language:    "Rust",
				Templates:   []string{"rust-axum"},
				PostCreate:  setupRustProject,
				Commands: map[string]string{
					"dev":   "cargo run",
					"build": "cargo build --release",
					"test":  "cargo test",
				},
			},
			"python": {
				Name:        "Python Flask",
				Description: "Python web application using Flask framework",
				Language:    "Python",
				Templates:   []string{"python-flask"},
				PostCreate:  setupPythonProject,
				Commands: map[string]string{
					"dev":     "python app.py",
					"install": "pip install -r requirements.txt",
					"test":    "python -m pytest",
				},
			},
			"java": {
				Name:        "Java Spring Boot",
				Description: "Java web application using Spring Boot",
				Language:    "Java",
				Templates:   []string{"java-spring"},
				PostCreate:  setupJavaProject,
				Commands: map[string]string{
					"dev":   "mvn spring-boot:run",
					"build": "mvn clean package",
					"test":  "mvn test",
				},
			},
		},
	}
}

func (em *ExampleManager) Create(framework string) error {
	config, exists := em.frameworks[framework]
	if !exists {
		return fmt.Errorf("framework '%s' not supported", framework)
	}

	logger.Info("Creating example project", "framework", framework)

	// Get project details
	projectName := em.generateProjectName(framework)
	projectPath := filepath.Join(".", projectName)

	// Check if directory already exists
	if utils.DirExists(projectPath) {
		return fmt.Errorf("directory '%s' already exists", projectPath)
	}

	// Create project directory
	if err := utils.EnsureDir(projectPath); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}

	// Prepare template data
	templateData := em.createTemplateData(projectName, framework, config)

	// Process templates
	for _, templateName := range config.Templates {
		if err := em.processTemplate(templateName, projectPath, templateData); err != nil {
			return fmt.Errorf("process template %s: %w", templateName, err)
		}
	}

	// Run post-create setup
	if config.PostCreate != nil {
		if err := config.PostCreate(projectPath); err != nil {
			return fmt.Errorf("post-create setup: %w", err)
		}
	}

	// Display success message
	em.displaySuccessMessage(projectName, projectPath, config)

	return nil
}

func (em *ExampleManager) ListFrameworks() {
	fmt.Println("📚 Available Example Frameworks:")
	fmt.Println()

	for key, config := range em.frameworks {
		fmt.Printf("  🔹 %s\n", key)
		fmt.Printf("     %s\n", config.Description)
		fmt.Printf("     Language: %s\n", config.Language)
		fmt.Println()
	}
}

func (em *ExampleManager) generateProjectName(framework string) string {
	// Generate a unique project name
	baseName := strings.ReplaceAll(framework, "-", "_")
	return fmt.Sprintf("nsm_%s_example", baseName)
}

func (em *ExampleManager) createTemplateData(projectName, framework string, config FrameworkConfig) ProjectTemplate {
	return ProjectTemplate{
		Name:        projectName,
		Description: fmt.Sprintf("NSM example project for %s", config.Name),
		Domain:      fmt.Sprintf("%s.dev", projectName),
		ProjectName: projectName,
		Language:    config.Language,
		Framework:   framework,
		Port:        5173,
		HTTPSPort:   8443,
		Author:      "NSM User",
		Email:       "user@example.com",
		Year:        "2024",
	}
}

func (em *ExampleManager) processTemplate(templateName, projectPath string, data ProjectTemplate) error {
	templateDir := fmt.Sprintf("templates/%s", templateName)

	return em.walkTemplateDir(templateDir, projectPath, data)
}

func (em *ExampleManager) walkTemplateDir(templateDir, outputDir string, data ProjectTemplate) error {
	entries, err := templates.ReadDir(templateDir)
	if err != nil {
		return fmt.Errorf("read template directory: %w", err)
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(templateDir, entry.Name())
		targetPath := filepath.Join(outputDir, entry.Name())

		if entry.IsDir() {
			// Create directory and recurse
			if err := utils.EnsureDir(targetPath); err != nil {
				return fmt.Errorf("create directory %s: %w", targetPath, err)
			}

			if err := em.walkTemplateDir(sourcePath, targetPath, data); err != nil {
				return err
			}
		} else {
			// Process file
			if err := em.processTemplateFile(sourcePath, targetPath, data); err != nil {
				return fmt.Errorf("process file %s: %w", sourcePath, err)
			}
		}
	}

	return nil
}

func (em *ExampleManager) processTemplateFile(sourcePath, targetPath string, data ProjectTemplate) error {
	// Read template content
	content, err := templates.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read template file: %w", err)
	}

	// Check if file is a template (has .tmpl extension)
	if strings.HasSuffix(sourcePath, ".tmpl") {
		// Remove .tmpl extension from target
		targetPath = strings.TrimSuffix(targetPath, ".tmpl")

		// Process as template
		tmpl, err := template.New("file").Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse template: %w", err)
		}

		// Create output file
		file, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer file.Close()

		// Execute template
		if err := tmpl.Execute(file, data); err != nil {
			return fmt.Errorf("execute template: %w", err)
		}
	} else {
		// Copy file as-is
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}

	return nil
}

func (em *ExampleManager) displaySuccessMessage(projectName, projectPath string, config FrameworkConfig) {
	fmt.Printf("🎉 Successfully created %s project!\n\n", config.Name)
	fmt.Printf("📁 Project: %s\n", projectPath)
	fmt.Printf("🌐 Domain: %s.dev\n\n", projectName)

	fmt.Println("📋 Next steps:")
	fmt.Printf("   cd %s\n", projectName)

	if config.Language == "TypeScript" || config.Language == "JavaScript" {
		fmt.Println("   npm install")
	}

	fmt.Println("   ./cmd/nsm.sh")
	fmt.Println()

	fmt.Println("🚀 Available commands:")
	for cmd, description := range config.Commands {
		fmt.Printf("   %s: %s\n", cmd, description)
	}
	fmt.Println()
}

// Post-create setup functions
func setupReactViteProject(projectPath string) error {
	logger.Info("Setting up React Vite project", "path", projectPath)

	// Create additional directories
	dirs := []string{
		filepath.Join(projectPath, "src", "components"),
		filepath.Join(projectPath, "src", "hooks"),
		filepath.Join(projectPath, "src", "utils"),
		filepath.Join(projectPath, "public"),
		filepath.Join(projectPath, "cmd"),
	}

	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return nil
}

func setupGoProject(projectPath string) error {
	logger.Info("Setting up Go project", "path", projectPath)

	// Create Go module
	// This would be done in the template, but we can add additional setup here
	dirs := []string{
		filepath.Join(projectPath, "cmd"),
		filepath.Join(projectPath, "internal"),
		filepath.Join(projectPath, "pkg"),
		filepath.Join(projectPath, "web"),
		filepath.Join(projectPath, "static"),
	}

	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return nil
}

func setupRustProject(projectPath string) error {
	logger.Info("Setting up Rust project", "path", projectPath)

	dirs := []string{
		filepath.Join(projectPath, "src", "handlers"),
		filepath.Join(projectPath, "src", "models"),
		filepath.Join(projectPath, "src", "utils"),
		filepath.Join(projectPath, "static"),
		filepath.Join(projectPath, "templates"),
		filepath.Join(projectPath, "cmd"),
	}

	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return nil
}

func setupPythonProject(projectPath string) error {
	logger.Info("Setting up Python project", "path", projectPath)

	dirs := []string{
		filepath.Join(projectPath, "app", "routes"),
		filepath.Join(projectPath, "app", "models"),
		filepath.Join(projectPath, "app", "utils"),
		filepath.Join(projectPath, "static"),
		filepath.Join(projectPath, "templates"),
		filepath.Join(projectPath, "tests"),
		filepath.Join(projectPath, "cmd"),
	}

	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return nil
}

func setupJavaProject(projectPath string) error {
	logger.Info("Setting up Java project", "path", projectPath)

	dirs := []string{
		filepath.Join(projectPath, "src", "main", "java", "com", "nsm", "example"),
		filepath.Join(projectPath, "src", "main", "resources"),
		filepath.Join(projectPath, "src", "test", "java"),
		filepath.Join(projectPath, "cmd"),
	}

	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return nil
}
