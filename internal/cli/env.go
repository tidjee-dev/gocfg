package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/gocfg/env"
	"github.com/tidjee-dev/gocfg/internal/envsync"
)

// newEnvCmd builds `gocfg env`.
func newEnvCmd() *cobra.Command {
	var envFile, exampleFile, defsFile string
	var check bool

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Synchronize .env from .env.example",
		Long: `Ensures .env contains every key declared in .env.example.
With --defs, generates from a definitions manifest instead (manifest
order preserved, secrets and required Vars appended empty).
Missing keys are appended (example values, empty for secrets); existing
values, comments and order are never touched. With --check, reports
without writing (useful in CI).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := envsync.Options{
				EnvFile: envFile,
				Check:   check,
			}
			var res *envsync.Result
			var err error
			if defsFile != "" {
				var defs []env.Definition
				if defs, err = loadManifest(defsFile); err == nil {
					res, err = envsync.RunDefs(opts, defs)
				}
			} else {
				opts.ExampleFile = exampleFile
				res, err = envsync.Run(opts)
			}
			if err != nil {
				return &ExitError{Code: 1, Err: err}
			}
			out := cmd.OutOrStdout()
			if res.Created {
				if check {
					fmt.Fprintf(out, "%s would create %s\n",
						skipStyle.Render("..."), envFile)
				} else {
					fmt.Fprintf(out, "%s created %s\n",
						okStyle.Render("✓"), envFile)
				}
			}
			for _, k := range res.Added {
				if check {
					fmt.Fprintf(out, "%s would add %s\n",
						skipStyle.Render("..."), k)
				} else {
					fmt.Fprintf(out, "%s added %s\n",
						okStyle.Render("✓"), k)
				}
			}
			for _, w := range res.Warnings {
				fmt.Fprintf(out, "%s %s\n", warnStyle.Render("!"), w)
			}
			if !res.Changed {
				fmt.Fprintln(out, "Environment is in sync.")
				return nil
			}
			if check {
				fmt.Fprintln(out, "\nEnvironment is out of sync.")
				return &ExitError{Code: 1, Err: fmt.Errorf("environment out of sync")}
			}
			fmt.Fprintln(out, "\nEnvironment synchronized.")
			return nil
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "env file to complete")
	cmd.Flags().StringVar(&exampleFile, "example-file", ".env.example", "schema file to read")
	cmd.Flags().StringVar(&defsFile, "defs", "", "manifest file to generate from instead")
	cmd.Flags().BoolVar(&check, "check", false, "report without writing")
	return cmd
}
