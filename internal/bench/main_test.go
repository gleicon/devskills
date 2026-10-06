package bench

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/gleicon/devskills/internal/benchtest"
)

// fakePWDEnv turns the test binary into a fake opencode that reports whether
// its inherited PWD names its working directory.
const fakePWDEnv = "BENCH_FAKE_OPENCODE_PWD"

func TestMain(m *testing.M) {
	if os.Getenv(fakePWDEnv) == "1" {
		os.Exit(fakeOpenCodePWD())
	}
	shim, err := benchtest.ShimAssistants()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(shim)
	os.Exit(code)
}

func fakeOpenCodePWD() int {
	text := "PWD is " + os.Getenv("PWD")
	cwd, err := os.Stat(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if pwd, err := os.Stat(os.Getenv("PWD")); err == nil && os.SameFile(pwd, cwd) {
		text = "PWD is cwd"
	}
	line, err := json.Marshal(map[string]any{"type": "text", "part": map[string]string{"text": text}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("%s\n%s\n", line, `{"type":"step_finish","part":{"cost":0,"tokens":{"input":1,"output":1,"reasoning":0,"cache":{"read":0,"write":0}}}}`)
	return 0
}
