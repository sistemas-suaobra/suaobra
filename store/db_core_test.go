package store

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreDbCandidates(t *testing.T) {
	mainDataDir = filepath.FromSlash("/app/data/main")
	candidates := coreDbCandidates()

	wantSuffixes := []string{
		"data/core.db",
		"data/core/core.db",
	}

	if len(candidates) < len(wantSuffixes) {
		t.Fatalf("expected at least %d candidates, got %d: %v", len(wantSuffixes), len(candidates), candidates)
	}

	for _, suffix := range wantSuffixes {
		if !containsPathSuffix(candidates, suffix) {
			t.Fatalf("missing candidate ending with %q in %v", suffix, candidates)
		}
	}
}

func TestCoreAttachSQLCandidates(t *testing.T) {
	mainDataDir = filepath.FromSlash("/app/data/main")

	tests := []struct {
		corePath     string
		wantSuffixes []string
	}{
		{
			corePath:     "/app/data/core.db",
			wantSuffixes: []string{"core.db", "../core.db"},
		},
		{
			corePath:     "/app/data/core/core.db",
			wantSuffixes: []string{"core/core.db", "../core/core.db"},
		},
	}

	for _, tt := range tests {
		got := coreAttachSQLCandidates(tt.corePath)
		for _, suffix := range tt.wantSuffixes {
			if !containsPathSuffix(got, suffix) {
				t.Fatalf("coreAttachSQLCandidates(%q) missing suffix %q, got %v", tt.corePath, suffix, got)
			}
		}
	}
}

func containsPathSuffix(values []string, suffix string) bool {
	suffix = filepath.ToSlash(suffix)
	for _, value := range values {
		if strings.HasSuffix(filepath.ToSlash(value), suffix) {
			return true
		}
	}
	return false
}
