package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/gocfg/internal/envexport"
)

// newExportCmd builds `gocfg export`.
func newExportCmd() *cobra.Command {
	var envFile, exampleFile, format string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Print resolved values for shells and tools",
		Long: `Prints every .env.example key resolved as OS > .env, for
sourcing or piping (eval "$(gocfg export)"). Only schema keys are
emitted, never the whole environment. Values print as-is on your
explicit request: keep the output out of logs. Missing keys export
empty with a stderr warning. Always exits 0 unless files are
unreadable (validation is validate's job).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			switch format {
			case "shell", "dotenv", "json":
			default:
				return &ExitError{Code: 2, Err: fmt.Errorf("unknown format %q (want shell, dotenv or json)", format)}
			}
			entries, err := envexport.Collect(envexport.Options{
				EnvFile: envFile, ExampleFile: exampleFile,
			})
			if err != nil {
				return &ExitError{Code: 1, Err: err}
			}
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()
			for _, e := range entries {
				if e.Missing {
					fmt.Fprintf(errOut, "%s %s is not set (exported empty)\n",
						warnStyle.Render("!"), e.Key)
				}
			}
			switch format {
			case "dotenv":
				fmt.Fprint(out, envexport.FormatDotenv(entries))
			case "json":
				s, err := envexport.FormatJSON(entries)
				if err != nil {
					return &ExitError{Code: 1, Err: err}
				}
				fmt.Fprint(out, s)
			default:
				fmt.Fprint(out, envexport.FormatShell(entries))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "env file to read")
	cmd.Flags().StringVar(&exampleFile, "example-file", ".env.example", "schema bounding output")
	cmd.Flags().StringVar(&format, "format", "shell", "output format: shell, dotenv or json")
	return cmd
}
