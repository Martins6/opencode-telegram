package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/martins6/acolyte/skills"
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

type StrictValidator struct {
	RequiredDirs  []string
	RequiredFiles []string
}

func DefaultStrictValidator() StrictValidator {
	return StrictValidator{
		RequiredDirs: []string{
			"downloads/images",
			"downloads/audio",
			"downloads/documents",
			"downloads/videos",
			"conversations",
			".logs",
			".opencode/agents",
			".opencode/skills",
			"MAIN-PROMPTS",
		},
		RequiredFiles: []string{
			"opencode.json",
			"AGENTS.md",
			"MAIN-PROMPTS/SOUL.md",
			"MAIN-PROMPTS/USER.md",
			"MAIN-PROMPTS/IDENTITY.md",
			"MAIN-PROMPTS/BOOTSTRAP.md",
			"MAIN-PROMPTS/TOOLS.md",
		},
	}
}

type StrictValidationError struct {
	Workspace    string
	MissingDirs  []string
	MissingFiles []string
}

func (e *StrictValidationError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "workspace %q is not a valid Acolyte workspace:", e.Workspace)
	if len(e.MissingDirs) > 0 {
		b.WriteString("\n  missing directories:")
		for _, d := range e.MissingDirs {
			fmt.Fprintf(&b, "\n    - %s", d)
		}
	}
	if len(e.MissingFiles) > 0 {
		b.WriteString("\n  missing files:")
		for _, f := range e.MissingFiles {
			fmt.Fprintf(&b, "\n    - %s", f)
		}
	}
	fmt.Fprintf(&b, "\n\nfix: run `acolyte new %s` to bootstrap the missing files", e.Workspace)
	return b.String()
}

func StrictValidate(workspacePath string) error {
	return StrictValidateWith(workspacePath, DefaultStrictValidator())
}

func StrictValidateWith(workspacePath string, validator StrictValidator) error {
	info, err := os.Stat(workspacePath)
	if err != nil {
		return fmt.Errorf("workspace directory %q is not accessible: %w", workspacePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("workspace %q exists but is not a directory", workspacePath)
	}

	var verr StrictValidationError
	verr.Workspace = workspacePath

	for _, dir := range validator.RequiredDirs {
		full := filepath.Join(workspacePath, dir)
		info, err := os.Stat(full)
		if err != nil || !info.IsDir() {
			verr.MissingDirs = append(verr.MissingDirs, dir)
		}
	}

	for _, file := range validator.RequiredFiles {
		full := filepath.Join(workspacePath, file)
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			verr.MissingFiles = append(verr.MissingFiles, file)
		}
	}

	if len(verr.MissingDirs) > 0 || len(verr.MissingFiles) > 0 {
		return &verr
	}
	return nil
}
