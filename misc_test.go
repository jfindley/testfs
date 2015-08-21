package testfs

import (
	"os"
	"testing"
)

func TestChmod(t *testing.T) {
	fs := NewTestFS()
	fs.dirTree.new("test", Uid, Gid, os.FileMode(0755))

	err := fs.Chmod("/test", os.FileMode(0644))
	if err != nil {
		t.Error(err)
	}

	if fs.dirTree.children["test"].mode != os.FileMode(0644) {
		t.Error("Bad file mode", fs.dirTree.children["test"].mode)
	}

	// Test other attributes are preserved
	fs.dirTree.children["test"].mode = os.FileMode(0755) | os.ModeDir

	err = fs.Chmod("/test", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	if fs.dirTree.children["test"].mode != os.FileMode(0700)|os.ModeDir {
		t.Error("Bad file mode")
	}

	fs.dirTree.children["test"].mode = os.FileMode(0755) | os.ModeSocket | os.ModeSetuid

	err = fs.Chmod("/test", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	if fs.dirTree.children["test"].mode != os.FileMode(0700)|os.ModeSocket {
		t.Error("Bad file mode")
	}
}

func TestChown(t *testing.T) {
	fs := NewTestFS()
	fs.dirTree.new("test", Uid, Gid, os.FileMode(0644))

	err := fs.Chown("/test", 666, 777)
	if err != nil {
		t.Error(err)
	}

	if fs.dirTree.children["test"].uid != 666 || fs.dirTree.children["test"].gid != 777 {
		t.Error("Bad ownership")
	}
}

func TestLink(t *testing.T) {
	fs := NewTestFS()
	fs.dirTree.new("src", Uid, Gid, os.FileMode(0644))

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

	if fs.dirTree.children["dst"].linkCount != 2 {
		t.Error("Wrong link count for inode")
	}
}

func TestReadlink(t *testing.T) {
	fs := NewTestFS()

	err := fs.dirTree.new("dst", Uid, Gid, os.FileMode(0644)|os.ModeSymlink)
	if err != nil {
		t.Error(err)
	}

	fs.dirTree.children["dst"].relName = "/src"

	res, err := fs.Readlink("/dst")
	if err != nil {
		t.Error(err)
	}

	if res != "/src" {
		t.Error("Bad link data")
	}
}

func TestRemove(t *testing.T) {
	fs := NewTestFS()
	Uid = 0

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

// func TestRemoveAll(t *testing.T) {

// }
