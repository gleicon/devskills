package sync

import "slices"

// The in-binary pruning ledger. install only ever copies the current catalog,
// so without this a renamed or dropped name lingers next to its replacement
// forever. Every entry names something devskills itself shipped — never a
// user-authored file. See DECISIONS.md § "The in-binary ledger: two lists".

// legacyCommands is the frozen pre-overhaul command catalog: the ds-*.md files
// that shipped as commands before the skills migration. install ships skills/
// only now, so any of these left in a harness's legacy command/prompt dir is
// stale and shadows or duplicates the new skill. History — never grows again.
var legacyCommands = []string{
	"ds-architecture-plan.md", "ds-blueprint.md", "ds-bug-review.md",
	"ds-caveman-lite-mode.md", "ds-caveman-ultra-mode.md", "ds-code-quality-review.md",
	"ds-code-review.md", "ds-comment-review.md", "ds-data-mode.md", "ds-data-review.md",
	"ds-debug.md", "ds-deslop.md", "ds-doc-quality-review.md", "ds-explore.md", "ds-git-mode.md",
	"ds-go-review.md", "ds-grill-me.md", "ds-handoff.md", "ds-java-review.md",
	"ds-notebook-review.md", "ds-osv.md", "ds-perf-plan.md", "ds-project-checkpoint.md",
	"ds-project-config.md", "ds-project-map.md", "ds-project-resume.md", "ds-python-review.md",
	"ds-quality-gate-mode.md", "ds-recall-capture.md", "ds-recall-setup.md", "ds-recall.md",
	"ds-roadmap.md", "ds-rust-review.md", "ds-security-review.md", "ds-senior-mode.md",
	"ds-spec.md", "ds-step-mode.md", "ds-tdd-mode.md", "ds-test-mode.md",
	"ds-test-quality-review.md", "ds-tiger-style-mode.md", "ds-tldt.md", "ds-ts-review.md",
	"ds-typeset.md", "ds-ui-mode.md", "ds-ui-quality-review.md", "ds-verify-this.md",
	"ds-workflow.md", "ds-write-a-command.md", "ds-zig-review.md", "ds-zoom-out.md",
}

// renamedCommands are earlier devskills command filenames retired across
// releases. Only ds- names — the ones we can prove are ours — are listed:
//
//	ds-project-plan.md -> ds-roadmap.md (a plan-generator; left the project-* family)
//	ds-modes.md, ds-review.md -> removed (launcher pickers of 10 modes / 8 reviews
//	                             could never render, so deleted rather than degraded)
//
// Pre-prefix bare names (test.md, spec.md, frontend.md, …) are deliberately not
// swept: a bare name may be a user's own command, and the tiny pre-prefix install
// base no longer carries them. Frozen alongside legacyCommands.
var renamedCommands = []string{
	"ds-project-plan.md", "ds-modes.md", "ds-review.md",
}

// legacyCommandFiles is every devskills command file a legacy command/prompt dir
// could hold, swept (global only) from ~/.claude/commands, ~/.opencode/commands,
// ~/.codex/prompts. Frozen: neither source list grows again.
var legacyCommandFiles = slices.Concat(legacyCommands, renamedCommands)

// retiredSkills are skill names devskills shipped and later renamed or removed;
// a rename lands the old name here while the new name rides in the catalog.
// Swept as directories from each harness's skills dir, both global and local.
// Unlike the frozen command ledgers, this grows one entry per drop/rename —
// removing or renaming a skill without adding its old name here orphans the dir
// on every user's machine. A test catches the rename half: an old name still
// pointing at a live skill.
//
//	ds-typeset      — removed; off-identity Markdown-typeset utility.
//	ds-senior-mode  — removed; a composite of git/test/step/deslop that had to be
//	                  hand-synced with all four, and drifted from them instead.
var retiredSkills = []string{"ds-typeset", "ds-senior-mode"}
