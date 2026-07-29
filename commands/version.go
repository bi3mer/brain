package commands

import (
	"fmt"
)

type Version struct{}

func (Version) Name() string       { return "version" }
func (Version) Usage() string      { return "version" }
func (Version) RequiresRoot() bool { return false }

func (c Version) Run(_ []string) error {
	fmt.Printf("brain %s (%s)\n", VersionString, buildVersion())

	return nil
}
