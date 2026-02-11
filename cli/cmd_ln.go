package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	commandProviders = append(commandProviders, newLnCmd)
}

func newLnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ln <src> <dst>",
		Args:    cobra.MinimumNArgs(2),
		Run:     runLnCmd,
		Short:   "Create a symlink",
		Example: "ln passwd passwd.bak",
	}
	return cmd
}

func runLnCmd(cmd *cobra.Command, args []string) {
	src := args[0]
	dst := cwd.Relative(args[1])

	if err := target.Symlink(src, dst); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}
