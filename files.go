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

	return createFile(d, file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func (t *TestFS) Open(name string) (File, error) {
	return t.OpenFile(name, os.O_RDONLY, 0)
}

// Create a new file and open it.  Fail if file exists.
func createFile(dir *inode, name string, flag int, perm os.FileMode) (File, error) {
	if !checkPerm(dir, 'r', 'w', 'x') {
		return nil, os.ErrPermission
	}

	dir.mu.Lock()
	defer dir.mu.Unlock()

	if _, err := dir.lookup([]string{name}); !os.IsNotExist(err) {
		return nil, os.ErrExist
	}

	err := dir.newSkipLock(name, Uid, Gid, perm)
	if err != nil {
		return nil, err
	}

	return newFile(dir.children[name], flag), nil
}

// Open an existing file.  Fail if it does not exist.
func openFile(dir *inode, name string, flag int) (File, error) {
	if !checkPerm(dir, 'r', 'x') {
		return nil, os.ErrPermission
	}

	dir.mu.Lock()
	defer dir.mu.Unlock()
	f, err := dir.lookup([]string{name})
	if err != nil {
		return nil, err
	}

	switch {

	case flag&os.O_RDWR == os.O_RDWR:
		if !checkPerm(f, 'r', 'w') {
			return nil, os.ErrPermission
		}

	case flag&os.O_WRONLY == os.O_WRONLY:
		if !checkPerm(f, 'w') {
			return nil, os.ErrPermission
		}

	default:
		if !checkPerm(f, 'r') {
			return nil, os.ErrPermission
		}
	}

	return newFile(f, flag), nil
}

func (t *TestFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	dir, file := path.Split(name)

	d, err := t.find(dir)
	if err != nil {
		return nil, err
	}

	if flag&os.O_CREATE == os.O_CREATE {

		f, err := createFile(d, file, flag, perm)

		switch {

		case flag&os.O_EXCL == os.O_EXCL:
			return f, err

		case os.IsExist(err):
			return openFile(d, file, flag)

		default:
			return f, err

		}

	}

	return openFile(d, file, flag)
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

// file is a thin layer over an inode that represents an open file.
// Most of its purpose is to simulate a concept of an open file.
type file struct {
	flag  int
	id    uintptr
	inode *inode
}

// fdCtr is a counter to generate unique fd numbers.
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

// Open a new file
func newFile(i *inode, flag int) *file {
	f := new(file)
	f.inode = i
	f.flag = flag
	f.id = fd.next()
	return f
}

// Methods to implement File
func (f *file) Chdir() error {
	return nil
}

func (f *file) Chmod(mode os.FileMode) error {
	return nil
}

func (f *file) Chown(uid, gid int) error {
	return nil
}

func (f *file) Close() error {
	f.inode = nil
	f = nil
	return nil
}

func (f *file) Fd() uintptr {
	return f.id
}

func (f *file) Name() string {
	return f.inode.name
}

func (f *file) Read(b []byte) (n int, err error) {
	return
}

func (f *file) ReadAt(b []byte, off int64) (n int, err error) {
	return
}

func (f *file) Readdir(n int) (fi []os.FileInfo, err error) {
	return
}

func (f *file) Readdirnames(n int) (names []string, err error) {
	return
}

func (f *file) Seek(offset int64, whence int) (ret int64, err error) {
	return
}

func (f *file) Stat() (fi os.FileInfo, err error) {
	return
}

func (f *file) Sync() (err error) {
	return
}

func (f *file) Truncate(size int64) error {
	return nil
}

func (f *file) Write(b []byte) (n int, err error) {
	return
}

func (f *file) WriteAt(b []byte, off int64) (n int, err error) {
	return
}

func (f *file) WriteString(s string) (ret int, err error) {
	return
}
