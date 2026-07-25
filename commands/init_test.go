package commands

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// vaultDirs is every directory `brain init` creates, in the order it creates
// them. Kept in sync with InitCommand.Run by hand. The order is part of the
// contract the tests lean on: dirs come before the marker file, and a later
// dir failing means every earlier one has to be rolled back.
var vaultDirs = []string{
	"index",
	"inbox",
	"literature",
	"permanent",
	"projects",
	filepath.Join("projects", "future"),
	filepath.Join("projects", "active"),
	filepath.Join("projects", "archive"),
	"stubs",
	"templates",
}

// vaultRootEntries is everything a fresh vault has sitting in its root: the
// dirs from vaultDirs with no parent, plus the marker file. Spelled out
// instead of filtered out of vaultDirs so a new entry has to be added here
// deliberately — that's the whole point of TestInitCreatesNothingExtra.
var vaultRootEntries = []string{
	RootFile,
	"index",
	"inbox",
	"literature",
	"permanent",
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

func TestInitCreatesVault(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := InitCommand{}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("init: %v", err)
	}

	for _, dir := range vaultDirs {
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

	want := slices.Clone(vaultRootEntries)
	slices.Sort(want)

	got := sortedEntryNames(t, ".")
	if !slices.Equal(got, want) {
		t.Errorf("vault root is %v, want %v", got, want)
	}
}

func TestInitRejectsArguments(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := InitCommand{}

	if err := cmd.Run([]string{"extra"}); err == nil {
		t.Fatal("init with an argument should fail")
	}

	// The argument guard runs first, so the directory is still empty.
	if got := sortedEntryNames(t, "."); len(got) != 0 {
		t.Errorf("init wrote %v after refusing its arguments", got)
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

	// The marker check happens before the first Mkdir, so a vault that was
	// only half set up keeps whatever layout it already had.
	for _, dir := range vaultDirs {
		wantNoSuchFile(t, dir)
	}
}

func TestInitCleansUpAfterFailure(t *testing.T) {
	t.Chdir(t.TempDir())

	// stubs is created late in the loop, so pre-creating it makes that
	// Mkdir fail EEXIST partway through and sends init down its cleanup
	// path with most of the vault already on disk.
	if err := os.Mkdir("stubs", 0755); err != nil {
		t.Fatalf("seed stubs: %v", err)
	}

	cmd := InitCommand{}

	if err := cmd.Run(nil); err == nil {
		t.Fatal("init should fail when one of its dirs already exists")
	}

	// Everything init made before stubs is rolled back, and templates was
	// never reached. Only the dir the test seeded survives.
	for _, dir := range vaultDirs {
		if dir == "stubs" {
			continue
		}
		wantNoSuchFile(t, dir)
	}

	wantDir(t, "stubs")
	wantNoSuchFile(t, RootFile)
}

func TestInitTwiceLeavesFirstVaultAlone(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := InitCommand{}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("first init: %v", err)
	}

	if err := cmd.Run(nil); err == nil {
		t.Fatal("second init should fail")
	}

	// The failed run must not drag the working vault into its cleanup.
	for _, dir := range vaultDirs {
		wantDir(t, dir)
	}

	if _, err := os.Stat(RootFile); err != nil {
		t.Errorf("marker %s gone after the second init: %v", RootFile, err)
	}
}
