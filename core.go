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
)

func init() {
	Uid = uint16(os.Getuid())
	Gid = uint16(os.Getgid())
}

// inode represents an entity in the filesystem.  Children are represented as
// pointers to allow us to simulate hardlinks.
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

func (i *inode) new(name string, uid, gid uint16, mode os.FileMode) error {
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
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.children[name]; ok {
		return os.ErrExist
	}
	i.children[name] = &entry
	i.mtime = time.Now()
	return nil
}

// Methods to satisfy os.FileInfo
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

// TestFS implements an in-memory filesystem.  We use maps rather than
// slices to allow us to scale to large file numbers more efficiently.
type TestFS struct {
	dirTree inode
	cwd     *inode
	cwdPath string
}

func NewTestFS() *TestFS {
	t := new(TestFS)
	t.dirTree.children = make(map[string]*inode)
	t.dirTree.mu = new(sync.Mutex)
	t.dirTree.uid = 0
	t.dirTree.gid = 0
	t.dirTree.mode = os.FileMode(0555) | os.ModeDir
	t.dirTree.xattrs = make(map[string]string)
	t.dirTree.linkCount = 1
	t.dirTree.name = sep
	t.cwd = &t.dirTree
	t.cwdPath = sep
	return t
}

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

		case "", ".":
			continue

		case "..":
			if len(terms) > 0 {
				terms = terms[:len(terms)-1]
			} else {
				return nil, os.ErrNotExist
			}

		default:
			terms = append(terms, elems[i])

		}
	}

	return terms[:len(terms)], nil
}

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

		// If we're at the end of the path, just return it
		if len(terms) == 1 {
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
