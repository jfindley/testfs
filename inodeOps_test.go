package testfs

import (
	"os"
	"testing"
)

func TestChmod(t *testing.T) {
	fs := NewTestFS()
	in := fs.newInum()
	fs.dirTree.newDentry(in, "test")
	fs.newInode(in, Uid, Gid, os.FileMode(0755))

	err := fs.Chmod("/test", os.FileMode(0644))
	if err != nil {
		t.Error(err)
	}

	if fs.files[in].mode != os.FileMode(0644) {
		t.Error("Bad file mode")
	}

	// Test other attributes are preserved
	i := fs.lookupInode(in)
	i.mode = os.FileMode(0755) | os.ModeDir

	err = fs.Chmod("/test", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	if i.mode != os.FileMode(0700)|os.ModeDir {
		t.Error("Bad file mode", fs.files[in].mode, os.FileMode(0700)|os.ModeDir)
	}

	i.mode = os.FileMode(0755) | os.ModeSocket | os.ModeSetuid

	err = fs.Chmod("/test", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	if i.mode != os.FileMode(0700)|os.ModeSocket {
		t.Error("Bad file mode")
	}

}

func TestChown(t *testing.T) {
	fs := NewTestFS()
	in := fs.newInum()
	fs.dirTree.newDentry(in, "test")
	fs.newInode(in, Uid, Gid, os.FileMode(0755))

	err := fs.Chown("/test", 666, 777)
	if err != nil {
		t.Error(err)
	}

	i := fs.lookupInode(in)

	if i.uid != 666 || i.gid != 777 {
		t.Error("Bad ownership")
	}

}
