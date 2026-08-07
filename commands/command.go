package commands

type Command interface {
	Name() string
	Usage() string
	RequiresRoot() bool
	Run(args []string) error
}
