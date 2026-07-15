package main

import (
	"context"
	"os"

	"github.com/gleicon/devskills/internal/cli"
)

func main() {
	if err := cli.Execute(context.Background()); err != nil {
		os.Exit(1)
	}
}
