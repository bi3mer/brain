package commands

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// brainDirs is every directory `brain init` creates, in the order it creates
// them. Kept in sync with InitCommand.Run by hand. The order is part of the
// contract the tests lean on: dirs come before the marker file, and a later
// dir failing means every earlier one has to be rolled back.
var brainDirs = []string{
	"index",
	"inbox",
	"literature",
	"permanent",
	"reference",
	"projects",
	filepath.Join("projects", "future"),
	filepath.Join("projects", "active"),
	filepath.Join("projects", "archive"),
	"stubs",
	"templates",
}

// brainRootEntries is everything a fresh brain has sitting in its root: the
// dirs from brainDirs with no parent, plus the marker file. Spelled out
// instead of filtered out of brainDirs so a new entry has to be added here
// deliberately — that's the whole point of TestInitCreatesNothingExtra.
var brainRootEntries = []string{
	RootFile,
	"index",
	"inbox",
	"literature",
	"permanent",
	"reference",
	"projects",
	"stubs",
	"templates",
}

// wantDir fails the test unless path is a directory the owner can actually
// use. The permission check is not decoration: a 0644 mkdir would leave a
// directory that exists, passes IsDir, and can't be entered or written to.
func wantDir(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("expected directory %s: %v", path, err)
		return
	}

	if !info.IsDir() {
		t.Errorf("%s exists but is not a directory", path)
		return
	}

	if info.Mode().Perm()&0700 != 0700 {
		t.Errorf("%s is not owner rwx: mode %v", path, info.Mode().Perm())
	}
}

// wantNoSuchFile fails the test if path exists in any form.
func wantNoSuchFile(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s to be absent: %v", path, err)
	}
}

// sortedEntryNames lists dir, sorted, so it can be compared against a
// hand-written list without caring what order the filesystem hands back.
func sortedEntryNames(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}

	slices.Sort(names)
	return names
}

func TestInitCreatesBrain(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := InitCommand{}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("init: %v", err)
	}

	for _, dir := range brainDirs {
		wantDir(t, dir)
	}

	info, err := os.Stat(RootFile)
	if err != nil {
		t.Fatalf("expected marker %s: %v", RootFile, err)
	}

	if !info.Mode().IsRegular() {
		t.Errorf("marker %s is not a regular file: mode %v", RootFile, info.Mode())
	}
}

func TestInitCreatesNothingExtra(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := InitCommand{}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("init: %v", err)
	}

	want := slices.Clone(brainRootEntries)
	slices.Sort(want)

	got := sortedEntryNames(t, ".")
	if !slices.Equal(got, want) {
		t.Errorf("brain root is %v, want %v", got, want)
	}
}

// TestInitIgnoresArguments pins the deliberate choice that init takes no
// arguments and does not care if it is handed some: extra args are dropped
// and the brain is built exactly as it would be with none.
func TestInitIgnoresArguments(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := InitCommand{}

	if err := cmd.Run([]string{"extra"}); err != nil {
		t.Fatalf("init with an argument: %v", err)
	}

	for _, dir := range brainDirs {
		wantDir(t, dir)
	}
}

func TestInitRefusesExistingMarker(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.WriteFile(RootFile, nil, 0644); err != nil {
		t.Fatalf("write marker %s: %v", RootFile, err)
	}

	cmd := InitCommand{}

	err := cmd.Run(nil)
	if err == nil {
		t.Fatal("init over an existing marker should fail")
	}

	want := RootFile + " already exists"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("init failed with %q, want it to mention %q", err, want)
	}

	// The absolute path is the useful half of the message: someone who ran
	// init in the wrong terminal learns which brain they already have.
	dir, wdErr := os.Getwd()
	if wdErr != nil {
		t.Fatalf("getwd: %v", wdErr)
	}

	wantPath := filepath.Join(dir, RootFile)
	if !strings.Contains(err.Error(), wantPath) {
		t.Errorf("init failed with %q, want it to name %q", err, wantPath)
	}

	// The marker check happens before the first Mkdir, so a brain that was
	// only half set up keeps whatever layout it already had.
	for _, dir := range brainDirs {
		wantNoSuchFile(t, dir)
	}
}

func TestInitCleansUpAfterFailure(t *testing.T) {
	t.Chdir(t.TempDir())

	// stubs is created late in the loop, so pre-creating it makes that
	// Mkdir fail EEXIST partway through and sends init down its cleanup
	// path with most of the brain already on disk.
	if err := os.Mkdir("stubs", 0755); err != nil {
		t.Fatalf("seed stubs: %v", err)
	}

	cmd := InitCommand{}

	if err := cmd.Run(nil); err == nil {
		t.Fatal("init should fail when one of its dirs already exists")
	}

	// Everything init made before stubs is rolled back, and templates was
	// never reached. Only the dir the test seeded survives.
	for _, dir := range brainDirs {
		if dir == "stubs" {
			continue
		}
		wantNoSuchFile(t, dir)
	}

	wantDir(t, "stubs")
	wantNoSuchFile(t, RootFile)
}

func TestInitTwiceLeavesFirstBrainAlone(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := InitCommand{}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("first init: %v", err)
	}

	if err := cmd.Run(nil); err == nil {
		t.Fatal("second init should fail")
	}

	// The failed run must not drag the working brain into its cleanup.
	for _, dir := range brainDirs {
		wantDir(t, dir)
	}

	if _, err := os.Stat(RootFile); err != nil {
		t.Errorf("marker %s gone after the second init: %v", RootFile, err)
	}
}
