package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type Update struct{}

func (Update) Name() string       { return "update" }
func (Update) Usage() string      { return "update" }
func (Update) RequiresRoot() bool { return false }

func (c Update) Run(args []string) error {
	if len(args) != 0 {
		return errors.New("update takes no arguments")
	}

	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go toolchain not found on PATH: %w", err)
	}

	fmt.Printf("current: %s (%s)\n", VersionString, buildVersion())

	cmd := exec.Command(goBin, "install", ModulePath+"@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Println("updated; run `brain version` to confirm")

	return nil
}
