package testfs

import (
	"os"
	"strings"
	"sync"
)

// inum is the inode number type
type inum uint64

// inode is loosely based on a POSIX inode.  It contains the metadata of a file.
type inode struct {
	attrs
	linkCount  uint16
	linkTarget inum
	sync.Mutex
}

// dentry is loosely based on a POSIX dentry.  It maps FS names to inodes.
type dentry struct {
	inode    inum
	children map[string]dentry
	sync.Mutex
}

// TestFS implements an in-memory filesystem.  We use maps rather than
// slices to allow us to scale to large file numbers more efficiently.
type TestFS struct {
	dirTree dentry
	files   map[inum]inode
	data    map[inum][]byte
	cwd     string
}

func NewTestFS() *TestFS {
	t := new(TestFS)
	t.dirTree.inode = 1
	t.dirTree.children = make(map[string]dentry)
	t.cwd = "/"
	return t
}

func (t *TestFS) lookupPath(path string) (inum, error) {
	// Root
	if path == "/" {
		return 1, nil
	}

	// Ignore trailing slashes
	if path[len(path)-1:] == "/" {
		path = path[:len(path)-1]
	}

	// If path does not start with /, prepend CWD.
	if path[0:1] != "/" {
		path = t.cwd + path
	}

	loc := t.dirTree

	// Skip the first /
	terms := strings.Split(path[1:], "/")

	for i := range terms {

		if _, ok := loc.children[terms[i]]; ok {

			// Final term
			if i == len(terms)-1 {
				return loc.children[terms[i]].inode, nil
			}

			loc = loc.children[terms[i]]
		}

	}

	return 0, os.ErrNotExist
}
