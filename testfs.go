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
	sync.Mutex
}

func newInode(uid, gid uint16, mode os.FileMode) *inode {
	var i inode
	i.xattrs = make(map[string]string)
	i.uid = uid
	i.gid = gid
	i.mode = mode

	return &i
}

// dentry is loosely based on a POSIX dentry.  It maps FS names to inodes.
type dentry struct {
	inode    inum
	children map[string]dentry
	sync.Mutex
}

func newDentry(i inum) *dentry {
	var d dentry
	d.children = make(map[string]dentry)
	d.inode = i
	return &d
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

			loc = this

		}

	}

	return nil, os.ErrNotExist
}

func checkPerm(i *inode, perms ...rune) bool {
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

func (t *TestFS) newInum() inum {
	t.Lock()
	t.maxInum++
	i := t.maxInum
	t.Unlock()
	return i
}

func (t *TestFS) Mkdir(name string, perm os.FileMode) error {
	// Ensure the dir mode is set
	perm |= os.ModeDir

	terms, err := t.parsePath(name)
	if err != nil {
		return err
	}

	// Don't try to create root
	if terms == nil {
		return nil
	}

	d, err := t.lookupPath(terms[:len(terms)-1])
	if err != nil {
		return err
	}

	d.Lock()
	defer d.Unlock()

	// Fail if dir exists
	if d.lookup(terms[len(terms)-1]) != nil {
		return os.ErrExist
	}

	// Create the directory
	i := t.newInum()
	d.children[terms[len(terms)-1]] = *newDentry(i)
	t.files[i] = *newInode(Uid, Gid, perm)

	return nil
}

func (t *TestFS) MkdirAll(name string, perm os.FileMode) error {
	// Ensure the dir mode is set
	perm |= os.ModeDir

	terms, err := t.parsePath(name)
	if err != nil {
		return err
	}

	// Don't try to create root
	if terms == nil {
		return nil
	}

	dir := t.dirTree

	for i := range terms {
		d, err := t.lookupPath(terms[:i+1])
		if os.IsNotExist(err) {

			// Create the directory
			newInum := t.newInum()

			dir.children[terms[i]] = *newDentry(newInum)
			dir = dir.children[terms[i]]
			t.files[newInum] = *newInode(Uid, Gid, perm)

		} else if err != nil {
			return err
		} else {
			dir = *d
		}
	}

	return nil
}
