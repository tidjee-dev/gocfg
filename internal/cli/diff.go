package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/gocfg/env"
	"github.com/tidjee-dev/gocfg/internal/envdiff"
)

// newDiffCmd builds `gocfg diff`.
func newDiffCmd() *cobra.Command {
	var envFile, exampleFile, defsFile string

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare manifest, .env.example and .env",
		Long: `Shows one row per key across the definitions manifest (optional),
.env.example and .env: missing keys, extras, default drift and secret
values in the example. File-level only (OS ignored); values never
print. Exits 1 on missing keys or drift.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var defs []env.Definition
			if defsFile != "" {
				var err error
				if defs, err = loadManifest(defsFile); err != nil {
					return &ExitError{Code: 1, Err: err}
				}
			}
			rows, err := envdiff.Compare(envFile, exampleFile, defs)
			if err != nil {
				return &ExitError{Code: 1, Err: err}
			}
			out := cmd.OutOrStdout()
			failed := false
			for _, r := range rows {
				switch r.Status {
				case envdiff.Ok:
					fmt.Fprintf(out, "%s %s\n", okStyle.Render("✓"), r.Key)
				case envdiff.Fail:
					failed = true
					fmt.Fprintf(out, "%s %s: %s\n", failStyle.Render("✗"), r.Key, r.Detail)
				default:
					fmt.Fprintf(out, "%s %s: %s\n", warnStyle.Render("!"), r.Key, r.Detail)
				}
			}
			if failed {
				fmt.Fprintln(out, "\nDifferences found.")
				return &ExitError{Code: 1, Err: fmt.Errorf("differences found")}
			}
			fmt.Fprintln(out, "\nNo differences.")
			return nil
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "env file to compare")
	cmd.Flags().StringVar(&exampleFile, "example-file", ".env.example", "example file to compare")
	cmd.Flags().StringVar(&defsFile, "defs", "", "manifest file to compare against")
	return cmd
}
