package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/willscott/go-nfs-client/nfs"
)

type cmdCommonAttrFlags struct {
	mode  string
	uid   int
	gid   int
	mtime string
	atime string
}

func (f cmdCommonAttrFlags) registerCommonAttrFlags(fl *pflag.FlagSet, defaultMode string) {
	fl.StringVarP(&f.mode, "mode", "m", defaultMode, "set file mode (numeric)")
	fl.IntVarP(&f.uid, "uid", "u", 0, "owner uid")
	fl.IntVarP(&f.gid, "gid", "g", 0, "owner gid")
	fl.StringVar(&f.mtime, "mtime", "", "set mtime (rfc3339 without timezone, or duration)")
	fl.StringVar(&f.atime, "atime", "", "set atime (rfc3339 without timezone, or duration)")
}

func (f cmdCommonAttrFlags) parseCommonAttrFlags(fl *pflag.FlagSet, includeModeAsAttr bool) ([]fileAttrOperation, os.FileMode, bool) {
	ops := []fileAttrOperation{}
	m := os.FileMode(0)

	if v, err := fl.GetString("mode"); err != nil {
		fmt.Fprintf(os.Stderr, "mode: %v\n", err)
		return ops, m, false
	} else if v != "" {
		if mode, err := parseMode(v); err != nil {
			fmt.Fprintf(os.Stderr, "mode: %v\n", err)
			return ops, m, false
		} else {
			if includeModeAsAttr {
				ops = append(ops, setMode(mode))
			}
			m = mode
		}
	}

	if fl.Changed("uid") {
		v, err := fl.GetInt("uid")
		if err != nil {
			fmt.Fprintf(os.Stderr, "uid: %v\n", err)
			return ops, m, false
		}
		ops = append(ops, setUid(v))
	}

	if fl.Changed("gid") {
		v, err := fl.GetInt("gid")
		if err != nil {
			fmt.Fprintf(os.Stderr, "gid: %v\n", err)
			return ops, m, false
		}
		ops = append(ops, setGid(v))
	}

	if v, err := fl.GetString("mtime"); err != nil {
		fmt.Fprintf(os.Stderr, "mtime: %v\n", err)
		return ops, m, false
	} else if v != "" {
		if t, err := parseTimeOrDuration(v); err != nil {
			fmt.Fprintf(os.Stderr, "mtime: %v\n", err)
			return ops, m, false
		} else {
			ops = append(ops, setMTime(t))
		}
	}

	if v, err := fl.GetString("atime"); err != nil {
		fmt.Fprintf(os.Stderr, "atime: %v\n", err)
		return ops, m, false
	} else if v != "" {
		if t, err := parseTimeOrDuration(v); err != nil {
			fmt.Fprintf(os.Stderr, "atime: %v\n", err)
			return ops, m, false
		} else {
			ops = append(ops, setATime(t))
		}
	}

	return ops, m, true
}

func setMode(mode os.FileMode) fileAttrOperation {
	return func(a *nfs.Sattr3) {
		a.Mode = nfs.SetMode{
			SetIt: true,
			Mode:  uint32(mode),
		}
	}
}

func setUid(uid int) fileAttrOperation {
	return func(a *nfs.Sattr3) {
		a.UID = nfs.SetUID{
			SetIt: true,
			UID:   uint32(uid),
		}
	}
}

func setGid(gid int) fileAttrOperation {
	return func(a *nfs.Sattr3) {
		a.GID = nfs.SetUID{
			SetIt: true,
			UID:   uint32(gid),
		}
	}
}

func setATime(t time.Time) fileAttrOperation {
	return func(a *nfs.Sattr3) {
		a.Atime = nfs.SetTime{
			SetIt: nfs.SetToClientTime,
			Time:  nfs.NFS3Time{Seconds: uint32(t.Unix()), Nseconds: 0},
		}
	}
}

func setMTime(t time.Time) fileAttrOperation {
	return func(a *nfs.Sattr3) {
		a.Mtime = nfs.SetTime{
			SetIt: nfs.SetToClientTime,
			Time:  nfs.NFS3Time{Seconds: uint32(t.Unix()), Nseconds: 0},
		}
	}
}
