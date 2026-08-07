package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

type OutgoingLinksCommand struct{}

func (OutgoingLinksCommand) Name() string       { return "outgoing-links" }
func (OutgoingLinksCommand) Usage() string      { return "outgoing-links <path>" }
func (OutgoingLinksCommand) RequiresRoot() bool { return true }

func (c OutgoingLinksCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: brain %s", c.Usage())
	}
	path := args[0]

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read note: %w", err)
	}

	dir := filepath.Dir(path)
	count := 0
	eachLink(data, func(text, href string) {
		count++
		fmt.Println(filepath.Join(dir, href))
	})

	fmt.Printf("\n%d outgoing links from %s\n", count, path)
	return nil
}
