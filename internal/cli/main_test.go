package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/gleicon/devskills/internal/benchtest"
)

func TestMain(m *testing.M) {
	shim, err := benchtest.ShimAssistants()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(shim)
	os.Exit(code)
}
