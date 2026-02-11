package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var cmdMkdirFlags struct {
	mode string
}

func init() {
	commandProviders = append(commandProviders, newMkdirCmd)
}

func newMkdirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mkdir <dirname>",
		Run:     runMkdirCmd,
		Args:    cobra.MinimumNArgs(1),
		Short:   "Create a directory",
		Example: "mkdir stuff",
	}
	cmd.Flags().StringVarP(&cmdMkdirFlags.mode, "mode", "m", "0777", "mode for new directory")
	return cmd
}

func runMkdirCmd(cmd *cobra.Command, args []string) {
	mode, err := strconv.ParseInt(cmdMkdirFlags.mode, 0, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse mode: %v\n", err)
		return
	}

	path := cwd.Relative(args[0])
	if _, err := target.Mkdir(path, os.FileMode(mode)); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
