package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	commandProviders = append(commandProviders, newMvCmd)
}

func newMvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mv <src> <dst>",
		Args:    cobra.MinimumNArgs(2),
		Run:     runMvCmd,
		Short:   "Renames a file",
		Example: "mv shadow shadow.bak",
	}
	return cmd
}

func runMvCmd(cmd *cobra.Command, args []string) {
	src := cwd.Relative(args[0])
	dst := cwd.Relative(args[1])

	if err := target.Rename(src, dst); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}
