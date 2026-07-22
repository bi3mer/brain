// brain is the CLI for this vault: creating, renaming, and deleting zettel
// notes, and assembling a manuscript from a chapter directory's .meta file.
// The command implementations live in internal/, exported so this file can
// wire them up; internal/ itself stays unimportable from outside this
// module (Go's "internal" convention).
//
// Usage:
//
//	go run ./brain create <type> <title...>
//	go run ./brain rename <old-path> <new-path> [new-title]
//	go run ./brain delete <path>
//	go run ./brain manuscript <chapters-dir>
//
// Run from the vault repo root — every path a command takes is relative to
// the working directory, not to this binary's location.
package main

import (
	"fmt"
	"os"

	"tools/internal"
)

// commands holds every registered subcommand, in registration order.
var commands []internal.Command

func register(c internal.Command) {
	commands = append(commands, c)
}

func lookup(name string) internal.Command {
	for _, c := range commands {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func init() {
	register(internal.RenameCommand{})
	register(internal.DeleteCommand{})
	register(internal.CreateCommand{})
	register(internal.DndCharacterCommand{})
	register(internal.ManuscriptCommand{})
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	for _, c := range commands {
		fmt.Fprintln(os.Stderr, "  brain", c.Usage())
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
	} else {
		c := lookup(os.Args[1])

		if c == nil {
			usage()
		} else if err := c.Run(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
