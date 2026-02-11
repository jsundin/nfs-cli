package cli

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/pschou/go-unixmode"
	"github.com/spf13/cobra"
	"github.com/willscott/go-nfs-client/nfs"
)

var cmdLsFlags struct {
	sort string
}

func init() {
	commandProviders = append(commandProviders, newLsCmd)
}

func newLsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls [<dir>]",
		Run:     runLsCmd,
		Short:   "List files in directory",
		Example: "ls path/to/dir",
	}
	cmd.Flags().StringVar(&cmdLsFlags.sort, "sort", "", "sort files (name, mtime)")
	return cmd
}

func runLsCmd(cmd *cobra.Command, args []string) {
	var path string
	if len(args) > 0 {
		path = cwd.Relative(args[0])
	} else {
		path = cwd.relativePath
	}

	var sorter func(a, b *nfs.EntryPlus) int
	if cmdLsFlags.sort != "" {
		switch cmdLsFlags.sort {
		case "name":
			sorter = func(a, b *nfs.EntryPlus) int { return strings.Compare(a.Name(), b.Name()) }

		case "mtime":
			sorter = func(a, b *nfs.EntryPlus) int { return a.ModTime().Compare(b.ModTime()) }

		default:
			fmt.Fprintf(os.Stderr, "unknown sort type: %s\n", cmdLsFlags.sort)
		}
	}

	entries, err := target.ReadDirPlus(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	if sorter != nil {
		slices.SortFunc(entries, sorter)
	}

	for _, ent := range entries {
		suffix := ""
		modeFix := ent.Mode()
		if ent.IsDir() {
			modeFix = modeFix | os.FileMode(unixmode.ModeDir)
			suffix = "/"
		} else {
			modeFix = modeFix | os.FileMode(unixmode.ModeRegular)
			suffix = "*"
		}

		m := unixmode.Mode(modeFix)
		fmt.Printf("%5s (%s)  %-5d  %-5d  %-6d  %s  [%s]%s\n", fmt.Sprintf("0%o", ent.Mode()), m.String(), ent.Attr.Attr.UID, ent.Attr.Attr.GID, ent.Size(), ent.ModTime().Format(displayTimeLayout), ent.Name(), suffix)
	}
}
