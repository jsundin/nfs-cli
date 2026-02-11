package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	commandProviders = append(commandProviders, newCdCommand)
}

func newCdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cd <dir>",
		Args:    cobra.MinimumNArgs(1),
		Run:     runCdCommand,
		Short:   "Change directory",
		Example: "cd path/to/dir",
	}
	return cmd
}

func runCdCommand(cmd *cobra.Command, args []string) {
	cwd = cwd.Get(args[0])
}
