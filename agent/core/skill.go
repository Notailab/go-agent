package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Skill struct {
	Paths     []string `json:"paths,omitempty"`
	SkillDir  []string `json:"skill_paths,omitempty"`
	SkillDesc []string `json:"skill_desc,omitempty"`
}

func NewSkill(paths ...string) *Skill {
	skill := &Skill{
		Paths:    paths,
		SkillDir: []string{},
	}
	if err := skill.parsePath(); err != nil {
		return nil
	}
	return skill
}

func (s *Skill) parsePath() error {
	for _, p := range s.Paths {
		err := filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && info.Name() == "SKILL.md" {
				absPath, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				s.SkillDir = append(s.SkillDir, absPath)
				s.SkillDesc = append(s.SkillDesc, s.readMetaData(absPath))
			}

			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Skill) readMetaData(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	content := string(b)
	content = strings.TrimLeft(content, "\ufeff \t\r\n")
	re := regexp.MustCompile(`(?s)^---\s*(.*?)\s*---`)
	m := re.FindStringSubmatch(content)
	if len(m) >= 2 {
		return strings.Join([]string{
			"- ",
			strings.TrimSpace(m[1]),
			fmt.Sprintf("\npath: %v", path),
		}, "")
	}
	return ""
}

func (s *Skill) SystemPrompt() string {
	template := `# Skills

The following skills extend your capabilities. To use a skill, read its SKILL.md file using the read_file tool.

%v`

	desc := fmt.Sprintf(template, strings.Join(s.SkillDesc, "\n\n"))
	return desc
}

func (s *Skill) Clone() *Skill {
	if s == nil {
		return nil
	}

	clone := &Skill{}
	clone.Paths = append([]string(nil), s.Paths...)
	clone.SkillDir = append([]string(nil), s.SkillDir...)
	clone.SkillDesc = append([]string(nil), s.SkillDesc...)
	return clone
}
