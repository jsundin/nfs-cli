package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cmdCatFlags struct {
	base64 bool
}

func init() {
	commandProviders = append(commandProviders, newCatCmd)
}

func newCatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cat <file>",
		Args:    cobra.MinimumNArgs(1),
		Run:     runCatCmd,
		Short:   "Downloads a file to stdout",
		Example: "cat id_rsa --base64",
	}
	cmd.Flags().BoolVar(&cmdCatFlags.base64, "base64", false, "base64 encode output")
	return cmd
}

func runCatCmd(cmd *cobra.Command, args []string) {
	fn := cwd.Relative(args[0])
	downloadFile(fn, os.Stdout, cmdCatFlags.base64)
	fmt.Println()
}
