package testfs

import (
	"os"
)

func (t *TestFS) Chmod(name string, mode os.FileMode) error {
	return nil
}

func (t *TestFS) Chown(name string, uid, gid int) error {
	return nil
}
