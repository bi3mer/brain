package internal

// Command is a brain subcommand. Name is the word typed on the command
// line (e.g. "rename"); Usage is the argument list shown after "brain
// <name>" in help and error output.
type Command interface {
	Name() string
	Usage() string
	Run(args []string) error
}
