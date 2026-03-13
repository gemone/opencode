// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	skillsDir      = ".libcode/skills"
	skillFileName  = "SKILL.md"
	maxSkillFiles  = 10
	maxSkillSize   = 100 * 1024 // 100KB
)

// SkillInfo represents a skill's metadata and content
type SkillInfo struct {
	Name        string
	Description string
	Location    string
	Content     string
	Files       []string
}

// SkillTool loads and manages specialized skills
type SkillTool struct {
	workingDir string
	homeDir    string
}

// NewSkillTool creates a new skill tool
func NewSkillTool(workingDir, homeDir string) *SkillTool {
	return &SkillTool{
		workingDir: workingDir,
		homeDir:    homeDir,
	}
}

// ID returns the tool identifier
func (t *SkillTool) ID() string {
	return "skill"
}

// Description returns the tool description
func (t *SkillTool) Description() string {
	skills := t.discoverSkills()
	if len(skills) == 0 {
		return "Load a specialized skill that provides domain-specific instructions and workflows. No skills are currently available."
	}

	var skillList []string
	for _, skill := range skills {
		skillList = append(skillList,
			"  <skill>",
			fmt.Sprintf("    <name>%s</name>", skill.Name),
			fmt.Sprintf("    <description>%s</description>", skill.Description),
			fmt.Sprintf("    <location>file://%s</location>", skill.Location),
			"  </skill>")
	}

	return fmt.Sprintf(`Load a specialized skill that provides domain-specific instructions and workflows.

When you recognize that a task matches one of the available skills listed below, use this tool to load the full skill instructions.

The skill will inject detailed instructions, workflows, and access to bundled resources (scripts, references, templates) into the conversation context.

Tool output includes a <skill_content name="..."> block with the loaded content.

The following skills provide specialized sets of instructions for particular tasks
Invoke this tool to load a skill when a task matches one of the available skills listed below:

<available_skills>
%s
</available_skills>`, strings.Join(skillList, "\n"))
}

// Parameters returns the parameter schema
func (t *SkillTool) Parameters() *framework.Schema {
	skills := t.discoverSkills()
	hint := ""
	if len(skills) > 0 {
		examples := make([]string, 0, min(3, len(skills)))
		for i := 0; i < min(3, len(skills)); i++ {
			examples = append(examples, fmt.Sprintf("'%s'", skills[i].Name))
		}
		hint = fmt.Sprintf(" (e.g., %s)", strings.Join(examples, ", "))
	}

	return framework.NewSchema().
		AddProperty("name", framework.Property{
			Type:        "string",
			Description: fmt.Sprintf("The name of the skill from available_skills%s", hint),
			Required:    true,
		})
}

// Execute loads and returns a skill
func (t *SkillTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	name, _ := params["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("skill name is required")
	}

	// Discover all skills
	skills := t.discoverSkills()

	// Find the requested skill
	var skill *SkillInfo
	for i := range skills {
		if skills[i].Name == name {
			skill = &skills[i]
			break
		}
	}

	if skill == nil {
		available := make([]string, 0, len(skills))
		for _, s := range skills {
			available = append(available, s.Name)
		}
		return nil, fmt.Errorf("skill '%s' not found. Available skills: %s", name, strings.Join(available, ", "))
	}

	// Ask for permission
	if execCtx.PermissionAsker != nil {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "skill",
			Patterns:   []string{name},
			Always:     []string{name},
			Metadata:   nil,
		})
		if err != nil {
			return nil, err
		}
	}

	// Load skill content
	content, err := t.loadSkillContent(skill.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to load skill content: %w", err)
	}

	// Discover skill files
	skillDir := filepath.Dir(skill.Location)
	files := t.discoverSkillFiles(skillDir, maxSkillFiles)

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("<skill_content name=\"%s\">\n", skill.Name))
	output.WriteString(fmt.Sprintf("# Skill: %s\n\n", skill.Name))
	output.WriteString(content)
	output.WriteString("\n")
	output.WriteString(fmt.Sprintf("Base directory for this skill: file://%s\n", skillDir))
	output.WriteString("Relative paths in this skill (e.g., scripts/, reference/) are relative to this base directory.\n")
	output.WriteString("Note: file list is sampled.\n\n")
	output.WriteString("<skill_files>\n")
	for _, file := range files {
		output.WriteString(fmt.Sprintf("<file>%s</file>\n", file))
	}
	output.WriteString("</skill_files>\n")
	output.WriteString("</skill_content>")

	return &framework.Result{
		Title: fmt.Sprintf("Loaded skill: %s", skill.Name),
		Output: output.String(),
		Metadata: map[string]any{
			"name": skill.Name,
			"dir":  skillDir,
		},
	}, nil
}

// discoverSkills finds all available skills
func (t *SkillTool) discoverSkills() []SkillInfo {
	var skills []SkillInfo

	// Check home directory skills
	homeSkills := filepath.Join(t.homeDir, skillsDir)
	if skillDirs, err := os.ReadDir(homeSkills); err == nil {
		for _, skillDir := range skillDirs {
			if !skillDir.IsDir() {
				continue
			}

			skillPath := filepath.Join(homeSkills, skillDir.Name(), skillFileName)
			if info, err := t.parseSkillInfo(skillPath); err == nil {
				skills = append(skills, info)
			}
		}
	}

	// Check project directory skills
	projectSkills := filepath.Join(t.workingDir, skillsDir)
	if skillDirs, err := os.ReadDir(projectSkills); err == nil {
		for _, skillDir := range skillDirs {
			if !skillDir.IsDir() {
				continue
			}

			skillPath := filepath.Join(projectSkills, skillDir.Name(), skillFileName)
			if info, err := t.parseSkillInfo(skillPath); err == nil {
				skills = append(skills, info)
			}
		}
	}

	return skills
}

// parseSkillInfo parses a SKILL.md file and extracts metadata
func (t *SkillTool) parseSkillInfo(skillPath string) (SkillInfo, error) {
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return SkillInfo{}, err
	}

	content := string(data)

	// Extract name from frontmatter or first heading
	name := ""
	description := ""

	lines := strings.Split(content, "\n")
	inFrontmatter := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for frontmatter
		if trimmed == "---" {
			if i == 0 {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				inFrontmatter = false
				continue
			}
		}

		if inFrontmatter {
			if strings.HasPrefix(trimmed, "name:") {
				name = strings.TrimSpace(strings.TrimPrefix(trimmed, "name:"))
				// Remove quotes if present
				name = strings.Trim(name, `"'`)
			} else if strings.HasPrefix(trimmed, "description:") {
				description = strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
				// Remove quotes if present
				description = strings.Trim(description, `"'`)
			}
		} else if name == "" && strings.HasPrefix(trimmed, "# ") {
			// Use first heading as name if not found in frontmatter
			name = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}

	// Use directory name as fallback
	if name == "" {
		name = filepath.Base(filepath.Dir(skillPath))
	}

	return SkillInfo{
		Name:        name,
		Description: description,
		Location:    skillPath,
		Content:     content,
	}, nil
}

// loadSkillContent reads the skill content from disk
func (t *SkillTool) loadSkillContent(skillPath string) (string, error) {
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return "", err
	}

	if len(data) > maxSkillSize {
		return string(data[:maxSkillSize]) + "\n\n... (content truncated)",
			nil
	}

	return string(data), nil
}

// discoverSkillFiles finds files in a skill directory
func (t *SkillTool) discoverSkillFiles(skillDir string, limit int) []string {
	var files []string

	err := filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and the SKILL.md file
		if info.IsDir() {
			return nil
		}

		if filepath.Base(path) == skillFileName {
			return nil
		}

		// Convert to relative path
		relPath, err := filepath.Rel(skillDir, path)
		if err != nil {
			return nil
		}

		files = append(files, filepath.Join(skillDir, relPath))

		if len(files) >= limit {
			return fmt.Errorf("limit reached")
		}

		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return files
	}

	return files
}
