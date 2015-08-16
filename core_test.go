package testfs

import (
	"os"
	"testing"
)

func TestParsePath(t *testing.T) {
	fs := NewTestFS()
	fs.dirTree.newDentry(2, "tmp")
	fs.dirTree.children["tmp"].newDentry(3, "test")

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
	fs.dirTree.newDentry(2, "tmp")
	fs.dirTree.children["tmp"].newDentry(3, "test")

	fs.newInode(2, Uid, Gid, os.FileMode(0777))
	fs.newInode(3, Uid, Gid, os.FileMode(0777))

	d, err := fs.lookupPath(nil)
	if err != nil {
		t.Fatal(err)
	}
	if d == nil {
		t.Fatal("No error and no dentry")
	}

	if d.inode != 1 {
		t.Error("Wrong root inode number")
	}

	d, err = fs.lookupPath([]string{"tmp"})
	if err != nil {
		t.Error(err)
	}
	if d.inode != 2 {
		t.Error("Wrong inode number")
	}

	// /tmp is not a dir, make sure this fails
	d, err = fs.lookupPath([]string{"tmp", "test"})
	if err != os.ErrInvalid || d != nil {
		t.Error("Bad error status")
	}

	i := fs.lookupInode(2)
	i.mode = os.FileMode(0777) | os.ModeDir
	d, err = fs.lookupPath([]string{"tmp", "test"})
	if err != nil {
		t.Fatal(err)
	}
	if d == nil {
		t.Fatal("No error and no dentry")
	}

	if d.inode != 3 {
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

	fs.newInode(2, Uid, 666, os.FileMode(0000))
	fs.newInode(3, 666, Gid, os.FileMode(0000))
	fs.newInode(4, 666, 666, os.FileMode(0000))
	fs.newInode(5, Uid, 666, os.FileMode(0700))
	fs.newInode(6, 666, Gid, os.FileMode(0070))
	fs.newInode(7, 666, 666, os.FileMode(0007))

	// Check failures
	if i := fs.lookupInode(2); fs.checkPerm(i, 'r') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(2); fs.checkPerm(i, 'w') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(2); fs.checkPerm(i, 'x') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(3); fs.checkPerm(i, 'r') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(3); fs.checkPerm(i, 'w') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(3); fs.checkPerm(i, 'x') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(4); fs.checkPerm(i, 'r') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(4); fs.checkPerm(i, 'w') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(4); fs.checkPerm(i, 'x') {
		t.Error("Permission check failed")
	}

	// Check success
	if i := fs.lookupInode(5); !fs.checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(6); !fs.checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	if i := fs.lookupInode(7); !fs.checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}

}

func BenchmarkCheckPerm(b *testing.B) {
	fs := NewTestFS()
	fs.newInode(2, Uid, Gid, os.FileMode(0644))

	for n := 0; n < b.N; n++ {
		if i := fs.lookupInode(2); !fs.checkPerm(i, 'r', 'w') {
			b.Error("Permission check failed")
		}
	}
}

func TestFind(t *testing.T) {
	fs := NewTestFS()

	_, err := fs.find("/tmp/test")
	if !os.IsNotExist(err) {
		t.Error("Bad error status")
	}

	testInum := fs.newInum()
	fs.dirTree.newDentry(testInum, "tmp")
	fs.newInode(testInum, Uid, Gid, os.FileMode(0755)|os.ModeDir)
	testInum = fs.newInum()
	fs.dirTree.children["tmp"].newDentry(testInum, "test")
	fs.newInode(testInum, Uid, Gid, os.FileMode(0755)|os.ModeDir)

	in, err := fs.find("/tmp/test")
	if err != nil {
		t.Error(err)
	}

	if fs.files[in].mode != os.FileMode(0755)|os.ModeDir {
		t.Error("Bad inode permissions")
	}
}
