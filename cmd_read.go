package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/go-nfs/nfsv3/nfs"
)

func init() {
	commands["cat"] = xcmd_cat
	commands["b64get"] = xcmd_b64get
	commands["get"] = xcmd_get
}

func xcmd_cat(ctx *ctx_t, args string) {
	xcmd_open_file(ctx, args, func(f *nfs.File) {
		io.Copy(os.Stdout, f)
	})
}

func xcmd_b64get(ctx *ctx_t, args string) {
	xcmd_open_file(ctx, args, func(f *nfs.File) {
		data := bytes.NewBuffer(nil)
		io.Copy(data, f)
		fmt.Println(base64.StdEncoding.EncodeToString(data.Bytes()))
	})
}

func xcmd_get(ctx *ctx_t, args string) {
	dst_fn := path.Base(args)
	if _, err := os.Stat(dst_fn); err == nil {
		fmt.Fprintf(os.Stderr, "file already exists: %s\n", dst_fn)
	}
	dst, err := os.Create(dst_fn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		defer dst.Close()
		xcmd_open_file(ctx, args, func(f *nfs.File) {
			io.Copy(dst, f)
		})
	}
}

func xcmd_open_file(ctx *ctx_t, name string, handler func(*nfs.File)) {
	if fi, err := ctx.mnt.Open(path.Join(ctx.cwd, name)); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		defer fi.Close()

		handler(fi)
	}
}
