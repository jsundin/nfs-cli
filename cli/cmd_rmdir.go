package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	commandProviders = append(commandProviders, newRmdirCmd)
}

func newRmdirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rmdir <dir>",
		Args:    cobra.MinimumNArgs(1),
		Run:     runRmdirCmd,
		Short:   "Removes a directory",
		Example: "rmdir etc",
	}
	return cmd
}

func runRmdirCmd(cmd *cobra.Command, args []string) {
	dir := cwd.Relative(args[0])
	if err := target.RmDir(dir); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
