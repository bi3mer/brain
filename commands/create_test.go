package commands

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// newBrain puts the test in a fresh brain and leaves the cwd at its root,
// which is where main.go has already chdir'd by the time a RequiresRoot
// command runs. Every create test needs one: create reads the templates
// init writes, so the two commands are only meaningful together.
func newBrain(t *testing.T) {
	t.Helper()

	t.Chdir(t.TempDir())

	if err := (InitCommand{}).Run(nil); err != nil {
		t.Fatalf("init: %v", err)
	}
}

// TestCreateWritesEveryNoteType walks createNoteTypes rather than a
// hand-written list, so a type added to the map is covered here without
// anyone remembering to add it.
func TestCreateWritesEveryNoteType(t *testing.T) {
	for typeName, nt := range createNoteTypes {
		t.Run(typeName, func(t *testing.T) {
			newBrain(t)

			if err := (CreateCommand{}).Run([]string{typeName, "My Note Title"}); err != nil {
				t.Fatalf("create %s: %v", typeName, err)
			}

			dest := filepath.Join(nt.dir, "my-note-title.md")
			data, err := os.ReadFile(dest)
			if err != nil {
				t.Fatalf("expected note at %s: %v", dest, err)
			}

			// An unsubstituted REPLACE means the template landed on disk
			// verbatim and the note has no title.
			if strings.Contains(string(data), "REPLACE") {
				t.Errorf("%s still contains REPLACE:\n%s", dest, data)
			}

			// Contains, not equality: the index template deliberately
			// prefixes its H1 with "Index: ".
			title, ok := h1Title(dest)
			if !ok {
				t.Fatalf("%s has no H1", dest)
			}
			if !strings.Contains(title, "My Note Title") {
				t.Errorf("H1 of %s is %q, want it to contain %q", dest, title, "My Note Title")
			}
		})
	}
}

// TestEveryNoteTypeResolves is the cross-file contract between create and
// init: create names a directory and a template for each type, and init has
// to have made both. A type whose template init never writes fails at
// runtime with "read template", not at compile time.
func TestEveryNoteTypeResolves(t *testing.T) {
	newBrain(t)

	for typeName, nt := range createNoteTypes {
		wantDir(t, nt.dir)

		template := filepath.Join(TemplateDirectory, nt.template)
		if _, err := os.Stat(template); err != nil {
			t.Errorf("type %s wants template %s: %v", typeName, template, err)
		}
	}
}

// TestCreateTypesMsgMatchesTheMap pins the usage string to the map it
// describes. These drifted apart once already: stub was advertised for
// months while createNoteTypes had no entry for it, so `brain create stub`
// answered "unknown type: stub".
func TestCreateTypesMsgMatchesTheMap(t *testing.T) {
	_, list, ok := strings.Cut(createTypesMsg, ":")
	if !ok {
		t.Fatalf("createTypesMsg %q has no colon to split on", createTypesMsg)
	}

	advertised := strings.Split(list, "|")
	for i, name := range advertised {
		advertised[i] = strings.TrimSpace(name)
	}
	slices.Sort(advertised)

	routed := make([]string, 0, len(createNoteTypes))
	for typeName := range createNoteTypes {
		routed = append(routed, typeName)
	}
	slices.Sort(routed)

	if !slices.Equal(advertised, routed) {
		t.Errorf("createTypesMsg lists %v, createNoteTypes routes %v", advertised, routed)
	}
}

// TestCreateStubUsesPermanentTemplate pins the one type whose template does
// not match its own name: a stub is a permanent note nobody has written yet.
func TestCreateStubUsesPermanentTemplate(t *testing.T) {
	newBrain(t)

	cmd := CreateCommand{}

	if err := cmd.Run([]string{"stub", "Same Title"}); err != nil {
		t.Fatalf("create stub: %v", err)
	}

	if err := cmd.Run([]string{"permanent", "Same Title"}); err != nil {
		t.Fatalf("create permanent: %v", err)
	}

	stub, err := os.ReadFile(filepath.Join("stubs", "same-title.md"))
	if err != nil {
		t.Fatalf("read stub: %v", err)
	}

	permanent, err := os.ReadFile(filepath.Join("permanent", "same-title.md"))
	if err != nil {
		t.Fatalf("read permanent: %v", err)
	}

	// Same template and same title, so the only difference is the directory.
	if string(stub) != string(permanent) {
		t.Errorf("stub and permanent notes differ:\n--- stub ---\n%s\n--- permanent ---\n%s", stub, permanent)
	}
}

// TestCreateSlugsTheTitle checks the title survives two different trips:
// into the filename as a slug, and into the H1 verbatim.
func TestCreateSlugsTheTitle(t *testing.T) {
	newBrain(t)

	title := "Zettelkasten: Notes, Links & Other Things!"

	if err := (CreateCommand{}).Run([]string{"permanent", title}); err != nil {
		t.Fatalf("create: %v", err)
	}

	dest := filepath.Join("permanent", "zettelkasten-notes-links-other-things.md")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected note at %s: %v", dest, err)
	}

	// The H1 keeps the punctuation the filename had to drop.
	if got, _ := h1Title(dest); got != title {
		t.Errorf("H1 is %q, want %q", got, title)
	}
}

func TestCreateRejectsUnknownType(t *testing.T) {
	newBrain(t)

	err := (CreateCommand{}).Run([]string{"bogus", "Some Title"})
	if err == nil {
		t.Fatal("create with an unknown type should fail")
	}

	if !strings.Contains(err.Error(), "unknown type: bogus") {
		t.Errorf("create failed with %q, want it to name the bad type", err)
	}

	// A rejected type must not leave a note anywhere.
	for _, nt := range createNoteTypes {
		if got := sortedEntryNames(t, nt.dir); len(got) != 0 {
			t.Errorf("%s holds %v after a rejected create", nt.dir, got)
		}
	}
}

func TestCreateRejectsMissingArguments(t *testing.T) {
	newBrain(t)

	cmd := CreateCommand{}

	// No type and no title.
	if err := cmd.Run(nil); err == nil {
		t.Error("create with no arguments should fail")
	}

	// A type but no title: the note would have an empty filename.
	if err := cmd.Run([]string{"permanent"}); err == nil {
		t.Error("create with no title should fail")
	}

	if got := sortedEntryNames(t, "permanent"); len(got) != 0 {
		t.Errorf("permanent holds %v after a rejected create", got)
	}
}

// TestCreateRefusesToOverwrite matters more than the usual "second call
// fails" check: the file it would clobber is a note the user has written.
func TestCreateRefusesToOverwrite(t *testing.T) {
	newBrain(t)

	cmd := CreateCommand{}

	if err := cmd.Run([]string{"permanent", "Same Title"}); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// Stand in for a note the user has since filled out.
	dest := filepath.Join("permanent", "same-title.md")
	written := "# Same Title\n\nWork that must survive.\n"
	if err := os.WriteFile(dest, []byte(written), 0644); err != nil {
		t.Fatalf("write note: %v", err)
	}

	err := cmd.Run([]string{"permanent", "Same Title"})
	if err == nil {
		t.Fatal("create over an existing note should fail")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("create failed with %q, want it to mention the file exists", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read note: %v", err)
	}
	if string(data) != written {
		t.Errorf("note was overwritten:\n%s", data)
	}
}

// TestCreateFromSubdirectory covers running `brain create` from somewhere
// deep in the brain rather than its root. create resolves both its template
// and its destination against the cwd, so without the EnterBrainRoot step
// main.go performs for every RequiresRoot command, it would look for
// templates/ under projects/active and write the note there too.
func TestCreateFromSubdirectory(t *testing.T) {
	newBrain(t)

	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if !(CreateCommand{}).RequiresRoot() {
		t.Fatal("create must require the root for the chdir below to happen")
	}

	t.Chdir(filepath.Join(root, "projects", "active"))

	if err := EnterBrainRoot(); err != nil {
		t.Fatalf("enter brain root: %v", err)
	}

	if err := (CreateCommand{}).Run([]string{"permanent", "Deep Note"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	// The note belongs to the brain, not to the directory it was typed from.
	if _, err := os.Stat(filepath.Join(root, "permanent", "deep-note.md")); err != nil {
		t.Errorf("expected note at permanent/deep-note.md: %v", err)
	}

	wantNoSuchFile(t, filepath.Join(root, "projects", "active", "permanent"))
	wantNoSuchFile(t, filepath.Join(root, "projects", "active", TemplateDirectory))
}

// TestCreateReportsAMissingTemplate covers the brain whose templates
// directory has been edited by hand since init ran.
func TestCreateReportsAMissingTemplate(t *testing.T) {
	newBrain(t)

	if err := os.Remove(filepath.Join(TemplateDirectory, "literature.md")); err != nil {
		t.Fatalf("remove template: %v", err)
	}

	err := (CreateCommand{}).Run([]string{"literature", "Some Title"})
	if err == nil {
		t.Fatal("create without a template should fail")
	}

	if !strings.Contains(err.Error(), "read template") {
		t.Errorf("create failed with %q, want it to blame the template", err)
	}

	wantNoSuchFile(t, filepath.Join("literature", "some-title.md"))
}
