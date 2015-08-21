package testfs

import (
	"os"
)

func (t *TestFS) Truncate(name string, size int64) error {
	return nil
}

func (t *TestFS) Create(name string) (file File, err error) {
	return nil, nil
}

func (t *TestFS) Open(name string) (file File, err error) {
	return nil, nil
}

func (t *TestFS) OpenFile(name string, flag int, perm os.FileMode) (file File, err error) {
	return nil, nil
}
