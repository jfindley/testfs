package testfs

import (
	"os"
)

// Filesystem defines the basic operation of a filesystem.  It provides
// most of the basic functionality of the os package.
type FileSystem interface {
	Chdir(dir string) error
	Chmod(name string, mode os.FileMode) error
	Chown(name string, uid, gid int) error
	Link(oldname, newname string) error
	Getwd() (dir string, err error)
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Readlink(name string) (string, error)
	Remove(name string) error
	RemoveAll(path string) error
	Rename(oldpath, newpath string) error
	Symlink(oldname, newname string) error
	Truncate(name string, size int64) error
	Create(name string) (file File, err error)
	Open(name string) (file File, err error)
	OpenFile(name string, flag int, perm os.FileMode) (file File, err error)
	Lstat(path string) (os.FileInfo, error)
	Stat(path string) (os.FileInfo, error)
}

// File is analogous to os.File, providing the same functions.
type File interface {
	Chdir() error
	Chmod(mode os.FileMode) error
	Chown(uid, gid int) error
	Close() error
	Fd() uintptr
	Name() string
	Read(b []byte) (n int, err error)
	ReadAt(b []byte, off int64) (n int, err error)
	Readdir(n int) (fi []os.FileInfo, err error)
	Readdirnames(n int) (names []string, err error)
	Seek(offset int64, whence int) (ret int64, err error)
	Stat() (fi os.FileInfo, err error)
	Sync() (err error)
	Truncate(size int64) error
	Write(b []byte) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
	WriteString(s string) (ret int, err error)
}
