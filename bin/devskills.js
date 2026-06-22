#!/usr/bin/env node
"use strict"

const { execSync, spawnSync } = require("child_process")
const path = require("path")
const fs = require("fs")

const DEVSKILLS_DIR = path.join(__dirname, "..")
const INSTALL_SCRIPT = path.join(DEVSKILLS_DIR, "install.sh")
const SETUP_SCRIPT = path.join(DEVSKILLS_DIR, "scripts", "setup.sh")
const UPDATE_SCRIPT = path.join(DEVSKILLS_DIR, "scripts", "update.sh")

const args = process.argv.slice(2)
const cmd = args[0]
const rest = args.slice(1)

function run(script, extraArgs) {
  const result = spawnSync("bash", [script, ...extraArgs], {
    stdio: "inherit",
    cwd: process.cwd()
  })
  if (result.status !== 0) process.exit(result.status ?? 1)
}

function help() {
  console.log(`
devskills — command pack for Claude Code, OpenCode, Cursor, VSCode

Commands:
  install              Install commands to all detected environments
  setup [--lang=<lang>] Configure current project (AGENTS.md baseline + options)
  update               Pull latest and reinstall commands
  list                 List available commands and language profiles
  version              Print the installed devskills version

Language profiles (optional): go, typescript, javascript, rust

Options (pass through to install/setup):
  --skip-external      Skip external tool installation (osv-scanner, tldt, ast-grep)
  --upgrade-deps       Also force-upgrade external tools (update)
  --claude-dir=<path>  Claude config dir (default: $CLAUDE_CONFIG_DIR or ~/.claude)
  --skip-cursor        Skip Cursor rules install (install)
  --skip-vscode        Skip VSCode Copilot install (install)
  --cursor             Install Cursor rules into current project (setup)
  --vscode             Install VSCode Copilot instructions (setup)
  --concise            Add a terse-response directive to AGENTS.md (setup/install)
  --phases             Add phase-aware Insight suggestions to AGENTS.md (setup/install)
  --uninstall          Remove devskills blocks from AGENTS.md/CLAUDE.md (setup)
  --dry-run            Show what would happen without writing files

Examples:
  npx devskills install
  npx devskills setup                       # AGENTS.md baseline only
  npx devskills setup --lang=go --cursor
  npx devskills install --lang=typescript --skip-external
  npx devskills install --claude-dir=~/.config/claude
  npx devskills update --upgrade-deps
`)
}

function list() {
  const commandsDir = path.join(DEVSKILLS_DIR, "commands")
  const langDir = path.join(DEVSKILLS_DIR, "prompts", "language")

  console.log("\nCommands (Claude Code / OpenCode):")
  for (const f of fs.readdirSync(commandsDir).sort()) {
    if (f.endsWith(".md")) {
      console.log(`  /${f.replace(".md", "")}`)
    }
  }

  console.log("\nLanguage profiles:")
  for (const f of fs.readdirSync(langDir).sort()) {
    if (f.endsWith(".md")) {
      console.log(`  ${f.replace(".md", "")}`)
    }
  }

  console.log("\nCursor rules:")
  const cursorDir = path.join(DEVSKILLS_DIR, "cursor", "rules")
  for (const f of fs.readdirSync(cursorDir).sort()) {
    console.log(`  ${f}`)
  }
  console.log()
}

function version() {
  const pkg = JSON.parse(fs.readFileSync(path.join(DEVSKILLS_DIR, "package.json"), "utf8"))
  console.log(`devskills ${pkg.version}`)
}

switch (cmd) {
  case "install":
    run(INSTALL_SCRIPT, rest)
    break
  case "setup":
    run(SETUP_SCRIPT, rest)
    break
  case "update":
    run(UPDATE_SCRIPT, rest)
    break
  case "list":
    list()
    break
  case "version":
  case "--version":
  case "-v":
    version()
    break
  case "help":
  case "--help":
  case "-h":
  case undefined:
    help()
    break
  default:
    console.error(`Unknown command: ${cmd}`)
    help()
    process.exit(1)
}
