package testfs

import (
	"os"
	"testing"
)

func TestChmod(t *testing.T) {
	fs.dirTree.new("testchmod", Uid, Gid, os.FileMode(0755))
	err := fs.Chmod("/testchmod", os.FileMode(0644))
	if err != nil {
		t.Error(err)
	}

	if fs.dirTree.children["testchmod"].mode != os.FileMode(0644) {
		t.Error("Bad file mode", fs.dirTree.children["testchmod"].mode)
	}

	// Test other attributes are preserved
	fs.dirTree.children["testchmod"].mode = os.FileMode(0755) | os.ModeDir

	err = fs.Chmod("/testchmod", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	if fs.dirTree.children["testchmod"].mode != os.FileMode(0700)|os.ModeDir {
		t.Error("Bad file mode")
	}

	fs.dirTree.children["testchmod"].mode = os.FileMode(0755) | os.ModeSocket | os.ModeSetuid

	err = fs.Chmod("/testchmod", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	if fs.dirTree.children["testchmod"].mode != os.FileMode(0700)|os.ModeSocket {
		t.Error("Bad file mode")
	}
}

func TestChown(t *testing.T) {
	fs.dirTree.new("testchown", Uid, Gid, os.FileMode(0644))
	err := fs.Chown("/testchown", 666, 777)
	if err != nil {
		t.Error(err)
	}

	if fs.dirTree.children["testchown"].uid != 666 || fs.dirTree.children["testchown"].gid != 777 {
		t.Error("Bad ownership")
	}
}

func TestLink(t *testing.T) {
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

	err := fs.dirTree.new("testreadlink", Uid, Gid, os.FileMode(0644)|os.ModeSymlink)
	if err != nil {
		t.Error(err)
	}

	fs.dirTree.children["testreadlink"].relName = "/src"

	res, err := fs.Readlink("/testreadlink")
	if err != nil {
		t.Error(err)
	}

	if res != "/src" {
		t.Error("Bad link data")
	}
}

func TestRemove(t *testing.T) {
	err := fs.Mkdir("/testrm", os.FileMode(0500))
	if err != nil {
		t.Error(err)
	}

	ref := fs.dirTree.children["testrm"]
	ref.linkCount = 2

	Uid = 20

	err = fs.Remove("/testrm")
	if !os.IsPermission(err) {
		t.Error(err)
	}

	Uid = 0

	err = fs.Remove("/testrm")
	if err != nil {
		t.Error(err)
	}

	if ref.linkCount != 1 {
		t.Error("Bad link count")
	}

	_, err = fs.find("/testrm")
	if !os.IsNotExist(err) {
		t.Error("Dir not removed")
	}
}

func TestRemoveAll(t *testing.T) {
	err := fs.MkdirAll("/testrmall/first/path", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}
	err = fs.dirTree.children["testrmall"].new("second", Uid, Gid, os.FileMode(0600))
	if err != nil {
		t.Error(err)
	}

	err = fs.RemoveAll("/testrmall")
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/testrmall")
	if !os.IsNotExist(err) {
		t.Error("Dir not removed")
	}
}

func TestRename(t *testing.T) {
	err := fs.Mkdir("/testrename", os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	err = fs.Rename("/testrename", "/testmv")
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/testmv")
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/testrename")
	if !os.IsNotExist(err) {
		t.Error("Old file still present")
	}
}

func TestSymlink(t *testing.T) {
	err := fs.Mkdir("/testsymlink", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	err = fs.Symlink("/testsymlink", "/testlns")
	if err != nil {
		t.Error(err)
	}

	ln, err := fs.find("/testlns")
	if err != nil {
		t.Error(err)
	}

	if ln.relName != "/testsymlink" || ln.rel != fs.dirTree.children["testsymlink"] || ln.mode&os.ModeSymlink == 0 {
		t.Error("Bad link data")
	}
}

func TestStat(t *testing.T) {
	err := fs.MkdirAll("/teststat/test", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	fi, err := fs.Stat("/teststat/test")
	if err != nil {
		t.Error(err)
	}

	if fi.Name() != "test" {
		t.Error("Bad name")
	}
}
