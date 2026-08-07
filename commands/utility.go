package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// linkScanDirs are the directories walked for markdown links that may
// reference a renamed or deleted note.
//
// TODO: just search all directories that don't start with .
var linkScanDirs = []string{
	"permanent",
	"stubs",
	"reference",
	"literature",
	"index",
	"writing",
	"projects",
	"inbox",
}

var smallWords = map[string]bool{
	"a": true, "an": true, "the": true,
	"and": true, "but": true, "or": true, "nor": true,
	"as": true, "at": true, "by": true, "for": true, "in": true,
	"of": true, "on": true, "per": true, "to": true, "up": true,
	"via": true, "vs": true, "with": true, "from": true, "into": true,
}

// FindRoot walks up from the current working directory looking for the
// RootFile marker (see InitCommand), returning its absolute path. It lets a
// command locate the brain root no matter which subdirectory it's run from.
// Errors if no RootFile is found before the filesystem root.
func FindRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, RootFile)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s not found in any parent directory", RootFile)
		}

		dir = parent
	}
}

// EnterBrainRoot locates the brain root with FindRoot and chdirs into it.
// Every command with RequiresRoot() true is dispatched through this, so it
// can treat brain-relative paths as relative to the cwd no matter which
// subdirectory the user actually ran brain from.
func EnterBrainRoot() error {
	root, err := FindRoot()
	if err != nil {
		return err
	}

	return os.Chdir(filepath.Dir(root))
}

// titleFromFilename derives a best-effort title from a kebab-case filename.
// It's lossy for titles with punctuation kebab-case can't encode
// (apostrophes, commas, etc), so it's only used as a fallback when a file
// has no real H1 to read (see resolveTitle) or as the default guess for a
// brand new filename that hasn't had its H1 written yet.
func titleFromFilename(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	words := strings.Split(base, "-")
	for i, w := range words {
		if w == "" {
			continue
		}
		lower := strings.ToLower(w)
		if i != 0 && i != len(words)-1 && smallWords[lower] {
			words[i] = lower
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
	return strings.Join(words, " ")
}

// filenameFromTitle derives a kebab-case filename from a note title: lower
// case, every run of non [a-z0-9] characters collapsed to a single hyphen,
// and no leading or trailing hyphen. It's the inverse of titleFromFilename,
// used when creating a new note from a title rather than reading one back.
func filenameFromTitle(title string) string {
	var b strings.Builder
	prevHyphen := true // treat start-of-string like a hyphen so leading runs are dropped
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}
		if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// h1Title reads path and returns the text of its first H1 heading — the
// first non-blank line, which must start with "# ". This is the source of
// truth for a note's title; unlike titleFromFilename it can't mismatch
// itself on punctuation a kebab-case filename can't encode.
func h1Title(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(trimmed[2:]), true
		}
		return "", false
	}
	return "", false
}

// resolveTitle returns a note's real title, read from its H1, falling back
// to a filename-derived guess only if the file has no H1 at all.
func resolveTitle(path string) string {
	if title, ok := h1Title(path); ok {
		return title
	}
	return titleFromFilename(path)
}

// rewriteH1 replaces path's first H1 heading (see h1Title) with newTitle,
// touching nothing else in the file. Errors if there's no H1 to replace.
func rewriteH1(path string, newTitle string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "# ") {
			break
		}
		lines[i] = "# " + newTitle
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	}
	return fmt.Errorf("no H1 heading found in %s", path)
}

// eachLink parses inline markdown links [text](href) in data and calls
// visit(text, href) for each internal one, skipping external links (href
// containing "://"). It's the read-only core shared by the link-reporting
// commands; it uses the same scan as rewriteMatchingLinks but rewrites
// nothing.
func eachLink(data []byte, visit func(text, href string)) {
	i := 0
	for i < len(data) {
		open := bytes.IndexByte(data[i:], '[')
		if open == -1 {
			return
		}
		open += i

		closeText := bytes.IndexByte(data[open:], ']')
		if closeText == -1 {
			return
		}
		closeText += open

		if closeText+1 >= len(data) || data[closeText+1] != '(' {
			i = closeText + 1
			continue
		}

		closeHref := bytes.IndexByte(data[closeText+2:], ')')
		if closeHref == -1 {
			return
		}
		closeHref += closeText + 2

		text := string(data[open+1 : closeText])
		href := string(data[closeText+2 : closeHref])
		i = closeHref + 1

		if strings.Contains(href, "://") {
			continue
		}
		visit(text, href)
	}
}

// walkBrainFiles calls visit(path) for every .md file under linkScanDirs.
func walkBrainFiles(visit func(path string)) {
	for _, dir := range linkScanDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			visit(path)
			return nil
		})
	}
}

func rewriteMatchingLinks(path, targetAbs string, replace func(filePath, fileAbsDir, text, href string) string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	dir := filepath.Dir(path)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}
	changed := false
	var out []byte

	i := 0
	for i < len(data) {
		open := bytes.IndexByte(data[i:], '[')
		if open == -1 {
			out = append(out, data[i:]...)
			break
		}
		open += i
		out = append(out, data[i:open]...)

		closeText := bytes.IndexByte(data[open:], ']')
		if closeText == -1 {
			out = append(out, data[open:]...)
			break
		}
		closeText += open

		if closeText+1 >= len(data) || data[closeText+1] != '(' {
			out = append(out, data[open:closeText+1]...)
			i = closeText + 1
			continue
		}

		closeHref := bytes.IndexByte(data[closeText+2:], ')')
		if closeHref == -1 {
			out = append(out, data[open:]...)
			break
		}
		closeHref += closeText + 2

		text := string(data[open+1 : closeText])
		href := string(data[closeText+2 : closeHref])

		if strings.Contains(href, "://") {
			out = append(out, data[open:closeHref+1]...)
			i = closeHref + 1
			continue
		}

		target, err := filepath.Abs(filepath.Join(dir, href))
		if err != nil || target != targetAbs {
			out = append(out, data[open:closeHref+1]...)
			i = closeHref + 1
			continue
		}

		out = append(out, []byte(replace(path, absDir, text, href))...)
		changed = true
		i = closeHref + 1
	}

	if changed {
		if err := os.WriteFile(path, out, 0644); err != nil {
			return false, err
		}
	}
	return changed, nil
}

func updateBrainLinks(targetAbs, skipAbs string, replace func(filePath, fileAbsDir, text, href string) string, onFileChanged func(path string)) int {
	filesChanged := 0
	walkBrainFiles(func(path string) {
		absPath, err := filepath.Abs(path)
		if err != nil || absPath == skipAbs {
			return
		}

		changed, err := rewriteMatchingLinks(path, targetAbs, replace)
		if err != nil {
			fmt.Fprintln(os.Stderr, "fix links in", path, ":", err)
			return
		}

		if changed {
			filesChanged++
			if onFileChanged != nil {
				onFileChanged(path)
			}
		}
	})
	return filesChanged
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}

	return info.Main.Version
}
