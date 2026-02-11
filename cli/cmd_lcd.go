package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	commandProviders = append(commandProviders, newLcdCmd)
}

func newLcdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lcd <local dir>",
		Args:    cobra.MinimumNArgs(1),
		Run:     runLcdCmd,
		Short:   "Change local directory",
		Example: "cd ..",
	}
	return cmd
}

func runLcdCmd(cmd *cobra.Command, args []string) {
	if err := os.Chdir(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "chdir: %v\n", err)
		return
	}
}
