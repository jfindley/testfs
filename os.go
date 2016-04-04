package testfs

import (
	"os"
)

type osfs struct{}

// NewOSFS returns an OS filesystem, implemented by the core os package.
func NewOSFS() FileSystem {
	return new(osfs)
}

func (o *osfs) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (o *osfs) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

func (o *osfs) Chown(name string, uid, gid int) error {
	return os.Chown(name, uid, gid)
}

func (o *osfs) Link(oldname, newname string) error {
	return os.Link(oldname, newname)
}

func (o *osfs) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (o *osfs) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (o *osfs) MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

func (o *osfs) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

func (o *osfs) Remove(name string) error {
	return os.Remove(name)
}

func (o *osfs) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (o *osfs) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (o *osfs) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (o *osfs) Truncate(name string, size int64) error {
	return os.Truncate(name, size)
}

func (o *osfs) Create(name string) (file File, err error) {
	return os.Create(name)
}

func (o *osfs) Open(name string) (file File, err error) {
	return os.Open(name)
}

func (o *osfs) OpenFile(name string, flag int, perm os.FileMode) (file File, err error) {
	return os.OpenFile(name, flag, perm)
}

func (o *osfs) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (o *osfs) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}