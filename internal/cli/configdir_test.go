package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustMkdirAll(t *testing.T, dir, sub string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestCheckConfigDirMatrix(t *testing.T) {
	cases := []struct {
		name     string
		dirs     []string
		dir      string
		explicit bool
		wantErr  bool
		wantWarn bool
	}{
		{"default present", []string{"internal/config"}, "internal/config", false, false, false},
		{"default legacy only", []string{"config"}, "internal/config", false, false, true},
		{"default both warn", []string{"internal/config", "config"}, "internal/config", false, false, true},
		{"default neither", nil, "internal/config", false, true, false},
		{"explicit present silent", []string{"custom", "config"}, "custom", true, false, false},
		{"explicit missing strict", []string{"config"}, "custom", true, true, false},
		{"explicit legacy itself", []string{"config"}, "config", true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, d := range tc.dirs {
				mustMkdirAll(t, dir, d)
			}
			chdir(t, dir)
			var warns []string
			err := checkConfigDir(tc.dir, tc.explicit, func(s string) { warns = append(warns, s) })
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantWarn && len(warns) == 0 {
				t.Fatal("expected legacy warning")
			}
			if !tc.wantWarn && len(warns) != 0 {
				t.Fatalf("unexpected warnings: %v", warns)
			}
			for _, w := range warns {
				if !strings.Contains(w, "legacy config/ found, move to internal/config") {
					t.Fatalf("unexpected warning: %q", w)
				}
			}
		})
	}
}
