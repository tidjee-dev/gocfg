package cli

import (
	"fmt"
	"maps"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/gocfg/env"
	"github.com/tidjee-dev/gocfg/internal/dotenv"
)

// newDoctorCmd builds `gocfg doctor`.
func newDoctorCmd() *cobra.Command {
	var envFile, exampleFile string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the local configuration setup",
		Long: `Prints a human-oriented diagnostic narrative: toolchain,
files and permissions, parse health, .gitignore coverage, secret
values in the example, duplicate keys and OS shadowing. Read-only;
always exits 0 (machine checks with exit codes are check's job).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			ok := func(format string, args ...any) {
				fmt.Fprintf(out, "%s %s\n", okStyle.Render("✓"), fmt.Sprintf(format, args...))
			}
			note := func(format string, args ...any) {
				fmt.Fprintf(out, "%s %s\n", warnStyle.Render("!"), fmt.Sprintf(format, args...))
			}
			bad := func(format string, args ...any) {
				fmt.Fprintf(out, "%s %s\n", failStyle.Render("✗"), fmt.Sprintf(format, args...))
			}

			fmt.Fprintln(out, "gocfg doctor")
			fmt.Fprintf(out, "\ntoolchain: %s\n", runtime.Version())
			if cwd, err := os.Getwd(); err == nil {
				fmt.Fprintf(out, "directory: %s\n", cwd)
			}
			fmt.Fprintln(out)

			stat := func(path, label string) (os.FileMode, []byte, bool) {
				raw, err := os.ReadFile(path)
				if err != nil {
					bad("%s: %s missing (%v)", label, path, err)
					return 0, nil, false
				}
				fi, err := os.Stat(path)
				if err != nil {
					bad("%s: %v", label, err)
					return 0, nil, false
				}
				return fi.Mode().Perm(), raw, true
			}

			if _, err := os.Stat("config"); err != nil || !isDir("config") {
				bad("config/: not found (run gocfg init)")
			} else {
				ok("config/ present")
			}

			envMode, rawEnv, haveEnv := stat(envFile, ".env")
			_, rawExample, haveExample := stat(exampleFile, ".env.example")
			if haveEnv {
				ok(".env present (%04o)", envMode)
			}
			if haveExample {
				ok(".env.example present")
			}

			var fileVars, schema map[string]string
			if haveEnv {
				var err error
				if fileVars, err = dotenv.Parse(rawEnv); err != nil {
					bad(".env: %v", err)
					haveEnv = false
				} else {
					ok(".env parses (%d keys)", len(fileVars))
				}
			}
			if haveExample {
				var err error
				if schema, err = dotenv.Parse(rawExample); err != nil {
					bad(".env.example: %v", err)
					haveExample = false
				} else {
					ok(".env.example parses (%d keys)", len(schema))
				}
			}

			checkGitignore(note)
			if haveEnv {
				reportDuplicates(note, rawEnv)
				reportShadowing(note, fileVars)
			}
			if haveExample {
				reportExampleSecrets(note, schema)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "env file to inspect")
	cmd.Flags().StringVar(&exampleFile, "example-file", ".env.example", "example file to inspect")
	return cmd
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// checkGitignore warns unless .env appears covered by .gitignore.
func checkGitignore(note func(string, ...any)) {
	raw, err := os.ReadFile(".gitignore")
	if err != nil {
		note("no .gitignore (consider ignoring .env)")
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		switch strings.TrimSpace(line) {
		case ".env", "/.env", "*.env", ".env*":
			return // covered, quiet
		}
	}
	note(".env not covered by .gitignore (risk of committing secrets)")
}

// reportDuplicates warns about keys declared more than once in raw
// (the parser applies last-wins silently).
func reportDuplicates(note func(string, ...any), raw []byte) {
	counts := map[string]int{}
	for _, line := range strings.Split(string(raw), "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if rest, ok := cutExport(t); ok {
			t = strings.TrimSpace(rest)
		}
		eq := strings.Index(t, "=")
		if eq < 0 {
			continue
		}
		if k := strings.TrimSpace(t[:eq]); validKeyChars(k) {
			counts[k]++
		}
	}
	for _, k := range slices.Sorted(maps.Keys(counts)) {
		if counts[k] > 1 {
			note("duplicate key in .env: %s (%dx, last wins)", k, counts[k])
		}
	}
}

func cutExport(s string) (string, bool) {
	if !strings.HasPrefix(s, "export") {
		return s, false
	}
	rest := s[len("export"):]
	if rest == "" || (rest[0] != ' ' && rest[0] != '\t') {
		return s, false
	}
	return rest, true
}

func validKeyChars(k string) bool {
	if k == "" {
		return false
	}
	for _, c := range k {
		if c != '_' && c != '-' && c != '.' &&
			!(c >= 'A' && c <= 'Z') && !(c >= 'a' && c <= 'z') &&
			!(c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// reportShadowing lists .env keys whose live value comes from OS.
func reportShadowing(note func(string, ...any), fileVars map[string]string) {
	var shadowed []string
	for k, v := range fileVars {
		if ov, ok := os.LookupEnv(k); ok && ov != v {
			shadowed = append(shadowed, k)
		}
	}
	slices.Sort(shadowed)
	if len(shadowed) > 0 {
		note("%d OS variable(s) shadow .env: %s", len(shadowed), strings.Join(shadowed, ", "))
	}
}

// reportExampleSecrets lists secret-looking keys carrying values.
func reportExampleSecrets(note func(string, ...any), schema map[string]string) {
	var found []string
	for _, k := range slices.Sorted(maps.Keys(schema)) {
		if schema[k] != "" && env.IsSecretKey(k) {
			found = append(found, k)
		}
	}
	if len(found) > 0 {
		note("%d secret(s) have values in .env.example: %s", len(found), strings.Join(found, ", "))
	}
}
