package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	commandProviders = append(commandProviders, newPwdCmd)
}

func newPwdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pwd",
		Run:     runPwdCmd,
		Short:   "Print working directory",
		Example: "pwd",
	}
	return cmd
}

func runPwdCmd(cmd *cobra.Command, args []string) {
	fmt.Printf("Mount:    %s\n", cwd.mountPath)
	fmt.Printf("Relative: %s\n", cwd.relativePath)
	fmt.Printf("Absolute: %s\n", cwd.Absolute())
}
