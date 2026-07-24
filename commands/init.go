package commands

import (
	"errors"
	"fmt"
	"os"
)

type InitCommand struct{}

func (InitCommand) Name() string       { return "init" }
func (InitCommand) Usage() string      { return "init" }
func (InitCommand) RequiresRoot() bool { return false }

func (c InitCommand) Run(args []string) error {
	if len(args) > 0 {
		return errors.New("'brain init' does not take command line arguments")
	}

	_, err := os.Stat(RootFile)
	if err == nil {
		return fmt.Errorf("%s already exists", RootFile)
	}

	if !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(RootFile)
	if err != nil {
		return err
	}

	return file.Close()
}
