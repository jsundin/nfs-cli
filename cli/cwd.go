package cli

import (
	"fmt"
	"path/filepath"
)

type CWD struct {
	mountPath    string
	relativePath string
}

func (cwd CWD) String() string {
	return fmt.Sprintf("(%s) %s", cwd.mountPath, cwd.relativePath)
}

func (cwd CWD) Absolute() string {
	return filepath.Join(cwd.mountPath, cwd.relativePath)
}

func (cwd CWD) Relative(dst string) string {
	if len(dst) > 0 && dst[0] == '/' {
		return dst
	}
	return filepath.Join(cwd.relativePath, dst)
}

func (cwd CWD) Get(dst string) CWD {
	return CWD{mountPath: cwd.mountPath, relativePath: cwd.Relative(dst)}
}
