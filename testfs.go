package testfs

import (
	"os"
	"strings"
	"sync"
)

const sep = "/"
const root = 1

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
	t.dirTree.inode = root
	t.dirTree.children = make(map[string]dentry)
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

func (t *TestFS) lookupPath(path string) (inum, error) {
	terms, err := t.parsePath(path)
	if err != nil {
		return 0, err
	}

	if terms == nil {
		return root, nil
	}

	loc := t.dirTree

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
