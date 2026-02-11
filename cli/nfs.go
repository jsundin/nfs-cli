package cli

import (
	"fmt"
	"os"

	"github.com/willscott/go-nfs-client/nfs"
)

type fileAttrOperation func(*nfs.Sattr3)

var (
	fileTypeMap = map[uint32]string{
		nfs.NF3Reg:  "regular file",
		nfs.NF3Dir:  "directory",
		nfs.NF3Blk:  "block device",
		nfs.NF3Chr:  "character device",
		nfs.NF3Lnk:  "link",
		nfs.NF3Sock: "socket",
		nfs.NF3FIFO: "fifo",
	}

	fileTypeMapDefault = "unknown"
)

func nfsOpen(fn string, handler func(*nfs.File)) {
	f, err := target.Open(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", fn, err)
		return
	}
	defer f.Close()
	handler(f)
}

func nfsCreate(fn string, mode os.FileMode, handler func(*nfs.File), opts ...fileAttrOperation) {
	fh, err := target.Create(fn, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: create: %v\n", fn, err)
		return
	}

	if _, err := target.SetAttr(fh, nfs.Sattr3{Size: nfs.SetSize{SetIt: true, Size: 0}}); err != nil {
		fmt.Fprintf(os.Stderr, "%s: trunc: %v\n", fn, err)
		return
	}

	opts = append(opts, setMode(mode))
	defer func() {
		attr := nfs.Sattr3{}
		for _, opt := range opts {
			opt(&attr)
		}
		if _, err := target.SetAttr(fh, attr); err != nil {
			fmt.Fprintf(os.Stderr, "%s: setattr: %v\n", fn, err)
		}
	}()

	nfsOpen(fn, handler)
}
