package testfs

import (
	"testing"
)

func TestLookupPath(t *testing.T) {
	fs := NewTestFS()

	i, err := fs.lookupPath("/")
	if err != nil {
		t.Error(err)
	}
	if i != 1 {
		t.Error("Wrong root inode number")
	}

	tmp := dentry{inode: 2}
	tmp.children = make(map[string]dentry)
	testing := dentry{inode: 3}

	fs.dirTree.children["tmp"] = tmp
	fs.dirTree.children["tmp"].children["testing"] = testing

	i, err = fs.lookupPath("/tmp")
	if err != nil {
		t.Error(err)
	}
	if i != 2 {
		t.Error("Wrong inode number")
	}

	i, err = fs.lookupPath("tmp")
	if err != nil {
		t.Error(err)
	}
	if i != 2 {
		t.Error("Wrong inode number")
	}

	i, err = fs.lookupPath("/tmp/testing")
	if err != nil {
		t.Error(err)
	}
	if i != 3 {
		t.Error("Wrong inode number")
	}

	i, err = fs.lookupPath("/tmp/testing/")
	if err != nil {
		t.Error(err)
	}
	if i != 3 {
		t.Error("Wrong inode number")
	}
}
