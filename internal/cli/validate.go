package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	gocfg "github.com/tidjee-dev/gocfg"
	"github.com/tidjee-dev/gocfg/env"
)

var failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

// newValidateCmd builds `gocfg validate`.
//
// v0.1 takes explicit variable specs via flags because the Var[T]
// definitions API is deferred to v0.2. The CLI loads the env file,
// resolves each declared variable with the typed env getters, and
// reports per-variable results. Exit 1 on any failure.
func newValidateCmd() *cobra.Command {
	var (
		envFile   string
		bools     []string
		ints      []string
		int64s    []string
		floats    []string
		durations []string
		required  []string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate typed configuration values",
		Long: `Loads the env file (default .env, OS values win) and validates
each declared variable, e.g.:

  gocfg validate --int APP_PORT --bool APP_DEBUG --required DATABASE_URL`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := gocfg.LoadEnv(envFile); err != nil {
				return &ExitError{Code: 1, Err: err}
			}
			out := cmd.OutOrStdout()
			failed := false
			check := func(key string, err error) {
				if err != nil {
					failed = true
					fmt.Fprintf(out, "%s %s: %v\n", failStyle.Render("✗"), key, err)
					return
				}
				fmt.Fprintf(out, "%s %s\n", okStyle.Render("✓"), key)
			}
			for _, k := range bools {
				_, err := env.Bool(k, false)
				check(k, err)
			}
			for _, k := range ints {
				_, err := env.Int(k, 0)
				check(k, err)
			}
			for _, k := range int64s {
				_, err := env.Int64(k, 0)
				check(k, err)
			}
			for _, k := range floats {
				_, err := env.Float64(k, 0)
				check(k, err)
			}
			for _, k := range durations {
				_, err := env.Duration(k, 0)
				check(k, err)
			}
			for _, k := range required {
				_, err := env.Required(k)
				check(k, err)
			}
			if failed {
				fmt.Fprintln(out, "\nConfiguration is invalid.")
				return &ExitError{Code: 1, Err: fmt.Errorf("validation failed")}
			}
			fmt.Fprintln(out, "\nConfiguration is valid.")
			return nil
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "env file to load")
	cmd.Flags().StringSliceVar(&bools, "bool", nil, "validate KEY as boolean")
	cmd.Flags().StringSliceVar(&ints, "int", nil, "validate KEY as integer")
	cmd.Flags().StringSliceVar(&int64s, "int64", nil, "validate KEY as int64")
	cmd.Flags().StringSliceVar(&floats, "float", nil, "validate KEY as float")
	cmd.Flags().StringSliceVar(&durations, "duration", nil, "validate KEY as duration")
	cmd.Flags().StringSliceVar(&required, "required", nil, "require KEY to be set and non-empty")
	return cmd
}
