package testfs

import (
	"os"
	"testing"
)

func TestChmod(t *testing.T) {
	fs := NewTestFS()
	in := fs.newInum()
	fs.dirTree.children["test"] = newDentry(in)
	fs.files[in] = newInode(Uid, Gid, os.FileMode(0755))

	err := fs.Chmod("/test", os.FileMode(0644))
	if err != nil {
		t.Error(err)
	}

	if fs.files[in].mode != os.FileMode(0644) {
		t.Error("Bad file mode")
	}

	// Test other attributes are preserved
	fs.files[in] = newInode(Uid, Gid, os.FileMode(0755)|os.ModeDir)
	err = fs.Chmod("/test", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	if fs.files[in].mode != os.FileMode(0700)|os.ModeDir {
		t.Error("Bad file mode", fs.files[in].mode, os.FileMode(0700)|os.ModeDir)
	}

	fs.files[in] = newInode(Uid, Gid, os.FileMode(0755)|os.ModeSocket|os.ModeSetuid)
	err = fs.Chmod("/test", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	if fs.files[in].mode != os.FileMode(0700)|os.ModeSocket {
		t.Error("Bad file mode")
	}

}
