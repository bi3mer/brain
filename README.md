# Brain

This is very much a tool-in-progress. The idea is for this to be a commandline tool that makes it easy to build a digital [Zettelkatsen](https://en.wikipedia.org/wiki/Zettelkasten).

I started with making markdown files and manually linking them together, but the process needs automation. The problem that I'm running into is that markdown paths being local ends up complicating a lot of the logic, and so I am going to need `brain` to have some kind of footprint to keep track of the zettelkasten state.

To that end, I think I'm going to move towards a `sqlite` solution behind the scenes to make everything easier to manage. Maybe I'll come up with something else later. I don't know.

## Install

Requires [Go](https://go.dev/dl/) 1.26.4 or newer.

```bash
go install github.com/bi3mer/brain@latest
```

This drops a `brain` binary in `$GOBIN`, or `$GOPATH/bin` (usually `~/go/bin`) if `GOBIN` isn't set. If that directory isn't on your `PATH`, add it:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

Use `~/.zshrc` instead if you're on zsh. Then check it worked:

```bash
brain
```

You should see the usage listing.

### From source

```bash
git clone https://github.com/bi3mer/brain.git
cd brain
go install .
```

## Tests

```bash
go test ./commands/
```
