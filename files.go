package testfs

import (
	"os"
	"path"
	"sync"
	"time"
)

func (t *TestFS) Truncate(name string, size int64) error {
	f, err := fs.find(name)
	if err != nil {
		return err
	}

	if !checkPerm(f, 'w') {
		return os.ErrPermission
	}

	if len(f.data) > int(size) {
		f.data = f.data[:int(size)]
	}
	return nil
}

func (t *TestFS) Create(name string) (File, error) {
	dir, file := path.Split(name)

	d, err := fs.find(dir)
	if err != nil {
		return nil, err
	}

	if !checkPerm(d, 'r', 'w', 'x') {
		return nil, os.ErrPermission
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, err = d.lookup([]string{file}); !os.IsNotExist(err) {
		return nil, os.ErrExist
	}

	err = d.newSkipLock(file, Uid, Gid, os.FileMode(0644))
	if err != nil {
		return nil, err
	}

	return newFd(d.children[file]), nil
}

func (t *TestFS) Open(name string) (file File, err error) {
	in, err := fs.find(name)
	if err != nil {
		return nil, err
	}
	return newFd(in), nil
}

func (t *TestFS) OpenFile(name string, flag int, perm os.FileMode) (file File, err error) {
	return nil, nil
}

// Methods to implement os.FileInfo
func (i *inode) Name() string {
	return i.name
}

func (i *inode) Size() int64 {
	return int64(len(i.data))
}

func (i *inode) Mode() os.FileMode {
	return i.mode
}

func (i *inode) ModTime() time.Time {
	return i.mtime
}

func (i *inode) IsDir() bool {
	if i.mode&os.ModeDir == 0 {
		return false
	}
	return true
}

func (i *inode) Sys() interface{} {
	return i
}

// fd is a thin layer over an inode that represents an open file.
// Most of its purpose is to simulate a concept of an open file.
type fd struct {
	mode  os.FileMode
	id    uintptr
	inode *inode
}

// fd counter to generate unique fd numbers.
type fdCtr struct {
	sync.Mutex
	ctr uintptr
}

// next returns the next fd number.
func (f *fdCtr) next() uintptr {
	f.Lock()
	defer f.Unlock()
	f.ctr++
	return f.ctr
}

// Create a new fd
func newFd(i *inode) *fd {
	f := new(fd)
	f.inode = i
	f.id = fdNum.next()
	return f
}

// Methods to implement File
func (f *fd) Chdir() error {
	return nil
}

func (f *fd) Chmod(mode os.FileMode) error {
	return nil
}

func (f *fd) Chown(uid, gid int) error {
	return nil
}

func (f *fd) Close() error {
	return nil
}

func (f *fd) Fd() uintptr {
	return f.id
}

func (f *fd) Name() string {
	return f.inode.name
}

func (f *fd) Read(b []byte) (n int, err error) {
	return
}

func (f *fd) ReadAt(b []byte, off int64) (n int, err error) {
	return
}

func (f *fd) Readdir(n int) (fi []os.FileInfo, err error) {
	return
}

func (f *fd) Readdirnames(n int) (names []string, err error) {
	return
}

func (f *fd) Seek(offset int64, whence int) (ret int64, err error) {
	return
}

func (f *fd) Stat() (fi os.FileInfo, err error) {
	return
}

func (f *fd) Sync() (err error) {
	return
}

func (f *fd) Truncate(size int64) error {
	return nil
}

func (f *fd) Write(b []byte) (n int, err error) {
	return
}

func (f *fd) WriteAt(b []byte, off int64) (n int, err error) {
	return
}

func (f *fd) WriteString(s string) (ret int, err error) {
	return
}
