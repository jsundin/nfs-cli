package cli

import (
	"fmt"
	"os"

	"github.com/pschou/go-unixmode"
	"github.com/spf13/cobra"
	"github.com/willscott/go-nfs-client/nfs"
)

func init() {
	commandProviders = append(commandProviders, newAttrCmd)
}

func newAttrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attr <fileentry>",
		Run:     runAttrCmd,
		Args:    cobra.MinimumNArgs(1),
		Short:   "Displays some file information, similar to stat",
		Example: "attr shadow",
	}
	return cmd
}

func runAttrCmd(cmd *cobra.Command, args []string) {
	path := cwd.Get(args[0])
	attr, err := target.Getattr(path.relativePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "getattr: %v\n", err)
		return
	}

	ftype, found := fileTypeMap[attr.Type]
	if !found {
		ftype = fileTypeMapDefault
	}

	m := unixmode.Mode(attr.Mode())

	var typeExtras string
	if attr.Type == nfs.NF3Lnk {
		if lnkf, err := target.Open(path.relativePath); err != nil {
			typeExtras = " => (unknown)"
		} else {
			if lnkdst, err := lnkf.Readlink(); err != nil {
				typeExtras = " => (unknown)"
			} else {
				typeExtras = fmt.Sprintf(" => %s", lnkdst)
				lnkf.Close()
			}
		}
	}

	fmt.Printf("  File: %s\n", path.Absolute())
	fmt.Printf("  Size: %d (%s%s)\n", attr.Size(), ftype, typeExtras)
	fmt.Printf("Access: (0%o/%s)  Uid: %d  Gid: %d\n", m, m.PermString(), attr.UID, attr.GID)
	fmt.Printf("Access: %s\n", formatTime(attr.Atime))
	fmt.Printf("Modify: %s\n", formatTime(attr.Mtime))
	fmt.Printf("Change: %s\n", formatTime(attr.Ctime))
}
