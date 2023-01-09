package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/go-nfs/nfsv3/nfs"
)

func init() {
	commands["put"] = xcmd_put
	commands["b64put"] = xcmd_b64put
	commands["type"] = xcmd_type
	commands["rm"] = xcmd_rm
}

func xcmd_rm(ctx *ctx_t, args string) {
	if err := ctx.mnt.Remove(path.Join(ctx.cwd, args)); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func xcmd_put(ctx *ctx_t, args string) {
	src, err := os.Open(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		defer src.Close()
		xcmd_create_file(ctx, args, func(f *nfs.File) {
			io.Copy(f, src)
		})
	}
}

func xcmd_b64put(ctx *ctx_t, args string) {
	fmt.Print(">> ")
	b64data := make([]byte, 8192)
	n, _ := os.Stdin.Read(b64data)
	raw, err := base64.StdEncoding.DecodeString(string(b64data[:n-1]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		src := bytes.NewBuffer(raw)
		xcmd_create_file(ctx, args, func(f *nfs.File) {
			io.Copy(f, src)
		})
	}
}

func xcmd_type(ctx *ctx_t, args string) {
	fmt.Println("End with a . on an empty line (or do an EOF):")
	reader := bufio.NewReader(os.Stdin)

	xcmd_create_file(ctx, args, func(f *nfs.File) {
		for {
			input, err := reader.ReadString('\n')
			if input == ".\n" {
				break
			}
			if _, err := f.Write([]byte(input)); err != nil {
				fmt.Fprintf(os.Stderr, "warning: write failed (will keep trying): %s\n", err)
			}
			if err != nil {
				break
			}
		}
	})
}

func xcmd_create_file(ctx *ctx_t, name string, handler func(*nfs.File)) {
	if fh, err := ctx.mnt.Create(name, 06777); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		if err := ctx.mnt.SetAttrByFh(fh, nfs.Sattr3{
			Mode: nfs.SetMode{
				SetIt: true,
				Mode:  06777,
			},
		}); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not set file mode (still trying to write the file though): %s\n", err)
		}

		if fo, err := ctx.mnt.OpenByFh(fh, nil); err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			defer fo.Close()
			handler(fo)
		}
	}
}
