package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/willscott/go-nfs-client/nfs"
)

var cmdSetattrFlags struct {
	cmdCommonAttrFlags
}

func init() {
	commandProviders = append(commandProviders, newSetattrCmd)
}

func newSetattrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setattr <dst>",
		Args:  cobra.MinimumNArgs(1),
		Run:   runSetattrCmd,
		Short: "Sets attributes for a file entry",
		Example: strings.Join([]string{
			"Set file mode: setattr -m 04755 bash",
			"Change owner to root: setattr -u 0 -g 0 bash",
			"Change mtime to 1 hour ago: setattr --mtime -1h job.json",
			"Change atime to specific time: setattr --atime 2012-08-06T05:17:00 job.json",
		}, "\n"),
	}
	cmdSetattrFlags.registerCommonAttrFlags(cmd.Flags(), "")
	return cmd
}

func runSetattrCmd(cmd *cobra.Command, args []string) {
	attrOps, _, ok := cmdSetattrFlags.parseCommonAttrFlags(cmd.Flags(), true)
	if !ok {
		return
	}

	path := cwd.Relative(args[0])

	if len(attrOps) == 0 {
		fmt.Fprintf(os.Stderr, "no changes requested\n")
		return
	}
	attr := nfs.Sattr3{}
	for _, op := range attrOps {
		op(&attr)
	}
	if err := target.Setattr(path, attr); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
