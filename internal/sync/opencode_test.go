package sync

import (
	"path/filepath"
	"strings"
	"testing"
)

// Configs whose rule survives an install/uninstall round trip byte for byte.
var (
	noPermission = `{
  "model": "x"
}
`
	noSkill = `{
  "permission": {
    "edit": "ask"
  }
}
`
	commented = `{
  // model picked by hand
  "model": "x",
  "permission": {
    "skill": {
      "*": "ask",
    },
  },
}
`
)

func TestDenySkills(t *testing.T) {
	tests := []struct {
		name, src, want string
	}{
		{name: "empty file gets a fresh config", src: " \n", want: freshConfig},
		{
			name: "missing permission is added",
			src:  noPermission,
			want: `{
  "model": "x",
  "permission": {"skill": {"ds-*": "deny"}}
}
`,
		},
		{
			name: "missing skill is added",
			src:  noSkill,
			want: `{
  "permission": {
    "edit": "ask",
    "skill": {"ds-*": "deny"}
  }
}
`,
		},
		{
			name: "a skill string becomes the catch-all rule",
			src:  `{"permission": {"skill": "ask"}}`,
			want: `{"permission": {"skill": {"*": "ask", "ds-*": "deny"}}}`,
		},
		{
			name: "an earlier ds-* rule moves last so it wins",
			src: `{
  "permission": {
    "skill": {
      "ds-*": "allow",
      "*": "ask"
    }
  }
}
`,
			want: `{
  "permission": {
    "skill": {
      "*": "ask",
      "ds-*": "deny"
    }
  }
}
`,
		},
		{
			name: "comments and trailing commas are kept",
			src:  commented,
			want: `{
  // model picked by hand
  "model": "x",
  "permission": {
    "skill": {
      "*": "ask",
      "ds-*": "deny",
    },
  },
}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := denySkills([]byte(tt.src))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
			again, err := denySkills(got)
			if err != nil {
				t.Fatal(err)
			}
			if string(again) != string(got) {
				t.Errorf("a second install changed the config:\n%s", again)
			}
		})
	}
}

func TestDenySkillsRejectsUnexpectedShapes(t *testing.T) {
	for _, src := range []string{
		`[]`,
		`{"permission": 3}`,
		`{"permission": {"skill": ["x"]}}`,
		`{"permission": `,
	} {
		if _, err := denySkills([]byte(src)); err == nil {
			t.Errorf("%s: want an error, got none", src)
		}
	}
}

func TestAllowSkillsUndoesDenySkills(t *testing.T) {
	for _, src := range []string{noPermission, noSkill, commented} {
		denied, err := denySkills([]byte(src))
		if err != nil {
			t.Fatal(err)
		}
		got, err := allowSkills(denied)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != src {
			t.Errorf("round trip changed the config:\n%s\nwant:\n%s", got, src)
		}
	}
}

func TestAllowSkills(t *testing.T) {
	got, err := allowSkills([]byte(noSkill))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != noSkill {
		t.Errorf("a config without the rule changed:\n%s", got)
	}

	got, err = allowSkills([]byte(freshConfig))
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("a fresh config should empty to nil so the file is deleted, got:\n%s", got)
	}
}

func TestOpenCodeConfigLifecycle(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")
	config := filepath.Join(filepath.Dir(skillsDir), "opencode.json")
	e := New(fakeCatalog())
	target := Target{SkillsDir: skillsDir, OpenCode: true}

	p, err := e.Plan(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertFile(t, config, freshConfig)

	if p, err = e.Plan(target); err != nil {
		t.Fatal(err)
	}
	if p.Config != nil {
		t.Errorf("a second install plans a config edit: %+v", p.Config)
	}

	if p, err = e.UninstallPlan(target); err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, config)
}

func TestOpenCodeConfigKeepsUserSettings(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")
	config := filepath.Join(filepath.Dir(skillsDir), "opencode.jsonc")
	mkfile(t, config, commented)
	e := New(fakeCatalog())
	target := Target{SkillsDir: skillsDir, OpenCode: true}

	p, err := e.Plan(target)
	if err != nil {
		t.Fatal(err)
	}
	if p.Config == nil || p.Config.Path != config {
		t.Fatalf("Config = %+v, want an edit of the existing %s", p.Config, config)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, filepath.Join(filepath.Dir(skillsDir), "opencode.json"))

	if p, err = e.UninstallPlan(target); err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertFile(t, config, commented)
}

func TestOpenCodeConfigBrokenFailsAtPlan(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")
	config := filepath.Join(filepath.Dir(skillsDir), "opencode.json")
	mkfile(t, config, `{"permission": `)

	_, err := New(fakeCatalog()).Plan(Target{SkillsDir: skillsDir, OpenCode: true})
	if err == nil || !strings.Contains(err.Error(), config) {
		t.Errorf("want a plan error naming %s, got %v", config, err)
	}
}

func TestPlanWithoutOpenCodeTouchesNoConfig(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")
	e := New(fakeCatalog())
	p, err := e.Plan(Target{SkillsDir: skillsDir})
	if err != nil {
		t.Fatal(err)
	}
	if p.Config != nil {
		t.Errorf("Config = %+v, want none", p.Config)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, filepath.Join(filepath.Dir(skillsDir), "opencode.json"))
}
