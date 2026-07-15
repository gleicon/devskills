package main

import (
	"context"
	"embed"
	"os"

	"github.com/gleicon/devskills/internal/cli"
)

// assets embeds the skill catalog and AGENTS.md profiles at build time — the
// skills are the binary's version, with no runtime file dependency. The
// directive must live in this root package: //go:embed can't reach ../skills
// from internal/. main passes the resulting fs.FS down to commands that need it.
//
//go:embed all:skills all:agents-md
var assets embed.FS

func main() {
	if err := cli.Execute(context.Background(), assets); err != nil {
		os.Exit(1)
	}
}
