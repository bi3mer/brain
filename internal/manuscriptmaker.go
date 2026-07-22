package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

type ManuscriptCommand struct{}

func (ManuscriptCommand) Name() string  { return "manuscript" }
func (ManuscriptCommand) Usage() string { return "manuscript <chapters-dir>" }

// Run reads <chapters-dir>.meta, a newline-separated list of chapter
// filenames relative to chapters-dir, concatenates each in order, and
// writes the result to manuscript.md.
func (c ManuscriptCommand) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: brain %s", c.Usage())
	}
	dir := args[0]

	content, err := os.ReadFile(dir + ".meta")
	if err != nil {
		return err
	}

	var manuscript strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		fmt.Println(scanner.Text())

		content, err = os.ReadFile(dir + scanner.Text())
		if err != nil {
			return err
		}

		manuscript.Write(content)
		manuscript.WriteString("\n\n\n")
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read .meta: %w", err)
	}

	if err := os.WriteFile("manuscript.md", []byte(manuscript.String()), 0644); err != nil {
		return err
	}

	fmt.Println("DONE")
	return nil
}
