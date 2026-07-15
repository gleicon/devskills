// Package scaffold manages the devskills-owned blocks inside a project's
// user-editable AGENTS.md and CLAUDE.md. Each block is delimited by
// <!-- BEGIN/END devskills:<id> --> markers (byte-compatible with the earlier
// bash installer, so a project it set up is recognized and upgraded in place).
// The engine only ever touches content between its own markers; everything else
// the user wrote is preserved, and any file it changes is backed up once first.
package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Engine upserts and removes managed blocks. It carries per-run bookkeeping so a
// pre-existing file is backed up at most once, and files it created this run are
// never backed up. Not safe for concurrent use — one run is sequential.
type Engine struct {
	dryRun  bool
	stamp   string          // suffix for .bak files, fixed for the whole run
	created map[string]bool // files created this run — never back up
	backed  map[string]bool // files already backed up this run
	log     func(string)
}

// New builds an engine stamped with the current time for backup filenames. log
// receives one progress line per action; pass nil to discard.
func New(dryRun bool, log func(string)) *Engine {
	return newEngine(dryRun, time.Now().Format("20060102-150405"), log)
}

func newEngine(dryRun bool, stamp string, log func(string)) *Engine {
	if log == nil {
		log = func(string) {}
	}
	return &Engine{dryRun: dryRun, stamp: stamp, created: map[string]bool{}, backed: map[string]bool{}, log: log}
}

var importLine = regexp.MustCompile(`(?m)^[[:space:]]*@AGENTS\.md`)

var blockBegin = regexp.MustCompile(`(?m)^<!-- BEGIN devskills:(\S+) -->$`)

// BlockIDs returns the ids of every devskills-managed block present in file, in
// file order — nil if the file is absent or has none. It is the input to
// RemoveBlocks when backing devskills out entirely, so any id (language:<lang>,
// a retired one) is caught without a hardcoded list.
func (e *Engine) BlockIDs(file string) ([]string, error) {
	content, existed, err := readMaybe(file)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	var ids []string
	for _, m := range blockBegin.FindAllStringSubmatch(content, -1) {
		ids = append(ids, m[1])
	}
	return ids, nil
}

// Upsert inserts or replaces the managed block <id> in file, wrapping body in the
// markers. An existing block is replaced in place; otherwise the block is
// appended after the current content. A byte-identical result writes nothing.
func (e *Engine) Upsert(file, id, body string) error {
	orig, existed, err := readMaybe(file)
	if err != nil {
		return err
	}
	wrapped := wrapBlock(id, body)
	var next string
	if re := blockRE(id); existed && re.MatchString(orig) {
		next = re.ReplaceAllLiteralString(orig, wrapped)
	} else {
		next = appendBlock(orig, wrapped)
	}
	return e.writeChange(file, orig, next, existed)
}

// EnsureClaudeImport makes claudePath import AGENTS.md via @AGENTS.md. A file that
// already imports it by hand (a bare @AGENTS.md line, no marker) is left alone,
// so we don't fight a user's own wiring.
func (e *Engine) EnsureClaudeImport(claudePath string) error {
	orig, existed, err := readMaybe(claudePath)
	if err != nil {
		return err
	}
	if existed && importLine.MatchString(orig) && !strings.Contains(orig, beginMarker("import")) {
		e.log("CLAUDE.md already imports AGENTS.md; leaving as-is.")
		return nil
	}
	return e.Upsert(claudePath, "import", "@AGENTS.md")
}

// RemoveBlocks strips the managed blocks named by ids from file. When only
// devskills content remains, the file is deleted rather than left empty. A file
// with none of our markers, or no file at all, is untouched.
func (e *Engine) RemoveBlocks(file string, ids ...string) error {
	base := filepath.Base(file)
	orig, existed, err := readMaybe(file)
	if err != nil {
		return err
	}
	if !existed {
		e.log(base + ": nothing to remove.")
		return nil
	}
	found := false
	for _, id := range ids {
		if strings.Contains(orig, beginMarker(id)) {
			found = true
		}
	}
	if !found {
		e.log(base + ": no devskills blocks; left as-is.")
		return nil
	}

	next := orig
	for _, id := range ids {
		next = blockRE(id).ReplaceAllLiteralString(next, "")
	}
	next = normalizeBlankLines(next)

	if next == "" {
		if e.dryRun {
			e.logf("[dry] would back up %s and remove it (only devskills content).", base)
			return nil
		}
		if err := e.backupOnce(file); err != nil {
			return err
		}
		if err := os.Remove(file); err != nil {
			return err
		}
		e.logf("removed %s (held only devskills content).", base)
		return nil
	}
	if next == orig {
		e.log(base + ": no change.")
		return nil
	}
	if e.dryRun {
		e.logf("[dry] would back up %s and strip devskills blocks.", base)
		return nil
	}
	if err := e.backupOnce(file); err != nil {
		return err
	}
	if err := os.WriteFile(file, []byte(next), 0o644); err != nil {
		return err
	}
	e.logf("stripped devskills blocks from %s.", base)
	return nil
}

// writeChange applies next to file, handling the no-op, dry-run, backup-once, and
// create-vs-update cases uniformly.
func (e *Engine) writeChange(file, orig, next string, existed bool) error {
	base := filepath.Base(file)
	if existed && next == orig {
		e.logf("%s already up to date.", base)
		return nil
	}
	if e.dryRun {
		if existed {
			e.logf("[dry] would back up %s and update it.", base)
		} else {
			e.logf("[dry] would create %s.", base)
		}
		return nil
	}
	if existed {
		if err := e.backupOnce(file); err != nil {
			return err
		}
		if err := os.WriteFile(file, []byte(next), 0o644); err != nil {
			return err
		}
		e.logf("updated %s", base)
		return nil
	}
	if err := os.WriteFile(file, []byte(next), 0o644); err != nil {
		return err
	}
	e.created[file] = true
	e.logf("created %s", base)
	return nil
}

// backupOnce copies file to a sibling timestamped .bak, at most once per run and
// never for a file this run created.
func (e *Engine) backupOnce(file string) error {
	if e.created[file] || e.backed[file] {
		return nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	bak := fmt.Sprintf("%s.%s.bak", file, e.stamp)
	if err := os.WriteFile(bak, data, 0o644); err != nil {
		return err
	}
	e.backed[file] = true
	e.logf("backed up %s -> %s", filepath.Base(file), filepath.Base(bak))
	return nil
}

func (e *Engine) logf(format string, a ...any) { e.log(fmt.Sprintf(format, a...)) }

func beginMarker(id string) string { return "<!-- BEGIN devskills:" + id + " -->" }
func endMarker(id string) string   { return "<!-- END devskills:" + id + " -->" }

// blockRE matches a whole managed block for id: the begin-marker line through the
// end-marker line, plus a trailing newline. Markers are matched as full lines.
func blockRE(id string) *regexp.Regexp {
	return regexp.MustCompile("(?ms)^" + regexp.QuoteMeta(beginMarker(id)) + "$\n.*?^" + regexp.QuoteMeta(endMarker(id)) + "$\n?")
}

// wrapBlock renders body between id's markers, each on its own line and the body
// newline-terminated.
func wrapBlock(id, body string) string {
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return beginMarker(id) + "\n" + body + endMarker(id) + "\n"
}

// appendBlock adds wrapped to content, separated by a blank line, preserving what
// was there. An empty file becomes just the block.
func appendBlock(content, wrapped string) string {
	if content == "" {
		return wrapped
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + "\n" + wrapped
}

// normalizeBlankLines drops leading and trailing blank lines and collapses
// internal blank runs to a single blank — so stripping a block doesn't leave a
// gap. All-blank input yields "".
func normalizeBlankLines(s string) string {
	var out []string
	printed, pending := false, false
	for ln := range strings.SplitSeq(s, "\n") {
		if strings.TrimSpace(ln) == "" {
			pending = true
			continue
		}
		if printed && pending {
			out = append(out, "")
		}
		out = append(out, ln)
		printed, pending = true, false
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n") + "\n"
}

func readMaybe(file string) (content string, existed bool, err error) {
	b, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	// Normalize CRLF so a Windows-checkout file matches the LF-anchored block
	// regex; it is rewritten with LF on the next change.
	return strings.ReplaceAll(string(b), "\r\n", "\n"), true, nil
}
