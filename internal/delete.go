package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

type DeleteCommand struct{}

func (DeleteCommand) Name() string  { return "delete" }
func (DeleteCommand) Usage() string { return "delete <path>" }

func (c DeleteCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: brain %s", c.Usage())
	}
	path := args[0]

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("note not found: %w", err)
	}

	targetAbs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("abs: %w", err)
	}

	// Capture the title and tear down every inbound link before touching
	// the file itself, so a failure partway through never destroys the
	// note while leaving dangling tombstones elsewhere.
	title := resolveTitle(path)
	tombstone := "[[" + title + " deleted]]"

	linksRemoved := 0
	filesChanged := updateVaultLinks(targetAbs, targetAbs, func(filePath, fileAbsDir, text, href string) string {
		linksRemoved++
		fmt.Printf("Removed link in %s: [%s](%s)\n", filePath, text, href)
		return tombstone
	}, nil)

	fmt.Printf("Removed %d links across %d files\n", linksRemoved, filesChanged)

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	fmt.Printf("Deleted %s\n", path)
	return nil
}
