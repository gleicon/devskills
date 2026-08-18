package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"strings"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name    string
		ldflags string
		info    *debug.BuildInfo
		want    string
	}{
		{"ldflags tag wins", "v1.2.3", nil, "v1.2.3"},
		{"falls back to build info", "dev", &debug.BuildInfo{Main: debug.Module{Version: "v0.2.0"}}, "v0.2.0"},
		{"ignores devel build info", "dev", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "dev"},
		{"no build info", "dev", nil, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.ldflags, tt.info); got != tt.want {
				t.Errorf("resolveVersion(%q, ...) = %q, want %q", tt.ldflags, got, tt.want)
			}
		})
	}
}

func TestReleaseNums(t *testing.T) {
	tests := []struct {
		in   string
		want [3]int
		ok   bool
	}{
		{"v1.2.3", [3]int{1, 2, 3}, true},
		{"0.6.0", [3]int{0, 6, 0}, true},
		{"dev", [3]int{}, false},
		{"v0.6.1-0.20260817120000-abcdef123456", [3]int{}, false},
		{"1.2", [3]int{}, false},
		{"1.2.3.4", [3]int{}, false},
		{"", [3]int{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, ok := releaseNums(tt.in)
			if ok != tt.ok || got != tt.want {
				t.Errorf("releaseNums(%q) = %v, %v; want %v, %v", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestReleaseStatus(t *testing.T) {
	tests := []struct {
		name, current, latest string
		wants                 []string
	}{
		{"up to date", "0.6.0", "v0.6.0", []string{"up to date", "v0.6.0"}},
		{"newer available", "v0.6.0", "v0.7.0", []string{"newer release available: v0.7.0", "current v0.6.0", "go install github.com/gleicon/devskills@latest"}},
		{"ahead of latest", "0.7.0", "v0.6.0", []string{"ahead of the latest release v0.6.0"}},
		{"dev build", "dev", "v0.6.0", []string{"latest release: v0.6.0", `"dev"`, "no comparison"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := releaseStatus(tt.current, tt.latest)
			for _, want := range tt.wants {
				if !strings.Contains(got, want) {
					t.Errorf("releaseStatus(%q, %q) = %q; missing %q", tt.current, tt.latest, got, want)
				}
			}
		})
	}
}

func TestFetchLatestTag(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
		wantErr string
	}{
		{
			"ok",
			func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"tag_name": "v0.7.0", "name": "v0.7.0"}`))
			},
			"v0.7.0", "",
		},
		{
			"non-200",
			func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) },
			"", "github api",
		},
		{
			"tag is not a release version",
			func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"tag_name": "nightly"}`))
			},
			"", `unexpected release tag "nightly"`,
		},
		{
			"invalid json",
			func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"tag_name": `)) },
			"", "decode github response",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			t.Cleanup(srv.Close)
			got, err := fetchLatestTag(context.Background(), srv.Client(), srv.URL)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("fetchLatestTag() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("fetchLatestTag() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("fetchLatestTag() = %q, want %q", got, tt.want)
			}
		})
	}
}
