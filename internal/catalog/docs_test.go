package catalog

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// go test runs with CWD at the package dir, so the repo root is two levels up.
const repoRoot = "../.."

// The reference docs are not embedded in the binary, so nothing fails when a new
// skill misses one. These tests are what make that a failure instead of a hole
// someone finds after the release.
func TestReferenceDocsNameEverySkill(t *testing.T) {
	headings := headingsOf(t, filepath.Join("docs", "skills.md"))
	readme := readRepoFile(t, "README.md")
	router := readRepoFile(t, filepath.Join("skills", "ds", "SKILL.md"))

	for _, name := range repoSkills(t) {
		// The /ds router indexes every other skill and has no entry of its own.
		if name == "ds" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			// A plain substring match would let /ds-recall-capture satisfy /ds-recall.
			ref := regexp.MustCompile(regexp.QuoteMeta("/"+name) + `([^a-z0-9-]|$)`)
			for _, doc := range []struct{ where, text string }{
				{"a ### heading in docs/skills.md", headings},
				{"README.md", readme},
				{"the /ds router (skills/ds/SKILL.md)", router},
			} {
				if !ref.MatchString(doc.text) {
					t.Errorf("/%s is missing from %s", name, doc.where)
				}
			}
		})
	}
}

// Both files restate the catalog's size, and acceptance_test.go's copy is behind
// the integration tag — drift there stays invisible to a plain go test ./....
func TestSkillCountsStayInSync(t *testing.T) {
	want := strconv.Itoa(len(repoSkills(t)))
	for _, c := range []struct {
		file  string
		count *regexp.Regexp
	}{
		{filepath.Join("internal", "acceptance", "doc.go"), regexp.MustCompile(`(\d+) skills`)},
		{filepath.Join("internal", "acceptance", "acceptance_test.go"), regexp.MustCompile(`skillCount\s*=\s*(\d+)`)},
	} {
		found := c.count.FindAllStringSubmatch(readRepoFile(t, c.file), -1)
		if len(found) == 0 {
			t.Errorf("%s: no skill count found — this guard's pattern has gone stale", c.file)
			continue
		}
		for _, m := range found {
			if m[1] != want {
				t.Errorf("%s: says %s skills, catalog holds %s", c.file, m[1], want)
			}
		}
	}
}

func repoSkills(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(repoRoot, "skills"))
	if err != nil {
		t.Fatalf("read skills/: %v", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		t.Fatal("no skills found under skills/")
	}
	return names
}

// headingsOf keeps only the ### lines: a skill mentioned in passing prose is not
// documented, it needs its own entry.
func headingsOf(t *testing.T, rel string) string {
	t.Helper()
	var headings []string
	for line := range strings.Lines(readRepoFile(t, rel)) {
		if strings.HasPrefix(line, "### ") {
			headings = append(headings, line)
		}
	}
	return strings.Join(headings, "")
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}
