package main

import (
	"fmt"
	"os"

	"github.com/bi3mer/brain/commands"
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
	register(commands.Update{})
	register(commands.Version{})
	register(commands.RenameCommand{})
	register(commands.DeleteCommand{})
	register(commands.CreateCommand{})
	register(commands.IncomingLinksCommand{})
	register(commands.OutgoingLinksCommand{})
	register(commands.DndCharacterCommand{})
	register(commands.ManuscriptCommand{})
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
		if err := commands.EnterBrainRoot(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err := c.Run(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
