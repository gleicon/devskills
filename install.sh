#!/usr/bin/env bash
set -euo pipefail

DEVSKILLS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_CONFIG_DIR="${CLAUDE_CONFIG_DIR:-${HOME}/.claude}"
# OpenCode reads global skills from $XDG_CONFIG_HOME/opencode/skills (default
# ~/.config/opencode/skills). The legacy ~/.opencode/commands dir is purged.
OPENCODE_CONFIG_DIR="${XDG_CONFIG_HOME:-${HOME}/.config}/opencode"
OPENCODE_SKILLS_DIR="${OPENCODE_CONFIG_DIR}/skills"
OPENCODE_COMMANDS_DIR="${HOME}/.opencode/commands"
# Codex honors CODEX_HOME (default ~/.codex). Skills live in skills/ (codex
# reads $CODEX_HOME/skills, marker-managed); the legacy prompts/ dir is purged.
CODEX_HOME_DIR="${CODEX_HOME:-${HOME}/.codex}"
CODEX_SKILLS_DIR="${CODEX_HOME_DIR}/skills"
CODEX_COMMANDS_DIR="${CODEX_HOME_DIR}/prompts"

# Parsed from package.json (bash-only, no node) for the agy plugin manifest.
DEVSKILLS_VERSION="$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "${DEVSKILLS_DIR}/package.json" | head -1)"

log() { printf '[devskills] %s\n' "$1"; }
warn() { printf '[devskills] WARN: %s\n' "$1" >&2; }

# Shared tldt logic (depends on log/warn above and DRY_RUN below).
# shellcheck source=scripts/lib/external-tools.sh
source "${DEVSKILLS_DIR}/scripts/lib/external-tools.sh"

# ------------------------------------------------------------
# Arguments
# ------------------------------------------------------------

LANG_PROFILE=""
SKIP_EXTERNAL=0
SKIP_CLAUDE=0
SKIP_OPENCODE=0
SKIP_CODEX=0
SKIP_AGY=0
CONCISE=0
PHASES=0
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --lang=*) LANG_PROFILE="${arg#--lang=}" ;;
    --claude-dir=*) CLAUDE_CONFIG_DIR="${arg#--claude-dir=}" ;;
    --skip-external) SKIP_EXTERNAL=1 ;;
    --skip-claude) SKIP_CLAUDE=1 ;;
    --skip-opencode) SKIP_OPENCODE=1 ;;
    --skip-codex) SKIP_CODEX=1 ;;
    --skip-agy) SKIP_AGY=1 ;;
    --concise) CONCISE=1 ;;
    --phases) PHASES=1 ;;
    --dry-run) DRY_RUN=1 ;;
    --help|-h)
      echo "Usage: install.sh [--lang=go|typescript|javascript|rust|python|java|zig] [--claude-dir=PATH] [--skip-external] [--skip-claude] [--skip-opencode] [--skip-codex] [--skip-agy] [--concise] [--phases] [--dry-run]"
      echo ""
      echo "  --lang=<profile>    Language profile to write: go|typescript|javascript|rust|python|java|zig"
      echo "  --claude-dir=PATH   Claude config dir (default: \$CLAUDE_CONFIG_DIR or \$HOME/.claude)"
      echo "  --skip-external     Skip external tool installation (osv-scanner, tldt, ast-grep)"
      echo "  --skip-claude       Skip Claude Code skills install"
      echo "  --skip-opencode     Skip OpenCode skills install"
      echo "  --skip-codex        Skip Codex skills install"
      echo "  --skip-agy          Skip Antigravity (agy) plugin install"
      echo "  --concise           Add a terse-response directive to AGENTS.md (with --lang)"
      echo "  --phases            Add phase-aware Insight suggestions to AGENTS.md (with --lang)"
      echo "  --dry-run           Show what would happen, write nothing"
      exit 0
      ;;
  esac
done

# Expand leading ~ in --claude-dir value.
# Quote the strip pattern: an unquoted ~/ undergoes tilde expansion itself,
# strips nothing, and yields "$HOME/~/.claude".
case "$CLAUDE_CONFIG_DIR" in
  "~") CLAUDE_CONFIG_DIR="${HOME}" ;;
  "~/"*) CLAUDE_CONFIG_DIR="${HOME}/${CLAUDE_CONFIG_DIR#"~/"}" ;;
esac
CLAUDE_COMMANDS_DIR="${CLAUDE_CONFIG_DIR}/commands"
CLAUDE_SKILLS_DIR="${CLAUDE_CONFIG_DIR}/skills"
DEVSKILLS_SKILLS_DIR="${DEVSKILLS_DIR}/skills"

# Auto-skip project-local writes when run from inside the devskills source
# repo — otherwise --lang writes contributor files into the repo itself.
case "${PWD}/" in
  "${DEVSKILLS_DIR}"/*)
    if [ -n "$LANG_PROFILE" ]; then
      warn "Running inside the devskills source repo; ignoring --lang to avoid writing CLAUDE.md into the repo."
      LANG_PROFILE=""
    fi
    ;;
esac

# AGENTS.md is only written when --lang is given (see install_lang_profile).
# Flag --concise used without --lang so it isn't a silent no-op.
if [ -z "$LANG_PROFILE" ] && [ "$CONCISE" -eq 1 ]; then
  warn "--concise applies with --lang; nothing written to AGENTS.md. Use scripts/setup.sh for a baseline-only project."
fi

# Validate --lang up front, before any install side effects: a bad profile
# should fail fast, not after tldt is already installed.
if [ -n "$LANG_PROFILE" ] && [ ! -f "${DEVSKILLS_DIR}/agents-md/language/${LANG_PROFILE}.md" ]; then
  warn "No language profile for '${LANG_PROFILE}'. Available: go, typescript, javascript, rust, python, java, zig"
  exit 1
fi

# ------------------------------------------------------------
# Helpers
# ------------------------------------------------------------

install_file() {
  local src="$1"
  local dst="$2"
  if [ "$DRY_RUN" -eq 1 ]; then
    log "[dry] would install $src -> $dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
  log "installed $dst"
}

# Install one skill directory (SKILL.md + any companion files) as a plain copy.
# rm -rf before copy keeps it idempotent and avoids cp -R nesting a stale tree
# inside itself on a re-run; a companion dropped upstream doesn't linger.
install_skill_dir() {
  local src="$1"  # .../skills/<name>
  local dst="$2"  # <target>/skills/<name>
  if [ "$DRY_RUN" -eq 1 ]; then
    log "[dry] would install skill $(basename "$src") -> $dst"
    return
  fi
  rm -rf "$dst"
  mkdir -p "$(dirname "$dst")"
  cp -R "$src" "$dst"
  log "installed skill $dst"
}

# Generate the Codex per-skill sidecar. allow_implicit_invocation: false keeps
# the skill out of the model's default context — user-invoked only via $name,
# never auto-fired — which is devskills' whole identity. Codex documents this
# field; it is generated at stage time, never committed to the repo.
emit_codex_sidecar() {
  local skill_dir="$1"
  [ "$DRY_RUN" -eq 1 ] && return
  mkdir -p "${skill_dir}/agents"
  printf 'policy:\n  allow_implicit_invocation: false\n' > "${skill_dir}/agents/openai.yaml"
}

# Copy every skill dir under skills/ into a target skills root. With sidecar=codex
# each installed skill also gets its generated agents/openai.yaml.
install_skills_to() {
  local target_root="$1" sidecar="${2:-}"
  [ "$DRY_RUN" -eq 1 ] || mkdir -p "$target_root"
  local d name
  for d in "${DEVSKILLS_SKILLS_DIR}/"*/; do
    name="$(basename "$d")"
    install_skill_dir "${d%/}" "${target_root}/${name}"
    if [ "$sidecar" = "codex" ]; then
      emit_codex_sidecar "${target_root}/${name}"
    fi
  done
  return 0
}

# Commands removed or renamed in past releases. install only ever copies, so
# without this the old name lingers next to its replacement forever (e.g. after
# update.sh). Remove the known stale files from a target commands dir; only
# touches names devskills itself shipped, never user-authored commands.
#   frontend.md     -> ui.md (now ds-ui-mode.md)
#   write-a-skill.md -> write-a-command.md (now ds-write-a-command.md)
#   ds-project-plan.md -> ds-roadmap.md (a plan-generator, not `.project` memory —
#                         so it left the project-* family; a post-prefix rename)
#   ds-modes.md, ds-review.md -> removed (no replacement): the host question UI
#                         caps a picker at four options, but these launchers had
#                         ten modes / eight reviews, so the picker could never
#                         render — deleted rather than degraded to a prose menu.
# Every command was namespaced with a `ds-` prefix (modes also gain a `-mode`
# suffix); the pre-prefix filenames below are retired here, plus the one
# post-prefix rename and the two removed launchers above. New names all carry
# the `ds-` prefix and none collide with the stale names being removed.
RENAMED_COMMANDS=(
  frontend.md write-a-skill.md
  bug-review.md caveman-lite.md caveman-ultra.md code-quality-review.md
  debug.md deslop.md doc-quality-review.md explore.md go-review.md grill-me.md
  handoff.md project-checkpoint.md project-map.md project-plan.md
  project-resume.md python-review.md quality-gate.md rust-review.md
  security-review.md spec.md tdd.md test-quality-review.md test.md
  tiger-style.md tldt.md ts-review.md ui-quality-review.md ui.md
  verify-this.md workflow.md write-a-command.md zoom-out.md
  ds-project-plan.md ds-modes.md ds-review.md
)

# The frozen pre-overhaul command catalog: the 51 ds-*.md files that shipped as
# commands before the skills migration. install now ships skills/ only, so any of
# these left in a harness's legacy command/prompt dir is stale and shadows or
# duplicates the new skill — remove it on upgrade. This set is history and never
# grows again (no new commands are authored). Combined with RENAMED_COMMANDS
# (pre-prefix and earlier-removed names) it covers every devskills file a legacy
# command dir could hold; it never names a user-authored command.
LEGACY_COMMANDS=(
  ds-architecture-plan.md ds-blueprint.md ds-bug-review.md
  ds-caveman-lite-mode.md ds-caveman-ultra-mode.md ds-code-quality-review.md
  ds-code-review.md ds-comment-review.md ds-data-mode.md ds-data-review.md
  ds-debug.md ds-deslop.md ds-doc-quality-review.md ds-explore.md ds-git-mode.md
  ds-go-review.md ds-grill-me.md ds-handoff.md ds-java-review.md
  ds-notebook-review.md ds-osv.md ds-perf-plan.md ds-project-checkpoint.md
  ds-project-config.md ds-project-map.md ds-project-resume.md ds-python-review.md
  ds-quality-gate-mode.md ds-recall-capture.md ds-recall-setup.md ds-recall.md
  ds-roadmap.md ds-rust-review.md ds-security-review.md ds-senior-mode.md
  ds-spec.md ds-step-mode.md ds-tdd-mode.md ds-test-mode.md
  ds-test-quality-review.md ds-tiger-style-mode.md ds-tldt.md ds-ts-review.md
  ds-typeset.md ds-ui-mode.md ds-ui-quality-review.md ds-verify-this.md
  ds-workflow.md ds-write-a-command.md ds-zig-review.md ds-zoom-out.md
)

# Full purge for a harness migrated to skills: remove the whole legacy command
# catalog plus historically renamed names from its old command/prompt dir.
purge_legacy_commands() {
  local dir="$1" name
  [ -d "$dir" ] || return 0
  for name in "${LEGACY_COMMANDS[@]}" "${RENAMED_COMMANDS[@]}"; do
    [ -f "${dir}/${name}" ] || continue
    if [ "$DRY_RUN" -eq 1 ]; then
      log "[dry] would remove legacy command ${dir}/${name}"
    else
      rm -f "${dir}/${name}"
      log "removed legacy command ${dir}/${name}"
    fi
  done
}

# ------------------------------------------------------------
# Claude Code commands
# ------------------------------------------------------------

install_claude() {
  if command -v claude &>/dev/null || [ -d "${CLAUDE_CONFIG_DIR}" ]; then
    log "Installing Claude Code skills to ${CLAUDE_SKILLS_DIR}"
    install_skills_to "${CLAUDE_SKILLS_DIR}"
    # Skills replace commands: drop any devskills ds-*.md left in the old dir.
    purge_legacy_commands "${CLAUDE_COMMANDS_DIR}"
  else
    warn "Claude Code not detected. Skipping. Install from https://claude.ai/code"
  fi
}

# ------------------------------------------------------------
# OpenCode skills
# ------------------------------------------------------------

# OpenCode discovers skills from ~/.config/opencode/skills (and the shared
# ~/.claude, ~/.agents dirs). It ignores unknown frontmatter, so the single
# SKILL.md installs as-is. Note: OpenCode skills are model-invocable — it has no
# per-skill user-only flag — so disable-model-invocation is silently ignored here.
install_opencode() {
  if command -v opencode &>/dev/null || [ -d "${OPENCODE_CONFIG_DIR}" ] || [ -d "${HOME}/.opencode" ]; then
    log "Installing OpenCode skills to ${OPENCODE_SKILLS_DIR}"
    install_skills_to "${OPENCODE_SKILLS_DIR}"
    purge_legacy_commands "${OPENCODE_COMMANDS_DIR}"
  else
    warn "OpenCode not detected. Skipping."
  fi
}

# ------------------------------------------------------------
# OpenAI Codex skills
# ------------------------------------------------------------

# Codex reads project AGENTS.md natively (built by setup.sh/profile.sh), so only
# the skill surface needs installing. Skills live in ${CODEX_HOME}/skills, each
# with a generated agents/openai.yaml pinning allow_implicit_invocation: false
# (user-invoked via $name). Custom prompts were removed in Codex 0.117.0; the old
# prompts/ dir is purged. Global target with no opt-out beyond --skip-codex.
install_codex() {
  if command -v codex &>/dev/null || [ -d "${CODEX_HOME_DIR}" ]; then
    log "Installing Codex skills to ${CODEX_SKILLS_DIR}"
    install_skills_to "${CODEX_SKILLS_DIR}" codex
    purge_legacy_commands "${CODEX_COMMANDS_DIR}"
  else
    warn "Codex not detected. Skipping. Install from https://developers.openai.com/codex"
  fi
}

# ------------------------------------------------------------
# Antigravity (agy) plugin
# ------------------------------------------------------------

# agy has no stable skills dir to copy into; it ingests a plugin directory
# (plugin.json + skills/) via `agy plugin install`, which copies it into agy's
# own store. Stage the whole skills tree in a temp plugin dir and hand it over.
install_agy() {
  if ! command -v agy &>/dev/null; then
    warn "Antigravity (agy) not detected. Skipping."
    return
  fi
  log "Installing Antigravity plugin via agy plugin install"
  if [ "$DRY_RUN" -eq 1 ]; then
    log "[dry] would stage devskills plugin and run: agy plugin install <staged dir>"
    return
  fi
  local stage
  stage="$(mktemp -d "${TMPDIR:-/tmp}/devskills-agy.XXXXXX")"
  printf '{\n  "name": "devskills",\n  "version": "%s",\n  "description": "devskills user-invoked skills"\n}\n' \
    "${DEVSKILLS_VERSION:-0.0.0}" > "${stage}/plugin.json"
  cp -R "${DEVSKILLS_SKILLS_DIR}" "${stage}/skills"
  if agy plugin install "$stage" >/dev/null 2>&1; then
    log "installed devskills plugin into agy"
  else
    warn "agy plugin install failed; skills not installed for Antigravity"
  fi
  rm -rf "$stage"
}

# ------------------------------------------------------------
# Language profile
# ------------------------------------------------------------

install_lang_profile() {
  local lang="$1"
  log "Writing AGENTS.md baseline${lang:+ + ${lang} profile} to ${PWD}"

  # shellcheck source=scripts/lib/profile.sh
  source "${DEVSKILLS_DIR}/scripts/lib/profile.sh"
  devskills_apply "${DEVSKILLS_DIR}/agents-md" "$PWD" "$DRY_RUN" "$lang" "$CONCISE" "$PHASES"
}

# ------------------------------------------------------------
# Main
# ------------------------------------------------------------

log "devskills installer"
log "source: ${DEVSKILLS_DIR}"

if [ "$SKIP_CLAUDE" -eq 0 ]; then install_claude; else log "Skipping Claude Code (--skip-claude)"; fi
if [ "$SKIP_OPENCODE" -eq 0 ]; then install_opencode; else log "Skipping OpenCode (--skip-opencode)"; fi
if [ "$SKIP_CODEX" -eq 0 ]; then install_codex; else log "Skipping Codex (--skip-codex)"; fi
if [ "$SKIP_AGY" -eq 0 ]; then install_agy; else log "Skipping Antigravity (--skip-agy)"; fi

if [ "$SKIP_EXTERNAL" -eq 0 ]; then
  log "Installing external tools..."
  devskills_osv install
  devskills_tldt install
  devskills_astgrep install
else
  log "Skipping external tools (--skip-external)"
fi

# RTK was removed after an upstream supply-chain compromise. Warn (never delete)
# if an earlier devskills install left it behind. Runs even with --skip-external,
# since skipping the install does not make an already-present compromised binary safe.
devskills_rtk_remediate

if [ -n "$LANG_PROFILE" ]; then
  install_lang_profile "$LANG_PROFILE"
fi

log ""
log "Done. Verify a skill is installed, e.g. ds-tiger-style-mode:"
log "  Claude Code:  /ds-tiger-style-mode"
log "  OpenCode:     ds-tiger-style-mode (via the skill tool)"
log "  Codex:        \$ds-tiger-style-mode"
log "  Antigravity:  /ds-tiger-style-mode"
log "  osv-scanner --version         — supply-chain vulnerability scanner"
log "  tldt --version                — text summarizer"
log "  ast-grep --version            — structural code search (enhances /ds-security-review)"
log ""
log "Set language profile in any project:"
log "  ./install.sh --lang=go"
log "  ./install.sh --lang=typescript"
