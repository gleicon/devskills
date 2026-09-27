package sync

import "testing"

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
	tests := []struct {
		name, src, want string
	}{
		{name: "no rule leaves the config alone", src: noSkill, want: noSkill},
		{name: "a fresh config empties", src: freshConfig, want: "{\n}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := allowSkills([]byte(tt.src))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
