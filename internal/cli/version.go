package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Version is injected at release build time via ldflags:
//
//	-X github.com/tidjee-dev/gocfg/internal/cli.Version=v0.1.0
//
// When empty (e.g. `go install ...@latest`, which runs no ldflags),
// resolveVersion falls back to the binary's build info.
var Version = "dev"

// resolveVersion reports the effective version: the ldflags-injected
// Version wins; otherwise the main module version from the build info
// (`go install pkg@version` records it); otherwise "dev".
func resolveVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}

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
				titleStyle.Render("gocfg"), dimStyle.Render(resolveVersion()))
			return err
		},
	}
}
