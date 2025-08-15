package config

import (
	"os"
	"path/filepath"
	"strings"
)

func detectProjectType(dir string) ProjectType {
	// Check for specific framework configs first
	if fileExists(filepath.Join(dir, "next.config.js")) ||
		fileExists(filepath.Join(dir, "next.config.ts")) {
		return ProjectTypeNext
	}

	// Check for Vite
	viteConfigs := []string{"vite.config.ts", "vite.config.js", "vite.config.mjs"}
	for _, config := range viteConfigs {
		if fileExists(filepath.Join(dir, config)) {
			return ProjectTypeVite
		}
	}

	// Check package.json for React/Node
	if fileExists(filepath.Join(dir, "package.json")) {
		if isReactProject(dir) {
			return ProjectTypeReact
		}
		return ProjectTypeNode
	}

	// Go detection
	if fileExists(filepath.Join(dir, "go.mod")) ||
		fileExists(filepath.Join(dir, "main.go")) {
		return ProjectTypeGo
	}

	// Rust detection
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return ProjectTypeRust
	}

	// Python detection
	pythonFiles := []string{"requirements.txt", "pyproject.toml", "app.py", "main.py"}
	for _, file := range pythonFiles {
		if fileExists(filepath.Join(dir, file)) {
			return ProjectTypePython
		}
	}

	// Java detection
	if fileExists(filepath.Join(dir, "pom.xml")) ||
		fileExists(filepath.Join(dir, "build.gradle")) {
		return ProjectTypeJava
	}

	// .NET detection
	dotnetExtensions := []string{".csproj", ".sln", ".fsproj"}
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, entry := range entries {
			for _, ext := range dotnetExtensions {
				if strings.HasSuffix(entry.Name(), ext) {
					return ProjectTypeDotNet
				}
			}
		}
	}

	return ""
}

func getDefaultCommand(projectType ProjectType) string {
	commands := map[ProjectType]string{
		ProjectTypeVite:   "npm run dev",
		ProjectTypeReact:  "npm start",
		ProjectTypeNext:   "npm run dev",
		ProjectTypeNode:   "npm run dev",
		ProjectTypeGo:     "go run .",
		ProjectTypeRust:   "cargo run",
		ProjectTypePython: "python app.py",
		ProjectTypeJava:   "mvn spring-boot:run",
		ProjectTypeDotNet: "dotnet run",
	}

	return commands[projectType]
}

func isReactProject(dir string) bool {
	packagePath := filepath.Join(dir, "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return false
	}

	contentStr := string(content)
	return strings.Contains(contentStr, "react") && !strings.Contains(contentStr, "next")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
