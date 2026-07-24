package commands

import (
	"fmt"
	"os"
)

type DndCharacterCommand struct{}

func (DndCharacterCommand) Name() string { return "dnd-char" }

func (DndCharacterCommand) Usage() string {
	return "dnd-char <type> <First Name> <Last Name>"
}

func (DndCharacterCommand) RequiresRoot() bool { return false }

func (c DndCharacterCommand) Run(args []string) error {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: brain", c.Usage())
		os.Exit(1)
	}

	return nil
}
