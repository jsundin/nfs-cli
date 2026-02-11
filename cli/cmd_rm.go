package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	commandProviders = append(commandProviders, newRmCmd)
}

func newRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <dst>",
		Args:    cobra.MinimumNArgs(1),
		Run:     runRmCmd,
		Short:   "Remove a file",
		Example: "rm etc/shadow",
	}
	return cmd
}

func runRmCmd(cmd *cobra.Command, args []string) {
	fn := cwd.Relative(args[0])
	if err := target.Remove(fn); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}
