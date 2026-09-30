package cli

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	gocfg "github.com/tidjee-dev/gocfg"
	"github.com/tidjee-dev/gocfg/env"
	"github.com/tidjee-dev/gocfg/internal/dotenv"
)

var failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
var warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

// newValidateCmd builds `gocfg validate`.
//
// With no flags it validates all: every key declared in the example file
// must resolve to a non-empty value (OS > .env), and the env file must
// parse. With flags it validates the given typed specs instead (Var[T]
// definitions arrive in v0.2). Exit 1 on any failure.
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
		Short: "Validate configuration values",
		Long: `Validates every variable declared in .env.example by default.
With flags, validates the given typed specs instead, e.g.:

  gocfg validate
  gocfg validate --int APP_PORT --bool APP_DEBUG --required DATABASE_URL`,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			if noSpecs(bools, ints, int64s, floats, durations, required) {
				if err := validateAll(out, envFile, check); err != nil {
					return &ExitError{Code: 1, Err: err}
				}
			} else {
				if err := gocfg.LoadEnv(envFile); err != nil {
					return &ExitError{Code: 1, Err: err}
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

func noSpecs(lists ...[]string) bool {
	for _, l := range lists {
		if len(l) > 0 {
			return false
		}
	}
	return true
}

// examplePath derives the schema file for an env file: the default
// ".env" pairs with ".env.example", a custom path with "<path>.example".
func examplePath(envFile string) string {
	if envFile == ".env" {
		return ".env.example"
	}
	return envFile + ".example"
}

// validateAll checks every key declared in the example file. It reports
// through check and warns about extra keys present in the env file.
// A returned error is a hard failure (unreadable schema, malformed env).
func validateAll(out io.Writer, envFile string, check func(key string, err error)) error {
	if err := gocfg.LoadEnv(envFile); err != nil {
		return err
	}
	example := examplePath(envFile)
	raw, err := os.ReadFile(example)
	if err != nil {
		return fmt.Errorf("cannot read schema %s: %w", example, err)
	}
	schema, err := dotenv.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid schema %s: %w", example, err)
	}
	for _, k := range slices.Sorted(maps.Keys(schema)) {
		_, err := env.Required(k)
		check(k, err)
	}
	// Warn about env file keys the schema does not declare.
	if data, err := os.ReadFile(envFile); err == nil {
		if fileVars, err := dotenv.Parse(data); err == nil {
			for _, k := range slices.Sorted(maps.Keys(fileVars)) {
				if _, ok := schema[k]; !ok {
					fmt.Fprintf(out, "%s %s (not in %s)\n",
						warnStyle.Render("!"), k, example)
				}
			}
		}
	}
	return nil
}
