package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/go-nfs/nfsv3/nfs"
)

func init() {
	commands["ls"] = xcmd_ls
	commands["cd"] = xcmd_cd
	commands["mkdir"] = xcmd_mkdir
	commands["rmdir"] = xcmd_rmdir
}

func xcmd_ls(ctx *ctx_t, args string) {
	if files, err := ctx.mnt.ReadDirPlus(ctx.cwd); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		var sorter func(files []*nfs.EntryPlus)

		if args == "-byname" {
			sorter = filesort_name
		} else if args != "-nosort" {
			sorter = filesort_mtime
		}
		if sorter != nil {
			sorter(files)
		}
		for _, f := range files {
			dirmarker := " "
			if f.IsDir() {
				dirmarker = "*"
			}
			fmt.Printf("%8o %5d %5d %8d | %s %s[%s]\n", f.Mode(), f.Attr.Attr.UID, f.Attr.Attr.GID, f.Size(), f.ModTime().Format(time.ANSIC), dirmarker, f.Name())
		}
	}
}

func xcmd_mkdir(ctx *ctx_t, args string) {
	if _, err := ctx.mnt.Mkdir(path.Join(ctx.cwd, args), 0777); err != nil {
		fmt.Println(err)
	}
}

func xcmd_rmdir(ctx *ctx_t, args string) {
	if err := ctx.mnt.RemoveAll(path.Join(ctx.cwd, args)); err != nil {
		fmt.Println(err)
	}
}

func xcmd_cd(ctx *ctx_t, args string) {
	if strings.HasPrefix(args, "/") {
		ctx.cwd = path.Clean(args)
	} else {
		ctx.cwd = path.Join(ctx.cwd, args)
	}
}

func filesort_mtime(files []*nfs.EntryPlus) {
	sort.Slice(files, func(i, j int) bool {
		ei := files[i]
		ej := files[j]
		if ei.Name() == "." {
			return true
		} else if ei.Name() == ".." {
			return ej.Name() != "."
		}
		return ei.ModTime().Before(ej.ModTime())
	})
}

func filesort_name(files []*nfs.EntryPlus) {
	sort.Slice(files, func(i, j int) bool {
		ei := files[i]
		ej := files[j]
		if ei.Name() == "." {
			return true
		} else if ei.Name() == ".." {
			return ej.Name() != "."
		}
		return strings.Compare(ei.Name(), ej.Name()) < 0
	})
}
