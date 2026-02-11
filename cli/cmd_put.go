package cli

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/willscott/go-nfs-client/nfs"
)

var cmdPutFlags struct {
	base64 bool

	cmdCommonAttrFlags
}

func init() {
	commandProviders = append(commandProviders, newPutCmd)
}

func newPutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put <local file> [<remote file>]",
		Args:  cobra.MinimumNArgs(1),
		Run:   runPutCmd,
		Short: "Uploads a file (or from stdin)",
		Example: strings.Join([]string{
			"Upload a file: put webshell.php webshell-3487298.php",
			"Upload a file using the local filename only, specify mode: put webshell.php --mode 0644",
			"Read base64 encoded input from stdin and create a remote file: put - authorized_keys --base64",
		}, "\n"),
	}
	cmd.Flags().BoolVar(&cmdPutFlags.base64, "base64", false, "input is base64 encoded and should be decoded before writing (useful for stdin input)")
	cmdPutFlags.registerCommonAttrFlags(cmd.Flags(), "06777")
	return cmd
}

func runPutCmd(cmd *cobra.Command, args []string) {
	attrOps, mode, ok := cmdPutFlags.parseCommonAttrFlags(cmd.Flags(), false)
	if !ok {
		return
	}

	var src string
	var dst string

	src = args[0]
	if len(args) == 1 {
		dst = filepath.Base(src)
	} else {
		dst = args[1]
	}
	dst = cwd.Relative(dst)

	var r io.Reader
	if src == "-" {
		r = stdinReader()
	} else {
		f, err := os.Open(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open: %s: %v\n", src, err)
			return
		}
		defer f.Close()
		r = f
	}

	nfsCreate(dst, os.FileMode(mode), func(f *nfs.File) {
		if cmdPutFlags.base64 {
			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, r); err != nil {
				fmt.Fprintf(os.Stderr, "copy: %v\n", err)
				return
			}

			str := strings.ReplaceAll(buf.String(), "\n", "")

			raw, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				fmt.Fprintf(os.Stderr, "base64 decoding failed: %v\n", err)
				return
			}

			if _, err := f.Write(raw); err != nil {
				fmt.Fprintf(os.Stderr, "write: %v\n", err)
				return
			}
		} else {
			if _, err := io.Copy(f, r); err != nil {
				fmt.Fprintf(os.Stderr, "copy: %v\n", err)
			}
		}
	}, attrOps...)
}
