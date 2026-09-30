package cli

import (
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/gocfg/env"
	"github.com/tidjee-dev/gocfg/internal/dotenv"
)

// newCheckCmd builds `gocfg check`.
//
// check is a shallow static inspection: file existence, parseability and
// key presence (in .env or OS, any value). It never mutates the process
// environment and performs no type coercion — that is validate's job.
// Exit 1 on missing/unparseable files or keys.
func newCheckCmd() *cobra.Command {
	var envFile, exampleFile string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check project configuration health",
		Long: `Inspects configuration files without touching the environment:
config/ exists, .env.example exists and parses, .env parses, and every
schema key is present in .env or the OS environment.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			failed := false
			ok := func(label string) {
				fmt.Fprintf(out, "%s %s\n", okStyle.Render("✓"), label)
			}
			bad := func(label string, err error) {
				failed = true
				fmt.Fprintf(out, "%s %s: %v\n", failStyle.Render("✗"), label, err)
			}
			warn := func(msg string) {
				fmt.Fprintf(out, "%s %s\n", warnStyle.Render("!"), msg)
			}

			if fi, err := os.Stat("config"); err != nil || !fi.IsDir() {
				bad("configuration directory", fmt.Errorf("config/ not found"))
			} else {
				ok("configuration directory")
			}

			rawExample, err := os.ReadFile(exampleFile)
			var schema map[string]string
			if err != nil {
				bad("environment example", err)
			} else if schema, err = dotenv.Parse(rawExample); err != nil {
				bad("environment example", fmt.Errorf("invalid %s: %w", exampleFile, err))
			} else {
				ok("environment example")
			}

			rawEnv, err := os.ReadFile(envFile)
			present := map[string]string{}
			if err != nil {
				if os.IsNotExist(err) {
					warn(fmt.Sprintf("%s not found — satisfying keys from OS environment", envFile))
				} else {
					bad("environment file", err)
				}
			} else if present, err = dotenv.Parse(rawEnv); err != nil {
				bad("environment file", fmt.Errorf("invalid %s: %w", envFile, err))
			} else {
				ok("environment file")
			}

			if schema != nil {
				missing := false
				for _, k := range slices.Sorted(maps.Keys(schema)) {
					if _, ok := present[k]; ok {
						continue
					}
					if _, ok := os.LookupEnv(k); ok {
						continue
					}
					missing = true
					bad(fmt.Sprintf("key %s", k), fmt.Errorf("not in %s nor OS environment", envFile))
				}
				if !missing {
					ok("keys present")
				}
				for _, k := range slices.Sorted(maps.Keys(present)) {
					if _, ok := schema[k]; !ok {
						warn(fmt.Sprintf("%s (not in %s)", k, exampleFile))
					}
				}
				for _, k := range slices.Sorted(maps.Keys(schema)) {
					if v := schema[k]; v != "" && env.IsSecretKey(k) {
						warn(fmt.Sprintf("%s has a value in %s", k, exampleFile))
					}
				}
			}

			if failed {
				fmt.Fprintln(out, "\nConfiguration looks inconsistent.")
				return &ExitError{Code: 1, Err: fmt.Errorf("check failed")}
			}
			fmt.Fprintln(out, "\nConfiguration looks consistent.")
			return nil
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "env file to inspect")
	cmd.Flags().StringVar(&exampleFile, "example-file", ".env.example", "schema file to inspect")
	return cmd
}
