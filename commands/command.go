package commands

// Command is a brain subcommand. Name is the word typed on the command
// line (e.g. "rename"); Usage is the argument list shown after "brain
// <name>" in help and error output. RequiresRoot reports whether the
// command needs the vault root resolved and chdir'd into before Run — true
// for commands that use vault-relative paths, false for ones that operate on
// an explicit argument or bootstrap the vault (init).
type Command interface {
	Name() string
	Usage() string
	RequiresRoot() bool
	Run(args []string) error
}
