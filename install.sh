#!/bin/sh
# devskills installer — fetches the prebuilt CLI binary from GitHub Releases and
# drops it on your PATH. This installs the `devskills` binary only; run
# `devskills install` afterward to sync the skills into your assistants.
#
#   curl -fsSL https://raw.githubusercontent.com/gleicon/devskills/main/install.sh | sh
#
# Overrides (env):
#   VERSION=v0.2.0        pin a release      (default: latest)
#   BINDIR=$HOME/.local/bin  install location (default: /usr/local/bin)
set -eu

REPO="gleicon/devskills"
BIN="devskills"
BINDIR="${BINDIR:-/usr/local/bin}"
VERSION="${VERSION:-}"

fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

have curl || fail "curl is required"
have tar || fail "tar is required"

os=$(uname -s)
case "$os" in
	Linux) os=linux ;;
	Darwin) os=darwin ;;
	*) fail "unsupported OS: $os — use 'go install github.com/$REPO@latest' instead" ;;
esac

arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) fail "unsupported architecture: $arch" ;;
esac

# Resolve the latest tag by following the /releases/latest redirect — no API
# token, no rate limit, no jq. The effective URL ends in /releases/tag/vX.Y.Z.
if [ -z "$VERSION" ]; then
	VERSION=$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
		"https://github.com/$REPO/releases/latest" 2>/dev/null | sed 's#.*/tag/##; s/[[:space:]]//g')
	[ -n "$VERSION" ] || fail "could not determine the latest version; set VERSION=vX.Y.Z"
fi

ver="${VERSION#v}" # goreleaser omits the leading v in asset names
asset="${BIN}_${ver}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$VERSION"

printf 'devskills %s (%s/%s)\n' "$VERSION" "$os" "$arch"

tmp=$(mktemp -d) || fail "could not create a temp dir"
trap 'rm -rf "$tmp"' EXIT INT TERM

curl -fsSL "$base/$asset" -o "$tmp/$asset" || fail "download failed: $asset"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt" || fail "download failed: checksums.txt"

expected=$(awk -v f="$asset" '$2==f {print $1}' "$tmp/checksums.txt")
[ -n "$expected" ] || fail "no checksum listed for $asset"
if have sha256sum; then
	actual=$(sha256sum "$tmp/$asset" | awk '{print $1}')
elif have shasum; then
	actual=$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')
else
	fail "need sha256sum or shasum to verify the download"
fi
[ "$expected" = "$actual" ] || fail "checksum mismatch for $asset"

tar -xzf "$tmp/$asset" -C "$tmp" || fail "could not extract $asset"
[ -f "$tmp/$BIN" ] || fail "binary '$BIN' not found in archive"
chmod +x "$tmp/$BIN"

if [ -d "$BINDIR" ] && [ -w "$BINDIR" ]; then
	mv "$tmp/$BIN" "$BINDIR/$BIN"
elif [ ! -e "$BINDIR" ] && mkdir -p "$BINDIR" 2>/dev/null; then
	mv "$tmp/$BIN" "$BINDIR/$BIN"
elif have sudo; then
	printf 'installing to %s (needs sudo)\n' "$BINDIR"
	sudo mkdir -p "$BINDIR" && sudo mv "$tmp/$BIN" "$BINDIR/$BIN"
else
	fail "$BINDIR is not writable; re-run with BINDIR=\$HOME/.local/bin"
fi

printf 'installed %s to %s\n' "$BIN" "$BINDIR/$BIN"

case ":$PATH:" in
	*":$BINDIR:"*) ;;
	*) printf 'note: %s is not on your PATH — add it to run %s directly\n' "$BINDIR" "$BIN" ;;
esac

printf '\nnext: run "%s install" to sync the skills into your assistants.\n' "$BIN"
