package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

type IncomingLinksCommand struct{}

func (IncomingLinksCommand) Name() string       { return "incoming-links" }
func (IncomingLinksCommand) Usage() string      { return "incoming-links <path>" }
func (IncomingLinksCommand) RequiresRoot() bool { return true }

func (c IncomingLinksCommand) Run(args []string) error {
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

	count := 0
	walkBrainFiles(func(p string) {
		absP, err := filepath.Abs(p)
		if err != nil || absP == targetAbs {
			return
		}

		data, err := os.ReadFile(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read", p, ":", err)
			return
		}

		dir := filepath.Dir(p)
		eachLink(data, func(text, href string) {
			resolved, err := filepath.Abs(filepath.Join(dir, href))
			if err != nil || resolved != targetAbs {
				return
			}
			count++
			fmt.Println(p)
		})
	})

	fmt.Printf("\n%d incoming links to %s\n", count, path)
	return nil
}
