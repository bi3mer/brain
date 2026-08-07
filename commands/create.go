package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type noteType struct {
	dir      string
	template string
}

var createNoteTypes = map[string]noteType{
	"literature": {"literature", "literature.md"},
	"index":      {"index", "index.md"},
	"reference":  {"reference", "reference.md"},
	"permanent":  {"permanent", "permanent.md"},
	"stub":       {"stubs", "permanent.md"},
	"project":    {filepath.Join("projects", "future"), "project.md"},
}

type CreateCommand struct{}

func (CreateCommand) Name() string       { return "create" }
func (CreateCommand) Usage() string      { return "create <type> <title...>" }
func (CreateCommand) RequiresRoot() bool { return true }

const createTypesMsg = "  type: literature | index | stub | permanent | reference | project"

func (c CreateCommand) Run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: brain %s\n%s", c.Usage(), createTypesMsg)
	}
	typeName := args[0]
	title := strings.Join(args[1:], " ")

	nt, ok := createNoteTypes[typeName]
	if !ok {
		return fmt.Errorf("unknown type: %s\nusage: brain %s\n%s", typeName, c.Usage(), createTypesMsg)
	}

	dest := filepath.Join(nt.dir, filenameFromTitle(title)+".md")
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("file already exists: %s", dest)
	}

	template := filepath.Join(TemplateDirectory, nt.template)
	data, err := os.ReadFile(template)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	content := strings.ReplaceAll(string(data), "REPLACE", title)
	if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	fmt.Printf("Created: %s\n", dest)
	return nil
}
