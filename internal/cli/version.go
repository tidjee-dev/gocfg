package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Version is injected at release build time via ldflags:
//
//	-X github.com/tidjee-dev/gocfg/internal/cli.Version=v0.1.0
var Version = "dev"

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	dimStyle   = lipgloss.NewStyle().Faint(true)
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the gocfg version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Plain-text first: readable even when styling is stripped.
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n",
				titleStyle.Render("gocfg"), dimStyle.Render(Version))
			return err
		},
	}
}
