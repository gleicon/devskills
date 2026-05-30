# Publishing and Distribution

This document covers how to distribute devskills: from a personal GitHub install to a public npm package.

---

## Option 1: Install directly from GitHub (no npm)

No publish step required. Users clone or curl-install from the repo.

### For users

```bash
# Clone and install
git clone https://github.com/gleicon/devskills.git ~/.devskills
cd ~/.devskills
./install.sh

# Or one-liner (skip external tools)
curl -fsSL https://raw.githubusercontent.com/gleicon/devskills/main/install.sh | bash -s -- --skip-external
```

### For updates

```bash
cd ~/.devskills && ./scripts/update.sh
```

### Per-project language profile

```bash
cd your-project
~/.devskills/scripts/setup.sh --lang=go --cursor
```

---

## Option 2: Publish to npm

Allows `npx devskills install` without cloning.

### Prerequisites

- npm account: https://www.npmjs.com/signup
- Logged in: `npm login`
- Package name available: `npm view devskills` (pick another name if taken)

### Steps

1. **Set the package name** in `package.json`:
   ```json
   { "name": "devskills" }
   ```
   If `devskills` is taken, use a scoped name:
   ```json
   { "name": "@your-npm-username/devskills" }
   ```

2. **Set the repository URL** in `package.json`:
   ```json
   {
     "repository": {
       "type": "git",
       "url": "https://github.com/gleicon/devskills.git"
     }
   }
   ```

3. **Bump the version** before each publish:
   ```bash
   npm version patch   # 0.1.0 → 0.1.1 (bug fixes)
   npm version minor   # 0.1.0 → 0.2.0 (new skills)
   npm version major   # 0.1.0 → 1.0.0 (breaking changes)
   ```

4. **Publish**:
   ```bash
   npm publish
   # or for scoped packages:
   npm publish --access public
   ```

5. **Verify**:
   ```bash
   npx devskills list
   npx devskills install --dry-run
   ```

### What gets published

The `files` array in `package.json` controls this. Current value:
```json
"files": ["bin/", "claude/", "opencode/", "cursor/", "vscode/", "prompts/", "scripts/", "docs/", "install.sh", "README.md"]
```

Node modules and `.git/` are excluded automatically.

### Verify before publishing

```bash
npm pack --dry-run    # list files that would be published
npm pack             # create .tgz, inspect locally
tar tzf devskills-*.tgz
```

---

## Option 3: GitHub Releases (binary-style distribution)

Useful for teams that do not use npm but want versioned installs.

1. Tag a release:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. Create a GitHub release via `gh`:
   ```bash
   gh release create v0.1.0 --title "v0.1.0" --notes "Initial release" --generate-notes
   ```

3. Users install from a specific tag:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/gleicon/devskills/v0.1.0/install.sh | bash
   ```

---

## Updating Published Skills

### After adding a new skill file

1. Copy the file to `opencode/commands/` if it belongs there.
2. Update `README.md` skills table.
3. Bump version: `npm version patch`.
4. `npm publish`.
5. `git push && git push --tags`.

### Keeping external tool references current

The install script references external tools by their latest published version. When GSD, RTK, or tldt release breaking changes:

1. Update the relevant install function in `install.sh`.
2. Update the command file that references the tool.
3. Update `docs/` if behavior changed.
4. Bump `npm version minor` (behavior change, not just bug fix).

---

## CI: GitHub Actions (optional)

Create `.github/workflows/publish.yml` to automate npm publish on tag push:

```yaml
name: publish

on:
  push:
    tags: ["v*"]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          registry-url: https://registry.npmjs.org
      - run: npm publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

Store `NPM_TOKEN` in GitHub repo Settings → Secrets → Actions.

---

## Naming a Scoped Package

If publishing under a GitHub org or personal namespace:

```json
{ "name": "@gleicon/devskills" }
```

Install command becomes:
```bash
npx @gleicon/devskills install
```

Scoped packages default to private. Always pass `--access public` when publishing a public scoped package.

---

## Checklist Before First Publish

- [ ] `package.json` has correct `name`, `version`, `repository`
- [ ] `README.md` has correct install URL
- [ ] `install.sh` references correct GitHub raw URL for curl installs
- [ ] `npm pack --dry-run` shows only intended files
- [ ] `npx devskills list` works from a fresh clone
- [ ] `./install.sh --dry-run --skip-external` completes without errors
- [ ] Git tag created and pushed
