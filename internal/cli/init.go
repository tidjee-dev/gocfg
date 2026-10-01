package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/tidjee-dev/gocfg/internal/scaffold"
)

var (
	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	skipStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

func newInitCmd() *cobra.Command {
	var force, dryRun, gitignore, definitions bool
	var name, configDir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the initial project structure",
		Long: `Creates .env, .env.example and the configuration directory
(default internal/config) in the current directory.
Existing files are skipped unless --force is given.
Ensures .env is covered by .gitignore unless --gitignore=false.
With --with-definitions, also scaffolds vars.go and
tools/gocfg-gen for the manifest workflow.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			if name == "" {
				name = filepath.Base(dir)
			}
			if configDir == "" {
				configDir = scaffold.DefaultConfigDir
			}
			res, err := scaffold.Run(scaffold.Options{
				Dir: dir, AppName: name, Force: force, DryRun: dryRun,
				Gitignore: gitignore, Definitions: definitions,
				ConfigDir: configDir,
			})
			if err != nil {
				return &ExitError{Code: 2, Err: err}
			}
			out := cmd.OutOrStdout()
			for _, f := range res.Files {
				if f.Path == ".gitignore" {
					switch {
					case f.Created && dryRun:
						fmt.Fprintf(out, "%s would create %s\n",
							skipStyle.Render("..."), f.Path)
					case f.Created:
						fmt.Fprintf(out, "%s created %s\n",
							okStyle.Render("✓"), f.Path)
					case f.Updated && dryRun:
						fmt.Fprintf(out, "%s would update %s\n",
							skipStyle.Render("..."), f.Path)
					case f.Updated:
						fmt.Fprintf(out, "%s updated %s\n",
							okStyle.Render("✓"), f.Path)
					case f.Skipped:
						fmt.Fprintf(out, "%s already ignores .env — skipped\n", f.Path)
					}
					continue
				}
				switch {
				case f.Created && dryRun:
					fmt.Fprintf(out, "%s would create %s\n",
						skipStyle.Render("..."), f.Path)
				case f.Created:
					fmt.Fprintf(out, "%s created %s\n",
						okStyle.Render("✓"), f.Path)
				case f.Skipped:
					fmt.Fprintf(out, "%s already exists — skipped\n", f.Path)
				}
			}
			if !dryRun {
				fmt.Fprintln(out, "\nProject initialized.")
				if res.ModuleFallback {
					fmt.Fprintln(out, "No go.mod found: tools/gocfg-gen uses the example.com/my-app placeholder — fix the import, then run:")
					fmt.Fprintln(out, "  go run ./tools/gocfg-gen > .gocfg.json")
				} else if definitions {
					fmt.Fprintln(out, "Next: go run ./tools/gocfg-gen > .gocfg.json")
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be created")
	cmd.Flags().BoolVar(&gitignore, "gitignore", true, "ensure .env is covered by .gitignore")
	cmd.Flags().BoolVar(&definitions, "with-definitions", false, "scaffold vars.go and tools/gocfg-gen for the manifest workflow")
	cmd.Flags().StringVar(&name, "name", "", "application name (default: directory name)")
	cmd.Flags().StringVar(&configDir, "config-dir", scaffold.DefaultConfigDir, "configuration directory to create")
	return cmd
}
