package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateTemplate(workspacePath string) error {
	dirs := []string{
		"downloads/images",
		"downloads/audio",
		"downloads/documents",
		"downloads/videos",
		"conversations",
		".logs",
		".opencode/agents",
		".opencode/skills",
		"MAIN-PROMPTS",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(workspacePath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	files := map[string]string{
		"opencode.json":             OpenCodeConfigContent,
		"MAIN-PROMPTS/SOUL.md":      SoulContent,
		"MAIN-PROMPTS/USER.md":      UserContent,
		"MAIN-PROMPTS/IDENTITY.md":  IdentityContent,
		"MAIN-PROMPTS/BOOTSTRAP.md": BootstrapContent,
		"MAIN-PROMPTS/TOOLS.md":     ToolsContent,
		"AGENTS.md":                 AgentsContent,
	}

	for filename, content := range files {
		fullPath := filepath.Join(workspacePath, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create file %s: %w", filename, err)
			}
		}
	}

	return nil
}

func ValidateWorkspace(workspacePath string) error {
	requiredDirs := []string{
		"downloads",
		"conversations",
	}

	for _, dir := range requiredDirs {
		fullPath := filepath.Join(workspacePath, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("required directory %s does not exist", dir)
		}
	}

	return nil
}
