package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the thingsexporter root command with its subcommands.
// When invoked without a subcommand, the behavior is equivalent to
// `thingsexporter export` with all defaults.
func NewRootCmd(deps Deps) *cobra.Command {
	f := defaultExportFlags()
	root := &cobra.Command{
		Use:           "thingsexporter",
		Short:         "Export Things 3 SQLite database to JSON or Markdown",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runExport(cmd.Context(), deps, f)
		},
	}
	addExportFlags(root, &f)
	root.SetOut(deps.Stdout)
	root.SetErr(deps.Stderr)

	root.AddCommand(newExportCmd(deps))
	root.AddCommand(newInspectCmd(deps))
	root.AddCommand(newVersionCmd(deps))
	root.AddCommand(newCompletionCmd(deps))
	return root
}

// Execute builds and runs the root command, printing any error to Stderr.
func Execute(deps Deps) error {
	root := NewRootCmd(deps)
	err := root.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "error:", err.Error())
	}
	return err
}
