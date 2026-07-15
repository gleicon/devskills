package catalog

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// TestValidateRealCatalog runs the contract over the actual on-disk skills/
// tree — the source the binary embeds. go test runs with CWD at the package
// dir, so the repo root is two levels up.
func TestValidateRealCatalog(t *testing.T) {
	if err := Validate(os.DirFS("../..")); err != nil {
		t.Fatalf("skills/ source violates the contract:\n%v", err)
	}
}

func skillFS(name, skillMD string) fstest.MapFS {
	return fstest.MapFS{
		"skills/" + name + "/SKILL.md": {Data: []byte(skillMD)},
	}
}

func TestValidateAcceptsWellFormed(t *testing.T) {
	fsys := skillFS("ds-good", "---\nname: ds-good\ndescription: \"does a thing\"\ndisable-model-invocation: true\n---\n\nBody.\n")
	if err := Validate(fsys); err != nil {
		t.Errorf("well-formed skill rejected: %v", err)
	}
}

func TestValidateRejects(t *testing.T) {
	tests := []struct {
		name    string
		fsys    fstest.MapFS
		wantSub string
	}{
		{
			name:    "no skills dir",
			fsys:    fstest.MapFS{"README.md": {Data: []byte("x")}},
			wantSub: "read skills/",
		},
		{
			name:    "name mismatch",
			fsys:    skillFS("ds-x", "---\nname: ds-y\ndescription: \"d\"\ndisable-model-invocation: true\n---\n"),
			wantSub: `want "ds-x"`,
		},
		{
			name:    "missing description",
			fsys:    skillFS("ds-x", "---\nname: ds-x\ndisable-model-invocation: true\n---\n"),
			wantSub: "missing description",
		},
		{
			name:    "not user-invoked",
			fsys:    skillFS("ds-x", "---\nname: ds-x\ndescription: \"d\"\n---\n"),
			wantSub: "disable-model-invocation: true",
		},
		{
			name:    "no frontmatter",
			fsys:    skillFS("ds-x", "just a body, no fences\n"),
			wantSub: "missing --- frontmatter",
		},
		{
			name:    "non-directory under skills",
			fsys:    fstest.MapFS{"skills/loose.md": {Data: []byte("x")}},
			wantSub: "not a directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.fsys)
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantSub)
			}
		})
	}
}
