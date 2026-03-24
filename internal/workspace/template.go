package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/martins6/opencode-telegram/skills"
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

	if err := copySkills(workspacePath); err != nil {
		return fmt.Errorf("failed to copy skills: %w", err)
	}

	return nil
}

func copySkills(workspacePath string) error {
	skillsDir := filepath.Join(workspacePath, ".opencode", "skills")

	return fs.WalkDir(skills.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := fs.ReadFile(skills.FS, path)
		if err != nil {
			return fmt.Errorf("failed to read skill file %s: %w", path, err)
		}

		destPath := filepath.Join(skillsDir, filepath.Base(path))
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write skill file %s: %w", path, err)
		}

		return nil
	})
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
