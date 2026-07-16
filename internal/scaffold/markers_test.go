package scaffold

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Tests use a fixed stamp so backup filenames are predictable.
func fixedEngine(dryRun bool) *Engine { return newEngine(dryRun, "teststamp", nil) }

func write(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func exists(path string) bool { _, err := os.Stat(path); return err == nil }

func baks(t *testing.T, dir string) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(dir, "*.bak"))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestUpsertCreatesFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	if err := fixedEngine(false).Upsert(f, "base", "hello"); err != nil {
		t.Fatal(err)
	}
	want := "<!-- BEGIN devskills:base -->\nhello\n<!-- END devskills:base -->\n"
	if got := read(t, f); got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
	if exists(f + ".teststamp.bak") {
		t.Error("a file we created must not be backed up")
	}
}

func TestUpsertAppendsPreservingUserContent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	write(t, f, "# My notes\nkeep me") // no trailing newline
	if err := fixedEngine(false).Upsert(f, "base", "principles"); err != nil {
		t.Fatal(err)
	}
	want := "# My notes\nkeep me\n\n<!-- BEGIN devskills:base -->\nprinciples\n<!-- END devskills:base -->\n"
	if got := read(t, f); got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
	if got := read(t, f+".teststamp.bak"); got != "# My notes\nkeep me" {
		t.Errorf("backup = %q, want the pre-change original", got)
	}
}

func TestUpsertReplacesInPlace(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	write(t, f, "top\n\n<!-- BEGIN devskills:base -->\nOLD\n<!-- END devskills:base -->\n\nbottom\n")
	if err := fixedEngine(false).Upsert(f, "base", "NEW"); err != nil {
		t.Fatal(err)
	}
	want := "top\n\n<!-- BEGIN devskills:base -->\nNEW\n<!-- END devskills:base -->\n\nbottom\n"
	if got := read(t, f); got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestUpsertNoOpWhenIdentical(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	content := "<!-- BEGIN devskills:base -->\nsame\n<!-- END devskills:base -->\n"
	write(t, f, content)
	if err := fixedEngine(false).Upsert(f, "base", "same"); err != nil {
		t.Fatal(err)
	}
	if got := read(t, f); got != content {
		t.Errorf("identical upsert changed content: %q", got)
	}
	if exists(f + ".teststamp.bak") {
		t.Error("a no-op upsert must not back up")
	}
}

func TestBackupOncePerRun(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	write(t, f, "user\n")
	e := fixedEngine(false)
	if err := e.Upsert(f, "base", "a"); err != nil {
		t.Fatal(err)
	}
	if err := e.Upsert(f, "concise", "b"); err != nil {
		t.Fatal(err)
	}
	if got := read(t, f+".teststamp.bak"); got != "user\n" {
		t.Errorf("backup = %q, want the run's original %q", got, "user\n")
	}
	if got := baks(t, dir); len(got) != 1 {
		t.Errorf("backups = %v, want exactly one for the run", got)
	}
}

func TestRemoveBlocksStripsLeavingUserContent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	write(t, f, "# Mine\n\n<!-- BEGIN devskills:base -->\nours\n<!-- END devskills:base -->\n\n# Also mine\n")
	if err := fixedEngine(false).RemoveBlocks(f, "base", "concise"); err != nil {
		t.Fatal(err)
	}
	want := "# Mine\n\n# Also mine\n"
	if got := read(t, f); got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
	if !exists(f + ".teststamp.bak") {
		t.Error("expected a backup before stripping")
	}
}

func TestRemoveBlocksDeletesFileWhenOnlyOurs(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "CLAUDE.md")
	write(t, f, "<!-- BEGIN devskills:import -->\n@AGENTS.md\n<!-- END devskills:import -->\n")
	if err := fixedEngine(false).RemoveBlocks(f, "import"); err != nil {
		t.Fatal(err)
	}
	if exists(f) {
		t.Error("a file holding only devskills content should be deleted")
	}
	if !exists(f + ".teststamp.bak") {
		t.Error("expected a backup before deleting")
	}
}

func TestRemoveBlocksNoMarkersLeavesAsIs(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	write(t, f, "# Just mine\n")
	if err := fixedEngine(false).RemoveBlocks(f, "base"); err != nil {
		t.Fatal(err)
	}
	if got := read(t, f); got != "# Just mine\n" {
		t.Errorf("content = %q, want it untouched", got)
	}
	if exists(f + ".teststamp.bak") {
		t.Error("no change must not back up")
	}
}

func TestEnsureClaudeImportRespectsManualImport(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "CLAUDE.md")
	write(t, f, "@AGENTS.md\n")
	if err := fixedEngine(false).EnsureClaudeImport(f); err != nil {
		t.Fatal(err)
	}
	if got := read(t, f); got != "@AGENTS.md\n" {
		t.Errorf("a hand-written import was rewritten: %q", got)
	}
}

func TestEnsureClaudeImportCreatesBlock(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "CLAUDE.md")
	if err := fixedEngine(false).EnsureClaudeImport(f); err != nil {
		t.Fatal(err)
	}
	want := "<!-- BEGIN devskills:import -->\n@AGENTS.md\n<!-- END devskills:import -->\n"
	if got := read(t, f); got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestBlockIDs(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	write(t, f, "# top\n\n<!-- BEGIN devskills:base -->\nx\n<!-- END devskills:base -->\n\n"+
		"<!-- BEGIN devskills:language:go -->\ny\n<!-- END devskills:language:go -->\n")
	ids, err := fixedEngine(false).BlockIDs(f)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(ids, []string{"base", "language:go"}) {
		t.Errorf("ids = %v, want [base language:go]", ids)
	}

	ids, err = fixedEngine(false).BlockIDs(filepath.Join(dir, "absent.md"))
	if err != nil {
		t.Fatal(err)
	}
	if ids != nil {
		t.Errorf("ids = %v, want nil for an absent file", ids)
	}
}

func TestUpsertReplacesBlockInCRLFFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	crlf := "top\r\n\r\n<!-- BEGIN devskills:base -->\r\nOLD\r\n<!-- END devskills:base -->\r\n"
	write(t, f, crlf)
	if err := fixedEngine(false).Upsert(f, "base", "NEW"); err != nil {
		t.Fatal(err)
	}
	got := read(t, f)
	if strings.Count(got, "<!-- BEGIN devskills:base -->") != 1 {
		t.Errorf("expected the block replaced in place, not duplicated:\n%q", got)
	}
	if !strings.Contains(got, "NEW") || strings.Contains(got, "OLD") {
		t.Errorf("block body not replaced:\n%q", got)
	}
	if strings.Contains(got, "\r") {
		t.Errorf("expected CRLF normalized to LF:\n%q", got)
	}
}

func TestRemoveBlocksStripsInCRLFFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	write(t, f, "# Mine\r\n\r\n<!-- BEGIN devskills:base -->\r\nx\r\n<!-- END devskills:base -->\r\n")
	if err := fixedEngine(false).RemoveBlocks(f, "base"); err != nil {
		t.Fatal(err)
	}
	if got := read(t, f); strings.Contains(got, "devskills:base") {
		t.Errorf("block not removed from a CRLF file:\n%q", got)
	}
}

// TestDryRunWritesNothing is the marker-engine --dry-run guarantee across both an
// upsert (would append) and a removal (would strip): neither touches the file or
// leaves a backup.
func TestDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "AGENTS.md")
	orig := "user\n\n<!-- BEGIN devskills:base -->\nb\n<!-- END devskills:base -->\n"
	write(t, f, orig)
	e := fixedEngine(true)
	if err := e.Upsert(f, "concise", "c"); err != nil {
		t.Fatal(err)
	}
	if err := e.RemoveBlocks(f, "base"); err != nil {
		t.Fatal(err)
	}
	if got := read(t, f); got != orig {
		t.Errorf("dry run mutated the file: %q", got)
	}
	if got := baks(t, dir); len(got) != 0 {
		t.Errorf("dry run created backups: %v", got)
	}
}
