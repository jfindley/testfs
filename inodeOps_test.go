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
	fs.newInode(in, Uid, Gid, os.FileMode(0644))

	err := fs.Chown("/test", 666, 777)
	if err != nil {
		t.Error(err)
	}

	i := fs.lookupInode(in)

	if i.uid != 666 || i.gid != 777 {
		t.Error("Bad ownership")
	}
}

func TestLink(t *testing.T) {
	fs := NewTestFS()
	in := fs.newInum()
	fs.dirTree.newDentry(in, "src")
	fs.newInode(in, Uid, Gid, os.FileMode(0644))

	err := fs.Link("/src", "/dst")
	if err != nil {
		t.Error(err)
	}

	src, err := fs.find("/src")
	if err != nil {
		t.Error(err)
	}

	dst, err := fs.find("/dst")
	if err != nil {
		t.Error(err)
	}

	if src != dst {
		t.Error("Wrong inode for link")
	}

	if fs.lookupInode(dst).linkCount != 2 {
		t.Error("Wrong link count for inode", fs.lookupInode(dst).linkCount)
	}
}

func TestReadlink(t *testing.T) {
	fs := NewTestFS()

	in, err := fs.create(&fs.dirTree, "tmp", os.FileMode(0644)|os.ModeSymlink)
	if err != nil {
		t.Error(err)
	}

	link := fs.lookupInode(in)
	link.rel = "test"

	res, err := fs.Readlink("tmp")
	if err != nil {
		t.Error(err)
	}

	if res != "test" {
		t.Error("Bad link data")
	}
}

func TestRemove(t *testing.T) {
	fs := NewTestFS()

	err := fs.Mkdir("test", os.FileMode(0500))
	if err != nil {
		t.Error(err)
	}

	Uid = uint16(os.Getuid())

	err = fs.Remove("/test")
	if !os.IsPermission(err) {
		t.Error(err)
	}

	Uid = 0

	err = fs.Remove("test")
	if err != nil {
		t.Error(err)
	}
}
