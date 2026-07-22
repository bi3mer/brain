package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

type RenameCommand struct{}

func (RenameCommand) Name() string  { return "rename" }
func (RenameCommand) Usage() string { return "rename <old-path> <new-path> [new-title]" }

func (c RenameCommand) Run(args []string) error {
	if len(args) != 2 && len(args) != 3 {
		return fmt.Errorf("usage: brain %s", c.Usage())
	}
	oldPath, newPath := args[0], args[1]

	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("old note not found: %w", err)
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("new path already exists: %s", newPath)
	}

	oldAbs, err := filepath.Abs(oldPath)
	if err != nil {
		return fmt.Errorf("abs: %w", err)
	}
	newAbs, err := filepath.Abs(newPath)
	if err != nil {
		return fmt.Errorf("abs: %w", err)
	}

	// Capture the real title before anything on disk moves.
	oldTitle := resolveTitle(oldPath)

	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	newTitle := titleFromFilename(newPath)
	if len(args) == 3 {
		newTitle = args[2]
	}

	if err := rewriteH1(newPath, newTitle); err != nil {
		os.Rename(newPath, oldPath)
		return fmt.Errorf("rewrite title: %w", err)
	}

	filesChanged := updateVaultLinks(oldAbs, newAbs, func(filePath, fileAbsDir, text, href string) string {
		newHref, err := filepath.Rel(fileAbsDir, newAbs)
		if err != nil {
			return "[" + text + "](" + href + ")"
		}
		if text == oldTitle {
			text = newTitle
		}
		return "[" + text + "](" + newHref + ")"
	}, func(path string) {
		fmt.Printf("Updated file: %s\n", path)
	})

	fmt.Printf("Updated %d files\n", filesChanged)
	return nil
}
