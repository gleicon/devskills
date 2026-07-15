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

// renamedCommands are pre-prefix and earlier-removed command filenames retired
// across releases:
//
//	frontend.md        -> ui.md (now ds-ui-mode.md)
//	write-a-skill.md   -> write-a-command.md (now ds-write-a-command.md)
//	ds-project-plan.md -> ds-roadmap.md (a plan-generator; left the project-* family)
//	ds-modes.md, ds-review.md -> removed (launcher pickers of 10 modes / 8 reviews
//	                             could never render, so deleted rather than degraded)
//
// Frozen alongside legacyCommands.
var renamedCommands = []string{
	"frontend.md", "write-a-skill.md",
	"bug-review.md", "caveman-lite.md", "caveman-ultra.md", "code-quality-review.md",
	"debug.md", "deslop.md", "doc-quality-review.md", "explore.md", "go-review.md", "grill-me.md",
	"handoff.md", "project-checkpoint.md", "project-map.md", "project-plan.md",
	"project-resume.md", "python-review.md", "quality-gate.md", "rust-review.md",
	"security-review.md", "spec.md", "tdd.md", "test-quality-review.md", "test.md",
	"tiger-style.md", "tldt.md", "ts-review.md", "ui-quality-review.md", "ui.md",
	"verify-this.md", "workflow.md", "write-a-command.md", "zoom-out.md",
	"ds-project-plan.md", "ds-modes.md", "ds-review.md",
}

// legacyCommandFiles is every devskills command file a legacy command/prompt dir
// could hold, swept (global only) from ~/.claude/commands, ~/.opencode/commands,
// ~/.codex/prompts. Frozen: neither source list grows again.
var legacyCommandFiles = slices.Concat(legacyCommands, renamedCommands)

// retiredSkills are skill names devskills shipped and later renamed or removed;
// a rename lands the old name here while the new name rides in the catalog.
// Swept as directories from each harness's skills dir, both global and local.
// Unlike the frozen command ledgers, this grows one entry per drop/rename.
//
//	ds-typeset — removed; off-identity Markdown-typeset utility.
var retiredSkills = []string{"ds-typeset"}
