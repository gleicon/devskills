// Package acceptance holds end-to-end acceptance tests that build the devskills
// binary and drive it against a throwaway sandbox. The tests live behind the
// `integration` build tag; this file keeps the package buildable without it.
//
// Run them with:
//
//	go test -tags integration ./internal/acceptance/
//
// # Acceptance criteria
//
// The overhaul's acceptance criteria are AC-1…AC-18. This is their canonical
// record: the spec that first stated them was archived once the overhaul
// shipped, and .project/ is git-ignored either way.
//
// Verified by the integration tests in this package (TestInstall/Init/Doctor/Version):
//
//	AC-4   install --all: 50 skills per harness; a ds-* legacy command is purged
//	       with no backup; a non-ds legacy command is left untouched (the ledger
//	       only sweeps names provably ours); a retired skill dir (ds-typeset) is
//	       pruned.
//	AC-5   install --dry-run leaves $HOME and the project byte-for-byte unchanged.
//	AC-6   every installed Codex skill carries agents/openai.yaml pinning
//	       allow_implicit_invocation: false.
//	AC-7   a second install produces an identical tree (idempotent; no new backups).
//	AC-13  init --lang go,typescript writes the base + language blocks and a
//	       CLAUDE.md @AGENTS.md import, preserves pre-existing user content,
//	       re-runs as a no-op, and uninstalls without dropping user content.
//	AC-17  doctor lists each external tool + owning skill; check-only changes
//	       nothing; --fix --dry-run prints install commands but never installs.
//	AC-18  version prints the ldflags-injected build stamp.
//
// Verified elsewhere (unit tests, static checks, or skill-prompt review):
//
//	AC-1   skills/ holds 50 skills and none of the removed/renamed names — the
//	       count by catalog.TestSkillCountsStayInSync, the names static.
//	AC-2   every SKILL.md: name==dir, a description, disable-model-invocation:true
//	       — the internal/catalog validator (go test ./...).
//	AC-3   the /ds router, README.md and docs/skills.md each name every other
//	       skill — catalog.TestReferenceDocsNameEverySkill.
//	AC-8   go test ./... passes and golangci-lint run is clean — the gate itself.
//	AC-9   retired — ds-project-compact is gone; state.md is bounded by its format
//	       and has nothing to archive. Numbering kept so AC-n stays stable.
//	AC-10  retired — .project/archive/ no longer exists (see AC-9).
//	AC-11  no live references to removed/renamed names in README/docs/agents-md — grep.
//	AC-12  commands.md/PUBLISHING.md/scripts/bin/package.json are absent; skills.md
//	       and install.sh are present — static.
//	AC-15  doc-quality-review drops --comments; comment-review flags planning-artifact
//	       cruft — skill content.
//	AC-16  go/python/java reviews carry a version companion; zig/rust/ts are
//	       single-file — static.
package acceptance
