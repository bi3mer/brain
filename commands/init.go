package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type InitCommand struct{}

func (InitCommand) Name() string       { return "init" }
func (InitCommand) Usage() string      { return "init" }
func (InitCommand) RequiresRoot() bool { return false }

func (c InitCommand) Run(args []string) error {
	if err := brainRootDoesNotExist(); err != nil {
		return err
	}

	var created []string
	if err := buildBrainDirectories(&created); err != nil {
		cleanup(created)
		return err
	}

	if err := createRootFile(&created); err != nil {
		cleanup(created)
		return err
	}

	if err := createTemplates(&created); err != nil {
		cleanup(created)
		return err
	}

	return nil
}

func brainRootDoesNotExist() error {
	root, err := FindRoot()
	if err == nil {
		return fmt.Errorf("%s already exists: %s", RootFile, root)
	}

	return nil
}

func cleanup(created []string) {
	for _, c := range slices.Backward(created) {
		os.Remove(c)
	}
}

func buildBrainDirectories(created *[]string) error {
	directories := []string{
		"index", "inbox", "literature", "permanent", "reference", "projects", filepath.Join("projects", "future"),
		filepath.Join("projects", "active"), filepath.Join("projects", "archive"), "stubs",
		TemplateDirectory,
	}

	for _, d := range directories {
		err := os.Mkdir(d, 0755)
		if err != nil {
			return err
		}

		*created = append(*created, d)
	}

	return nil
}

func createRootFile(created *[]string) error {
	file, err := os.Create(RootFile)
	if err != nil {
		return err
	}

	*created = append(*created, RootFile)
	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func createTemplates(created *[]string) error {
	if err := createLiteratureNote(created); err != nil {
		return err
	}

	if err := createIndexNote(created); err != nil {
		return err
	}

	if err := createPermanentNoteTemplate(created); err != nil {
		return err
	}

	if err := createReferenceTemplate(created); err != nil {
		return err
	}

	if err := createProjectTemplate(created); err != nil {
		return err
	}

	return nil
}

func createLiteratureNote(created *[]string) error {
	var s strings.Builder

	s.WriteString(`# REPLACE

## Key points

-

## Permanent notes

-

## Citation

**Chicago:** [Last, First]. _REPLACE_. [City]: [Publisher], [Year]. [URL or ISBN]

**BibTeX:**

`)
	// The fences can't live in the raw string above: a Go raw string literal
	// ends at the first backtick.
	s.WriteString("```bibtex\n")
	s.WriteString(`@book{key,
  author    = {},
  title     = {},
  publisher = {},
  year      = {},
  address   = {},
}
`)
	s.WriteString("```\n")

	return writeTemplate(created, "literature.md", s.String())
}

func createIndexNote(created *[]string) error {
	return writeTemplate(created, "index.md", `# Index: REPLACE

[One sentence orienting the reader to this cluster.]

- [notes] — one line description, if you want
`)
}

func createPermanentNoteTemplate(created *[]string) error {
	return writeTemplate(created, "permanent.md", `# REPLACE

[One atomic idea in complete sentences. Written for a future self with no memory of the source.]

Links:

- [Related Notes]

Source:

- [Literature Note if relevant]
`)
}

func createReferenceTemplate(created *[]string) error {
	return writeTemplate(created, "reference.md", `# REPLACE

[Reference info]

Source: [Literature Note or Webpage]
`)
}

func createProjectTemplate(created *[]string) error {
	return writeTemplate(created, "project.md", `# REPLACE

**Goal:** [What we are trying to do]

## Next Action

- [ ] ...

## Notes

[Context, constraints, open questions.]

## Links

- [Related notes]

## Log

-
`)
}

func writeTemplate(created *[]string, name, content string) error {
	path := filepath.Join(TemplateDirectory, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}

	*created = append(*created, path)
	return nil
}
