package testfs

import (
	"errors"
	"os"
	"path"
	"sort"
	"sync"
)

// file is a thin layer over an inode to simulate the concept
// of an open file.
type file struct {
	flag  int     // Permission bits
	id    uintptr // Unique ID
	inode *inode  // Reference to an inode
	pos   int     // Read/Write position
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

// Create a new file and open it.  Fail if file exists.
func createFile(dir *inode, name string, flag int, perm os.FileMode) (File, error) {
	if dir == nil {
		return nil, os.ErrInvalid
	}

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
	if dir == nil {
		return nil, os.ErrInvalid
	}

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

func truncateData(data []byte, size int64) []byte {
	s := int(size)

	if len(data) > s {
		// We do this to avoid memory leaks when truncating large files.
		newData := make([]byte, s)
		copy(newData, data[:s])
		data = newData
	}

	return data
}

func (t *TestFS) Truncate(name string, size int64) error {
	f, err := fs.find(name)
	if err != nil {
		return err
	}

	if !checkPerm(f, 'w') {
		return os.ErrPermission
	}

	f.data = truncateData(f.data, size)

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

// Methods to implement File

func (f *file) writable() bool {
	switch {

	case f.flag&os.O_RDWR == os.O_RDWR:
		return true

	case f.flag&os.O_WRONLY == os.O_WRONLY:
		return true

	default:
		return false

	}
}

func (f *file) readable() bool {
	switch {

	case f.flag&os.O_RDWR == os.O_RDWR:
		return true

	case f.flag&os.O_RDONLY == os.O_RDONLY:
		return true

	default:
		return false

	}
}

// Chdir is actually quite difficult to support
// as we currently don't support walking upwards,
// and testfs.chdir requires the full path.
// As it doesn't seem a particularly useful function
// given the presence of the TestFs.Chdir() function,
// for now just return an error.
func (f *file) Chdir() error {
	return errors.New("Unsupported function")
}

func (f *file) Chmod(mode os.FileMode) error {

	if f.inode == nil {
		return os.ErrInvalid
	}
	if !f.writable() {
		return os.ErrPermission
	}

	return f.inode.chmod(mode)
}

func (f *file) Chown(uid, gid int) error {

	if f.inode == nil {
		return os.ErrInvalid
	}
	if !f.writable() {
		return os.ErrPermission
	}

	return f.inode.chown(uid, gid)
}

func (f *file) Close() error {
	// Clear the inode reference before clearing the pointer
	// in case some other function happens to keep a reference to it.
	f.inode = nil
	f = nil
	return nil
}

func (f *file) Fd() uintptr {
	if f == nil || f.inode == nil {
		return 0
	}

	return f.id
}

func (f *file) Name() string {
	if f == nil || f.inode == nil {
		return ""
	}

	return f.inode.name
}

func (f *file) Read(b []byte) (n int, err error) {
	if f == nil || f.inode == nil {
		return 0, os.ErrInvalid
	}
	if !f.readable() {
		return 0, os.ErrPermission
	}

	return f.ReadAt(b, 0)
}

func (f *file) ReadAt(b []byte, off int64) (n int, err error) {
	if f == nil || f.inode == nil {
		return 0, os.ErrInvalid
	}
	if !f.readable() {
		return 0, os.ErrPermission
	}

	start := int(off) + f.pos
	end := len(b) + f.pos

	switch {

	case start > len(f.inode.data):
		return -1, os.ErrInvalid

	case start+end > len(f.inode.data):
		copy(b, f.inode.data[start:])
		n = len(f.inode.data) - start

	default:
		copy(b, f.inode.data[start:end])
		n = len(b)

	}

	// Set the new fd position
	f.pos = start + n

	return
}

// Return a sorted array of directory contents
func (f *file) ls() ([]os.FileInfo, error) {
	if f == nil || f.inode == nil {
		return nil, os.ErrInvalid
	}
	if !f.readable() {
		return nil, os.ErrPermission
	}

	if !f.inode.IsDir() {
		return nil, os.ErrInvalid
	}

	f.inode.mu.Lock()
	defer f.inode.mu.Unlock()

	var entries []string

	for name := range f.inode.children {
		if name == ".." {
			continue
		}
		entries = append(entries, name)
	}

	sort.Strings(entries)

	fi := make([]os.FileInfo, len(entries))

	for i := range entries {
		fi[i] = f.inode.children[entries[i]]
	}

	return fi, nil

}

func (f *file) Readdir(n int) ([]os.FileInfo, error) {
	entries, err := f.ls()
	if err != nil {
		return nil, err
	}

	if n > 0 && n < len(entries) {
		return entries[:n], nil
	}
	return entries, nil
}

func (f *file) Readdirnames(n int) (names []string, err error) {
	entries, err := f.ls()
	if err != nil {
		return nil, err
	}

	for i := range entries {

		switch {

		case n > 0 && n <= len(names):
			return

		default:
			names = append(names, entries[i].Name())

		}

	}
	return
}

func (f *file) Seek(offset int64, whence int) (ret int64, err error) {
	if f == nil || f.inode == nil {
		return 0, os.ErrInvalid
	}

	switch whence {

	case 0:
		f.pos = int(offset)

	case 1:
		f.pos += int(offset)

	case 2:
		f.pos = len(f.inode.data) + int(offset)

	default:
		return 0, os.ErrInvalid

	}

	if f.pos > len(f.inode.data) {
		return 0, os.ErrInvalid
	}

	return int64(f.pos), nil
}

func (f *file) Stat() (fi os.FileInfo, err error) {
	if f == nil || f.inode == nil {
		return nil, os.ErrInvalid
	}

	if !f.readable() {
		return nil, os.ErrPermission
	}

	return f.inode, nil
}

// This makes absolutely no sense in a memory-backed FS.  Don't do anything.
func (f *file) Sync() error {
	if f == nil || f.inode == nil {
		return os.ErrInvalid
	}
	if !f.writable() {
		return os.ErrPermission
	}
	return nil
}

func (f *file) Truncate(size int64) error {
	if f == nil || f.inode == nil {
		return os.ErrInvalid
	}
	if !f.writable() {
		return os.ErrPermission
	}

	f.inode.data = truncateData(f.inode.data, size)

	return nil
}

func (f *file) Write(b []byte) (n int, err error) {
	if f == nil || f.inode == nil {
		return 0, os.ErrInvalid
	}
	if !f.writable() {
		return 0, os.ErrPermission
	}
	return f.WriteAt(b, 0)
}

func (f *file) WriteAt(b []byte, off int64) (n int, err error) {
	if f == nil || f.inode == nil {
		return 0, os.ErrInvalid
	}
	if !f.writable() {
		return 0, os.ErrPermission
	}
	return
}

func (f *file) WriteString(s string) (ret int, err error) {
	if f == nil || f.inode == nil {
		return 0, os.ErrInvalid
	}
	if !f.writable() {
		return 0, os.ErrPermission
	}
	return f.WriteAt([]byte(s), 0)
}
