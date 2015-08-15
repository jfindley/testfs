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

func BenchmarkParsePath(b *testing.B) {
	path := "/test/path/with/five/elements"

	fs := NewTestFS()

	for n := 0; n < b.N; n++ {
		_, err := fs.parsePath(path)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestLookupPath(t *testing.T) {
	fs := NewTestFS()
	tmp := dentry{inode: 2}
	tmp.children = make(map[string]dentry)
	test := dentry{inode: 3}

	fs.dirTree.children["tmp"] = tmp
	fs.dirTree.children["tmp"].children["test"] = test
	fs.files[2] = *newInode(Uid, Gid, os.FileMode(0777))
	fs.files[3] = *newInode(Uid, Gid, os.FileMode(0777))

	i, err := fs.lookupPath(nil)
	if err != nil {
		t.Error(err)
	}
	if i.inode != 1 {
		t.Error("Wrong root inode number")
	}

	i, err = fs.lookupPath([]string{"tmp"})
	if err != nil {
		t.Error(err)
	}
	if i.inode != 2 {
		t.Error("Wrong inode number")
	}

	i, err = fs.lookupPath([]string{"tmp", "test"})
	if err != nil {
		t.Error(err)
	}
	if i.inode != 3 {
		t.Error("Wrong inode number")
	}

}

func BenchmarkLookupPath(b *testing.B) {
	path := "/test/path/with/five/elements"

	fs := NewTestFS()
	err := fs.MkdirAll(path, os.FileMode(0775))
	if err != nil {
		b.Error(err)
	}

	terms, err := fs.parsePath(path)
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		_, err = fs.lookupPath(terms)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestCheckPerm(t *testing.T) {
	fs := NewTestFS()

	fs.files[1] = *newInode(Uid, 666, os.FileMode(0000))
	fs.files[2] = *newInode(666, Gid, os.FileMode(0000))
	fs.files[3] = *newInode(666, 666, os.FileMode(0000))
	fs.files[4] = *newInode(Uid, 666, os.FileMode(0700))
	fs.files[5] = *newInode(666, Gid, os.FileMode(0070))
	fs.files[6] = *newInode(666, 666, os.FileMode(0007))

	// Check failures
	if fs.checkPerm(1, 'r') {
		t.Error("Permission check failed")
	}
	if fs.checkPerm(1, 'w') {
		t.Error("Permission check failed")
	}
	if fs.checkPerm(1, 'x') {
		t.Error("Permission check failed")
	}
	if fs.checkPerm(2, 'r') {
		t.Error("Permission check failed")
	}
	if fs.checkPerm(2, 'w') {
		t.Error("Permission check failed")
	}
	if fs.checkPerm(2, 'x') {
		t.Error("Permission check failed")
	}
	if fs.checkPerm(3, 'r') {
		t.Error("Permission check failed")
	}
	if fs.checkPerm(3, 'w') {
		t.Error("Permission check failed")
	}
	if fs.checkPerm(3, 'x') {
		t.Error("Permission check failed")
	}

	// Check success
	if !fs.checkPerm(4, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	if !fs.checkPerm(5, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	if !fs.checkPerm(6, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}

}

func BenchmarkCheckPerm(b *testing.B) {
	fs := NewTestFS()
	fs.files[1] = *newInode(Uid, Gid, os.FileMode(0644))
	for n := 0; n < b.N; n++ {
		if !fs.checkPerm(1, 'r', 'w') {
			b.Error("Permission check failed")
		}
	}
}
