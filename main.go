package main

import (
	"fmt"
	"os"
	"path/filepath"

	"brain/commands"
)

// registry holds every registered subcommand, in registration order.
var registry []commands.Command

func register(c commands.Command) {
	registry = append(registry, c)
}

func lookup(name string) commands.Command {
	for _, c := range registry {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func init() {
	register(commands.InitCommand{})
	register(commands.RenameCommand{})
	register(commands.DeleteCommand{})
	register(commands.CreateCommand{})
	register(commands.DndCharacterCommand{})
	register(commands.ManuscriptCommand{})
	register(commands.IncomingLinksCommand{})
	register(commands.OutgoingLinksCommand{})
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	for _, c := range registry {
		fmt.Fprintln(os.Stderr, "  brain", c.Usage())
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	c := lookup(os.Args[1])
	if c == nil {
		usage()
		return
	}

	if c.RequiresRoot() {
		root, err := commands.FindRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		if err := os.Chdir(filepath.Dir(root)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}

	if err := c.Run(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
