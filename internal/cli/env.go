package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/gocfg/internal/envsync"
)

// newEnvCmd builds `gocfg env`.
func newEnvCmd() *cobra.Command {
	var envFile, exampleFile string
	var check bool

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Synchronize .env from .env.example",
		Long: `Ensures .env contains every key declared in .env.example.
Missing keys are appended (example values, empty for secrets); existing
values, comments and order are never touched. With --check, reports
without writing (useful in CI).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, err := envsync.Run(envsync.Options{
				EnvFile:     envFile,
				ExampleFile: exampleFile,
				Check:       check,
			})
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
	cmd.Flags().BoolVar(&check, "check", false, "report without writing")
	return cmd
}
