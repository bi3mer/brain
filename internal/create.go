package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// createNoteTypes maps a note type (as passed to `zettle create`) to the
// directory it's created in. The template for each type lives at
// templates/<type>.md.
var createNoteTypes = map[string]string{
	"literature": "literature",
	"moc":        "index",
	"person":     "index",
	"permanent":  "permanent",
	"project":    filepath.Join("projects", "future"),
}

type CreateCommand struct{}

func (CreateCommand) Name() string  { return "create" }
func (CreateCommand) Usage() string { return "create <type> <title...>" }

const createTypesMsg = "  type: literature | moc | permanent | person | project"

func (c CreateCommand) Run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: brain %s\n%s", c.Usage(), createTypesMsg)
	}
	noteType := args[0]
	title := strings.Join(args[1:], " ")

	dir, ok := createNoteTypes[noteType]
	if !ok {
		return fmt.Errorf("unknown type: %s\nusage: brain %s\n%s", noteType, c.Usage(), createTypesMsg)
	}

	dest := filepath.Join(dir, filenameFromTitle(title)+".md")
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("file already exists: %s", dest)
	}

	template := filepath.Join("templates", noteType+".md")
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
