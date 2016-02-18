package testfs

import (
	"os"
	"strings"
	"sync"
	"time"
)

const sep = "/"
const inodeAllocSize = 4096

var (
	Uid, Gid uint16
	fd       fdCtr
)

func init() {
	Uid = uint16(os.Getuid())
	Gid = uint16(os.Getgid())
	fd.ctr = 0
}

// inode represents an entity in the filesystem.  Children are represented as
// pointers to allow us to simulate hardlinks.  This is not entirely like a POSIX
// inode, but is named as this to clarify that it can refer to any sort of FS object.
type inode struct {
	name      string
	uid       uint16
	gid       uint16
	mode      os.FileMode
	xattrs    map[string]string
	linkCount uint16
	rel       *inode
	relName   string
	mtime     time.Time
	data      []byte
	children  map[string]*inode
	mu        *sync.Mutex
}

// Create a new inode as a child of this one
func (i *inode) new(name string, uid, gid uint16, mode os.FileMode) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.newSkipLock(name, uid, gid, mode)
}

// Unsafe.  This creates a new inode without locking.  Should only be used
// if the calling function is locking seperately.
func (i *inode) newSkipLock(name string, uid, gid uint16, mode os.FileMode) error {
	if !checkPerm(i, 'w', 'x') {
		return os.ErrPermission
	}
	entry := inode{
		mu:        new(sync.Mutex),
		xattrs:    make(map[string]string),
		name:      name,
		uid:       uid,
		gid:       gid,
		mode:      mode,
		mtime:     time.Now(),
		linkCount: 1,
	}
	if i.IsDir() {
		entry.children = make(map[string]*inode)
		entry.children[".."] = i
	}
	if _, ok := i.children[name]; ok {
		return os.ErrExist
	}
	i.children[name] = &entry
	i.mtime = time.Now()
	return nil
}

// TestFS implements an in-memory filesystem.  We use maps rather than
// slices to allow us to scale to large file numbers more efficiently.
type TestFS struct {
	dirTree inode
	cwd     *inode
	cwdPath string
}

// Creates and initialises a new TestFS filesystem.  Creating a TestFS
// filesystem any other way is not supported.
func NewTestFS(uid, gid int) *TestFS {
	t := new(TestFS)
	t.dirTree.children = make(map[string]*inode)
	t.dirTree.mu = new(sync.Mutex)
	t.dirTree.uid = uint16(uid)
	t.dirTree.gid = uint16(gid)
	t.dirTree.mode = os.FileMode(0755) | os.ModeDir
	t.dirTree.xattrs = make(map[string]string)
	t.dirTree.linkCount = 1
	t.dirTree.name = sep
	t.cwd = &t.dirTree
	t.cwdPath = sep
	return t
}

func NewLocalTestFS() *TestFS {
    return NewTestFS(os.Getuid(), os.Getgid())
    
}

// Split a filesystem path into elements.
func parsePath(path string) ([]string, error) {
	if path == sep {
		return nil, nil
	}

	// Ignore trailing slashes
	if path[len(path)-1:] == sep {
		path = path[:len(path)-1]
	}

	elems := strings.Split(path, sep)
	terms := make([]string, 0, len(elems))

	for i := range elems {

		switch elems[i] {

		// We parse out the . path rather than
		// creating self-referential structures.
		case "", ".":
			continue

		default:
			terms = append(terms, elems[i])

		}
	}

	return terms[:len(terms)], nil
}

// Look up child inodes, recursively if there is more than one term.
func (i *inode) lookup(terms []string) (*inode, error) {

	if len(terms) == 0 {
		return i, nil
	}

	if !i.IsDir() {
		return nil, os.ErrInvalid
	}

	if this, ok := i.children[terms[0]]; ok {

		// Follow symlinks
		if this.mode&os.ModeSymlink == os.ModeSymlink {
			return this.rel.lookup(terms[1:])
		}

		// If we're at the end of the path, check for read perms and return it
		if len(terms) == 1 {

			if !checkPerm(this, 'r') {
				return nil, os.ErrPermission
			}

			return this, nil
		}

		// Make sure we can read the new subdir
		if !checkPerm(this, 'r', 'x') {
			return nil, os.ErrPermission
		}

		return this.lookup(terms[1:])

	}

	return nil, os.ErrNotExist
}

// Look up a symlink as a direct child inode.  This does not
// recurse.
func (i *inode) lookupSymlink(name string) (*inode, error) {
	if !i.IsDir() {
		return nil, os.ErrInvalid
	}

	l, ok := i.children[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	if l.mode&os.ModeSymlink == 0 || l.relName == "" || l.rel == nil {
		return l, os.ErrInvalid
	}

	return l, nil
}

// Verify if the current Uid/Gid has access to the inode.
// Accepts 'r', 'w' and 'x' as permission bits to check.
func checkPerm(i *inode, perms ...rune) bool {
	if Uid == 0 {
		// root can do anything
		return true
	}
	var offset uint

	switch {
	case i.uid == Uid:
		offset = 0
	case i.gid == Gid:
		offset = 3
	default:
		offset = 6
	}

	for _, p := range perms {

		switch p {

		case 'r':
			if i.mode&(1<<uint(9-1-offset)) == 0 {
				return false
			}

		case 'w':
			if i.mode&(1<<uint(9-1-offset-1)) == 0 {
				return false
			}

		case 'x':
			if i.mode&(1<<uint(9-1-offset-2)) == 0 {
				return false
			}

		}

	}
	return true
}

// Find an inode by name in the filesystem
func (t *TestFS) find(path string) (*inode, error) {

	if path == "/" {
		return &t.dirTree, nil
	}

	if path == "" || path == "." {
		return t.cwd, nil
	}

	terms, err := parsePath(path)

	if err != nil {
		return nil, err
	}

	if path[0] == '/' {
		return t.dirTree.lookup(terms)
	}

	return t.cwd.lookup(terms)
}
