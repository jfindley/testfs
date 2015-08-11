package testfs

import (
	"os"
	"testing"
)

func TestParsePath(t *testing.T) {
	fs := NewTestFS()
	tmp := dentry{inode: 2}
	tmp.children = make(map[string]dentry)
	test := dentry{inode: 3}

	fs.dirTree.children["tmp"] = tmp
	fs.dirTree.children["tmp"].children["test"] = test

	terms, err := fs.parsePath("/")
	if err != nil || terms != nil {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 1 || terms[0] != "tmp" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 1 || terms[0] != "tmp" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/test")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/test//")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/./test/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/test/../test/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/../../test/")
	if err != os.ErrNotExist {
		t.Error("Parse failure")
	}

	fs.cwd = "/tmp"
	terms, err = fs.parsePath("test")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

}

func TestLookupPath(t *testing.T) {
	fs := NewTestFS()
	tmp := dentry{inode: 2}
	tmp.children = make(map[string]dentry)
	test := dentry{inode: 3}

	fs.dirTree.children["tmp"] = tmp
	fs.dirTree.children["tmp"].children["test"] = test

	i, err := fs.lookupPath("/")
	if err != nil {
		t.Error(err)
	}
	if i != root {
		t.Error("Wrong root inode number")
	}

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

	i, err = fs.lookupPath("/tmp/test")
	if err != nil {
		t.Error(err)
	}
	if i != 3 {
		t.Error("Wrong inode number")
	}

}
