// Package cli implements the gocfg Cobra commands.
// It depends on the gocfg core (env, dotenv); the core must never
// import this package nor Cobra/Lipgloss.
package cli

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gocfg",
		Short: "Type-safe configuration manager for Go applications",
		Long: `gocfg manages Go application configuration: initialization,
environment files, and validation. Configuration itself stays
plain Go code; this CLI only handles project files.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newEnvCmd())
	root.AddCommand(newCheckCmd())
	root.AddCommand(newExportCmd())
	root.AddCommand(newDiffCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newValidateCmd())
	return root
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
