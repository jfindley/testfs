package testfs

import (
	"os"
	"strings"
	"sync"
)

const sep = "/"
const inodeAllocSize = 1024
const dentryAllocSize = 64

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
	id        inum
	uid       uint16
	gid       uint16
	mode      os.FileMode
	xattrs    map[string]string
	linkCount uint16
	relNum    inum
	data      []byte
	mu        *sync.Mutex
}

// dentry is loosely based on a POSIX dentry.  It maps FS names to inodes.
type dentry struct {
	inode    inum
	children map[string]dentry
	mu       *sync.Mutex
}

func (d dentry) newDentry(i inum, name string) error {
	newDentry := dentry{
		children: make(map[string]dentry),
		inode:    i,
		mu:       new(sync.Mutex),
	}
	d.mu.Lock()
	if _, ok := d.children[name]; ok {
		return os.ErrExist
	}
	d.children[name] = newDentry
	d.mu.Unlock()
	return nil
}

func (d dentry) lookup(name string) *dentry {
	if this, ok := d.children[name]; ok {
		return &this
	}
	return nil
}

// TestFS implements an in-memory filesystem.  We use maps rather than
// slices to allow us to scale to large file numbers more efficiently.
type TestFS struct {
	dirTree dentry
	files   []inode
	cwd     string
	maxInum inum
	sync.Mutex
}

func NewTestFS() *TestFS {
	t := new(TestFS)
	t.dirTree.inode = 1
	t.maxInum = 1
	t.dirTree.children = make(map[string]dentry)
	t.dirTree.mu = new(sync.Mutex)
	t.files = make([]inode, 1, inodeAllocSize)
	t.newInode(1, 0, 0, os.FileMode(0555)|os.ModeDir)
	t.cwd = "/"
	return t
}

func (t *TestFS) newInode(i inum, uid, gid uint16, mode os.FileMode) {

	t.files = append(t.files, inode{
		id:        i,
		xattrs:    make(map[string]string),
		uid:       uid,
		gid:       gid,
		mode:      mode,
		mu:        new(sync.Mutex),
		linkCount: 1,
	})
}

func (t *TestFS) lookupInode(in inum) *inode {
	for i := range t.files {
		if t.files[i].id == in {
			return &t.files[i]
		}
	}
	return nil
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

			thisInode := t.lookupInode(this.inode)
			if thisInode == nil {
				return nil, os.ErrNotExist
			}

			// Make sure we can read the new subdir
			if !checkPerm(thisInode, 'r', 'x') {
				return nil, os.ErrPermission
			}

			// Make sure this is actually a directory before ascending the tree
			if thisInode.mode&os.ModeDir == 0 {
				return nil, os.ErrInvalid
			}

			loc = this

		}

	}
	return nil, os.ErrNotExist
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

func (t *TestFS) newInum() inum {
	t.Lock()
	t.maxInum++
	i := t.maxInum
	t.Unlock()
	return i
}

func (t *TestFS) find(path string) (inum, error) {
	if path == "/" {
		return t.dirTree.inode, nil
	}

	terms, err := t.parsePath(path)
	if err != nil {
		return 0, err
	}

	d, err := t.lookupPath(terms)
	if err != nil {
		return 0, err
	}

	i := t.lookupInode(d.inode)
	if i == nil {
		return 0, os.ErrNotExist
	}

	if !checkPerm(i, 'r') {
		return 0, os.ErrPermission
	}
	return d.inode, nil
}

func (t *TestFS) findDentry(path string) (*dentry, error) {
	if path == "/" {
		return &t.dirTree, nil
	}

	terms, err := t.parsePath(path)
	if err != nil {
		return nil, err
	}

	d, err := t.lookupPath(terms)
	if err != nil {
		return nil, err
	}

	i := t.lookupInode(d.inode)
	if i == nil {
		return nil, os.ErrNotExist
	}

	if !checkPerm(i, 'r', 'x') {
		return nil, os.ErrPermission
	}
	return d, nil
}

func (t *TestFS) create(dir *dentry, name string, perm os.FileMode) (inum, error) {
	// Check permissions first
	dirInode := t.lookupInode(dir.inode)
	if dirInode == nil {
		return 0, os.ErrNotExist
	}

	if !checkPerm(dirInode, 'w') {
		return 0, os.ErrPermission
	}

	i := t.newInum()

	err := dir.newDentry(i, name)
	if err != nil {
		return 0, err
	}

	t.newInode(i, Uid, Gid, perm)

	return i, nil

}
