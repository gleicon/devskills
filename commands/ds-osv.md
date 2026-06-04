Scan project dependencies for known vulnerabilities using OSV Scanner.

Queries the Google Open Source Vulnerability (OSV) database against every dependency manifest found in scope. Covers Go, npm, PyPI, Cargo, Maven, RubyGems, NuGet, PHP Composer, and more. Complements `/ds-security-review` (code logic) — this audits what your dependencies brought in, not what you wrote.

If `osv-scanner` is not installed, show installation instructions and stop.

## Arguments

- No args: scan the current directory recursively.
- `<path>`: scan a specific directory or lockfile.
- `--fix`: after reporting, bump vulnerable **direct** dependencies to the minimum fixed version by driving the ecosystem's own package manager (see Process step 6). Transitive-only findings are reported but not auto-fixed — they require the intermediate package to release a patch.
- `--verbose`: expand LOW / informational findings, which are otherwise summarized as a count only.

## Process

1. Check for `osv-scanner` binary. If missing:
   ```
   Install osv-scanner:
     macOS:  brew install osv-scanner
     Go:     go install github.com/google/osv-scanner/cmd/osv-scanner@latest
     Other:  https://github.com/google/osv-scanner/releases
   ```
   Stop here. Do not proceed without the binary.

2. Run the scan:
   ```bash
   osv-scanner --recursive --format json <path-or-.>
   ```

3. Parse JSON output. Group findings:
   - **CRITICAL / HIGH** — report first, always actionable
   - **MEDIUM** — report with context
   - **LOW / informational** — summarize count only unless `--verbose`

4. Classify each finding as **direct** or **transitive**. A finding is *direct* if its package appears in the manifest's own declared dependency list — e.g. `go.mod` `require`, `package.json` `dependencies`/`devDependencies`, `Cargo.toml` `[dependencies]`/`[dev-dependencies]`, top-level entries in `requirements.txt`. Otherwise it is *transitive*. Determine this by reading the manifest yourself; do not rely on a scanner field, which is not consistently populated across ecosystems.

5. For each finding report:
   - Package name + current version
   - Vulnerability ID (OSV ID and CVE alias if present)
   - Severity + CVSS score if available
   - **Fixed version** (the minimum version that resolves it)
   - Whether it's a direct or transitive dependency (from step 4)
   - One-line description of the vulnerability class

6. With `--fix`: for **direct** dependencies (per step 4) where a fixed version exists, bump to the **minimum** fixed version by driving the ecosystem's native package-manager command — never by hand-editing manifest text — so the manifest and lockfile stay consistent. Use the upgrade primitive native to the manifest's ecosystem, for example:
   - Go: `go get <pkg>@<version>`
   - npm: `npm install <pkg>@<version>`
   - Cargo: `cargo update -p <pkg> --precise <version>`
   - …and the equivalent for any other detected ecosystem (PyPI, Maven, RubyGems, NuGet, Composer).

   Where no package-manager primitive cleanly applies (e.g. a raw `requirements.txt` with no pip-tools), fall back to editing the manifest and instruct the user to regenerate the lockfile. Note each change. Transitive-only findings are reported, not auto-fixed.

## Output

```
OSV scan: <N> vulnerabilities found (<C> critical, <H> high, <M> medium, <L> low)

CRITICAL / HIGH
  <package>@<version>  <OSV-ID> / <CVE>
  Severity: <score>  Fixed: <version>  Dependency: direct|transitive
  <one-line class: e.g. "arbitrary code execution via malformed input">

MEDIUM
  <package>@<version>  <OSV-ID>
  ...

LOW/INFO: <N> findings — run with --verbose to expand.

Next steps:
  Re-run your build and tests to confirm each bump is compatible.
  Re-run /ds-osv to confirm findings resolved.
```

If zero findings: report clean with ecosystem coverage summary (which manifests were scanned).

## Rules

- Apply fixes by driving the ecosystem's package manager, not by hand-editing files. Never hand-edit a lockfile; let the package manager update the manifest and lockfile together so they stay consistent.
- Do not bump a version beyond the stated minimum fixed version — pick the minimum that resolves the CVE.
- If a finding has no fixed version available, report it clearly and suggest tracking the upstream issue.
- Transitive-only findings with no direct-dependency path to a fix should be flagged for the user to escalate to the upstream package.
