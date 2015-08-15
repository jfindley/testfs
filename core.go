package testfs

import (
	"os"
	"strings"
	"sync"
)

const sep = "/"
const inodeAllocSize = 128

var (
	Uid, Gid uint16
)

func init() {
	Uid = uint16(os.Getuid())
	Gid = uint16(os.Getgid())
}

// inum is the inode number type
type inum uint64

// inode is loosely based on a POSIX inode.  It contains the metadata of a file.
type inode struct {
	attrs
	linkCount  uint16
	linkTarget inum
	mu         *sync.Mutex
}

func newInode(uid, gid uint16, mode os.FileMode) inode {
	var i inode
	i.xattrs = make(map[string]string)
	i.uid = uid
	i.gid = gid
	i.mode = mode
	i.mu = new(sync.Mutex)

	return i
}

// dentry is loosely based on a POSIX dentry.  It maps FS names to inodes.
type dentry struct {
	inode    inum
	children map[string]dentry
	mu       sync.Mutex
}

func newDentry(i inum) dentry {
	var d dentry
	d.children = make(map[string]dentry)
	d.inode = i
	return d
}

func (d *dentry) lookup(name string) *dentry {
	if this, ok := d.children[name]; ok {
		return &this
	}
	return nil
}

// TestFS implements an in-memory filesystem.  We use maps rather than
// slices to allow us to scale to large file numbers more efficiently.
type TestFS struct {
	dirTree dentry
	files   map[inum]inode
	data    map[inum][]byte
	cwd     string
	maxInum inum
	sync.Mutex
}

func NewTestFS() *TestFS {
	t := new(TestFS)
	t.dirTree.inode = 1
	t.maxInum = 1
	t.dirTree.children = make(map[string]dentry)
	t.files = make(map[inum]inode)
	t.data = make(map[inum][]byte)
	t.cwd = "/"
	return t
}

func (t *TestFS) parsePath(path string) ([]string, error) {
	if path == sep {
		return nil, nil
	}

	// Ignore trailing slashes
	if path[len(path)-1:] == sep {
		path = path[:len(path)-1]
	}

	// If path does not start with /, prepend CWD.
	if path[0:1] != sep {
		path = t.cwd + sep + path
	}

	elems := strings.Split(path[1:], sep)
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

func (t *TestFS) lookupPath(terms []string) (*dentry, error) {

	if terms == nil || len(terms) == 0 {
		return &t.dirTree, nil
	}

	loc := &t.dirTree

	for i := range terms {

		if this := loc.lookup(terms[i]); this != nil {

			if i == len(terms)-1 {
				return this, nil
			}

			// Make sure we can read the new subdir
			if !t.checkPerm(this.inode, 'r', 'x') {
				return nil, os.ErrPermission
			}

			loc = this

		}

	}

	return nil, os.ErrNotExist
}

func (t *TestFS) checkPerm(in inum, perms ...rune) bool {
	var (
		i      inode
		offset uint
		ok     bool
	)

	if i, ok = t.files[in]; !ok {
		return false
	}

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

func (t *TestFS) newInum() inum {
	t.Lock()
	t.maxInum++
	i := t.maxInum
	t.Unlock()
	return i
}

func (t *TestFS) find(path string) (inum, error) {

	terms, err := t.parsePath(path)
	if err != nil {
		return 0, err
	}

	d, err := t.lookupPath(terms)
	if err != nil {
		return 0, err
	}

	if !t.checkPerm(d.inode, 'r') {
		return 0, os.ErrNotExist
	}
	return d.inode, nil
}
