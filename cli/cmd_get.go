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

var cmdGetFlags struct {
	stdout bool
	base64 bool
	force  bool
}

func init() {
	commandProviders = append(commandProviders, newGetCmd)
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <remote file> [<local file>]",
		Args:  cobra.MinimumNArgs(1),
		Run:   runGetCmd,
		Short: "Download a file",
		Example: strings.Join([]string{
			"Download a file to local file 'passwd', overwrite if it already exists: get passwd --force",
			"Download a file to stdout, base64 encoded: get passwd --stdout --base64",
		}, "\n"),
	}
	cmd.Flags().BoolVar(&cmdGetFlags.stdout, "stdout", false, "display on stdout")
	cmd.Flags().BoolVar(&cmdGetFlags.base64, "base64", false, "base64 encode contents")
	cmd.Flags().BoolVarP(&cmdGetFlags.force, "force", "f", false, "force overwriting local files")
	return cmd
}

func runGetCmd(cmd *cobra.Command, args []string) {
	var w io.Writer
	downloadFilename := cwd.Relative(args[0])

	if cmdGetFlags.stdout {
		w = os.Stdout
	} else {
		var targetFilename string
		if len(args) > 1 {
			targetFilename = args[1]
		} else {
			targetFilename = filepath.Base(downloadFilename)
		}

		if _, err := os.Stat(targetFilename); err == nil && !cmdGetFlags.force {
			fmt.Fprintf(os.Stderr, "%s: exists (try --force)\n", targetFilename)
			return
		}

		f, err := os.Create(targetFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create: %s: %v\n", targetFilename, err)
			return
		}
		defer f.Close()
		w = f
	}

	downloadFile(downloadFilename, w, cmdGetFlags.base64)
}

func downloadFile(fn string, dst io.Writer, base64encodeOutput bool) {
	nfsOpen(fn, func(f *nfs.File) {
		if base64encodeOutput {
			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, f); err != nil {
				fmt.Fprintf(os.Stderr, "copy failed: %v\n", err)
			} else if _, err := dst.Write([]byte(base64.StdEncoding.EncodeToString(buf.Bytes()))); err != nil {
				fmt.Fprintf(os.Stderr, "base64 encoding failed: %v\n", err)
			}
		} else {
			if _, err := io.Copy(dst, f); err != nil {
				fmt.Fprintf(os.Stderr, "copy failed: %v\n", err)
			}
		}
	})
}
