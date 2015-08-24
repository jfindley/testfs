package testfs

import (
	"os"
	"path"
	"time"
	"unsafe"
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

	return d.children[file], nil
}

func (t *TestFS) Open(name string) (file File, err error) {
	return nil, nil
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

// Methods to implement File
func (i *inode) Chdir() error {
	return nil
}

func (i *inode) Chmod(mode os.FileMode) error {
	return nil
}

func (i *inode) Chown(uid, gid int) error {
	return nil
}

func (i *inode) Close() error {
	return nil
}

func (i *inode) Fd() uintptr {
	return uintptr(unsafe.Pointer(i))
}

func (i *inode) Read(b []byte) (n int, err error) {
	return
}

func (i *inode) ReadAt(b []byte, off int64) (n int, err error) {
	return
}

func (i *inode) Readdir(n int) (fi []os.FileInfo, err error) {
	return
}

func (i *inode) Readdirnames(n int) (names []string, err error) {
	return
}

func (i *inode) Seek(offset int64, whence int) (ret int64, err error) {
	return
}

func (i *inode) Stat() (fi os.FileInfo, err error) {
	return
}

func (i *inode) Sync() (err error) {
	return
}

func (i *inode) Truncate(size int64) error {
	return nil
}

func (i *inode) Write(b []byte) (n int, err error) {
	return
}

func (i *inode) WriteAt(b []byte, off int64) (n int, err error) {
	return
}

func (i *inode) WriteString(s string) (ret int, err error) {
	return
}
