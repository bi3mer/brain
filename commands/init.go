package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type InitCommand struct{}

func (InitCommand) Name() string       { return "init" }
func (InitCommand) Usage() string      { return "init" }
func (InitCommand) RequiresRoot() bool { return false }

func (c InitCommand) Run(args []string) error {
	if len(args) > 0 {
		return errors.New("'brain init' does not take command line arguments")
	}

	// check if brainroot already exists
	_, err := os.Stat(RootFile)
	if err == nil {
		return fmt.Errorf("%s already exists", RootFile)
	}

	if !os.IsNotExist(err) {
		return err
	}

	// create directories
	directories := []string{
		"index", "inbox", "literature", "permanent", "projects", filepath.Join("projects", "future"),
		filepath.Join("projects", "active"), filepath.Join("projects", "archive"), "stubs",
		"templates",
	}
	var created []string

	for _, d := range directories {
		err := os.Mkdir(d, 0755)
		if err != nil {
			cleanup(created)
			return err
		}

		created = append(created, d)
	}

	// create root file
	file, err := os.Create(RootFile)
	if err != nil {
		cleanup(created)
		return err
	}

	created = append(created, RootFile)
	if err := file.Close(); err != nil {
		cleanup(created)
		return err
	}

	return nil
}

func cleanup(created []string) {
	for _, c := range slices.Backward(created) { //cleanup
		os.Remove(c)
	}
}
