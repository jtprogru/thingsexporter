package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:       "completion <bash|zsh|fish|powershell>",
		Short:     "Generate shell completion script",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(deps.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(deps.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(deps.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(deps.Stdout)
			}
			return fmt.Errorf("unknown shell %q", args[0])
		},
	}
}
