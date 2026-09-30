// Package scaffold generates the initial project layout for gocfg init.
// Templates are embedded so the CLI binary is self-contained.
package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Options controls a scaffold run.
type Options struct {
	// Dir is the target directory (created if missing).
	Dir string
	// AppName fills APP_NAME in generated env files.
	AppName string
	// Force overwrites existing files.
	Force bool
	// DryRun reports what would happen without writing.
	DryRun bool
	// Gitignore ensures .env is covered by .gitignore.
	Gitignore bool
}

// FileResult describes the outcome for one file.
type FileResult struct {
	Path    string // relative to Dir
	Created bool   // written (or would be written in DryRun)
	Updated bool   // modified in place (or would be in DryRun)
	Skipped bool   // already existed and Force is false
}

// Result aggregates a scaffold run.
type Result struct {
	Files  []FileResult
	DryRun bool
}

type target struct {
	rel  string
	tmpl string // empty means render from tmpl name
	perm os.FileMode
}

var targets = []target{
	{rel: ".env", tmpl: "env.tmpl", perm: 0o600},
	{rel: ".env.example", tmpl: "env.example.tmpl", perm: 0o644},
	{rel: filepath.Join("config", "app.go"), tmpl: "app.go.tmpl", perm: 0o644},
	{rel: filepath.Join("config", "config.go"), tmpl: "config.go.tmpl", perm: 0o644},
}

// Run generates the project layout. Writes are atomic (tmp + rename).
func Run(opts Options) (*Result, error) {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.AppName == "" {
		opts.AppName = "my-app"
	}
	data := map[string]string{"AppName": opts.AppName}
	res := &Result{DryRun: opts.DryRun}

	for _, tg := range targets {
		content, err := render(tg.tmpl, data)
		if err != nil {
			return nil, err
		}
		if err := writeFile(opts, tg, content, res); err != nil {
			return nil, err
		}
	}
	if opts.Gitignore {
		if err := ensureGitignore(opts, res); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// gitignoreCovered patterns count as ignoring .env.
var gitignoreCovered = []string{".env", "/.env", "*.env"}

// ensureGitignore creates .gitignore with `.env` or appends the entry,
// preserving everything else. Covered files are skipped (idempotent,
// unaffected by Force). Writes are atomic.
func ensureGitignore(opts Options, res *Result) error {
	const rel = ".gitignore"
	dst := filepath.Join(opts.Dir, rel)
	raw, err := os.ReadFile(dst)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("scaffold: %s: %w", rel, err)
	}
	if err == nil {
		for _, line := range strings.Split(string(raw), "\n") {
			t := strings.TrimSpace(line)
			if t == "" || strings.HasPrefix(t, "#") {
				continue
			}
			for _, p := range gitignoreCovered {
				if t == p {
					res.Files = append(res.Files, FileResult{Path: rel, Skipped: true})
					return nil
				}
			}
		}
	}

	fr := FileResult{Path: rel}
	if len(raw) == 0 {
		fr.Created = true
	} else {
		fr.Updated = true
	}
	if opts.DryRun {
		res.Files = append(res.Files, fr)
		return nil
	}
	var out []byte
	if len(raw) == 0 {
		out = []byte(".env\n")
	} else {
		out = append([]byte{}, raw...)
		if out[len(out)-1] != '\n' {
			out = append(out, '\n')
		}
		out = append(out, ".env\n"...)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("scaffold: %s: %w", rel, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".gocfg-*")
	if err != nil {
		return fmt.Errorf("scaffold: %s: %w", rel, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(out); err != nil {
		tmp.Close()
		return fmt.Errorf("scaffold: %s: %w", rel, err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return fmt.Errorf("scaffold: %s: %w", rel, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("scaffold: %s: %w", rel, err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return fmt.Errorf("scaffold: %s: %w", rel, err)
	}
	res.Files = append(res.Files, fr)
	return nil
}

func render(name string, data map[string]string) ([]byte, error) {
	raw, err := templates.ReadFile("templates/" + name)
	if err != nil {
		return nil, fmt.Errorf("scaffold: %w", err)
	}
	tmpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("scaffold: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("scaffold: %w", err)
	}
	return buf.Bytes(), nil
}

func writeFile(opts Options, tg target, content []byte, res *Result) error {
	dst := filepath.Join(opts.Dir, tg.rel)
	if _, err := os.Lstat(dst); err == nil && !opts.Force {
		res.Files = append(res.Files, FileResult{Path: tg.rel, Skipped: true})
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("scaffold: %s: %w", tg.rel, err)
	}

	fr := FileResult{Path: tg.rel, Created: true}
	if opts.DryRun {
		res.Files = append(res.Files, fr)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("scaffold: %s: %w", tg.rel, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".gocfg-*")
	if err != nil {
		return fmt.Errorf("scaffold: %s: %w", tg.rel, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return fmt.Errorf("scaffold: %s: %w", tg.rel, err)
	}
	if err := tmp.Chmod(tg.perm); err != nil {
		tmp.Close()
		return fmt.Errorf("scaffold: %s: %w", tg.rel, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("scaffold: %s: %w", tg.rel, err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return fmt.Errorf("scaffold: %s: %w", tg.rel, err)
	}
	res.Files = append(res.Files, fr)
	return nil
}
