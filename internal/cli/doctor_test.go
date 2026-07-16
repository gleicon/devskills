package cli

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
)

// fakeLookPath reports the named binaries as present and everything else missing.
func fakeLookPath(present ...string) func(string) (string, error) {
	return func(name string) (string, error) {
		if slices.Contains(present, name) {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
}

func equalCmds(a, b [][]string) bool { return slices.EqualFunc(a, b, slices.Equal) }

func TestRunDoctorCheckReportsStatus(t *testing.T) {
	var out bytes.Buffer
	env := doctorEnv{lookPath: fakeLookPath("osv-scanner", "brew"), goos: "darwin"}
	if err := runDoctor(&out, env, false, false); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "ok") || !strings.Contains(s, "osv-scanner") {
		t.Errorf("expected osv-scanner reported ok:\n%s", s)
	}
	if !strings.Contains(s, "missing") || !strings.Contains(s, "ast-grep") || !strings.Contains(s, "ds-security-review") {
		t.Errorf("expected ast-grep missing with its skill:\n%s", s)
	}
	// Missing but brew present → check surfaces the install command.
	if !strings.Contains(s, "brew install ast-grep") {
		t.Errorf("expected the install hint for ast-grep:\n%s", s)
	}
}

func TestRunDoctorFixRunsInstallers(t *testing.T) {
	var ran [][]string
	env := doctorEnv{
		lookPath: fakeLookPath("brew", "go"), // prerequisites present; tools missing
		goos:     "darwin",
		run: func(name string, args ...string) error {
			ran = append(ran, append([]string{name}, args...))
			return nil
		},
	}
	if err := runDoctor(io.Discard, env, true, false); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"brew", "install", "osv-scanner"},
		{"brew", "install", "ast-grep"},
		{"go", "install", "github.com/gleicon/tldt/cmd/tldt@latest"},
	}
	if !equalCmds(ran, want) {
		t.Errorf("ran = %v, want %v", ran, want)
	}
}

func TestRunDoctorFixDryRunRunsNothing(t *testing.T) {
	var ran [][]string
	var out bytes.Buffer
	env := doctorEnv{
		lookPath: fakeLookPath("brew", "go"),
		goos:     "darwin",
		run: func(name string, args ...string) error {
			ran = append(ran, append([]string{name}, args...))
			return nil
		},
	}
	if err := runDoctor(&out, env, true, true); err != nil {
		t.Fatal(err)
	}
	if len(ran) != 0 {
		t.Errorf("dry-run ran commands: %v", ran)
	}
	if !strings.Contains(out.String(), "would") {
		t.Errorf("expected dry-run to describe commands:\n%s", out.String())
	}
}

func TestRunDoctorFixReportsFailure(t *testing.T) {
	env := doctorEnv{
		lookPath: fakeLookPath("brew", "go"),
		goos:     "darwin",
		run:      func(string, ...string) error { return errors.New("boom") },
	}
	if err := runDoctor(io.Discard, env, true, false); err == nil {
		t.Fatal("want an error when an install command fails")
	}
}

func TestRunDoctorFixNoInstallerIsNotAFailure(t *testing.T) {
	var out bytes.Buffer
	// Linux with no brew/go/npm: nothing is installable, but that's a report,
	// not an error, and no installer runs.
	env := doctorEnv{
		lookPath: fakeLookPath(),
		goos:     "linux",
		run:      func(string, ...string) error { t.Fatal("must not run an installer"); return nil },
	}
	if err := runDoctor(&out, env, true, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "no installer") {
		t.Errorf("expected a 'no installer' note:\n%s", out.String())
	}
}
