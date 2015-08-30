package testfs

import (
	"os"
	"testing"
)

func TestInodeFileInfo(t *testing.T) {
	fs.Mkdir("/testfileinfo", os.FileMode(0775))

	in, err := fs.find("/testfileinfo")
	if err != nil {
		t.Error(err)
	}

	var fileInfo os.FileInfo

	fileInfo = in

	in.data = []byte("testdata")

	if fileInfo.Name() != "testfileinfo" {
		t.Error("Bad name")
	}

	if !fileInfo.IsDir() {
		t.Error("Bad dir status")
	}

	if fileInfo.Mode() != os.FileMode(0775)|os.ModeDir {
		t.Error("Bad file mode")
	}

	if fileInfo.ModTime() != in.mtime {
		t.Error("Bad modtime")
	}

	if fileInfo.Size() != int64(len(in.data)) {
		t.Error("Bad size")
	}

	if fileInfo.Sys() != in {
		t.Error("Bad sys interface")
	}

}

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
	fs.dirTree.children["testreadlink"].rel = new(inode)

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

	ln, ok := fs.dirTree.children["testlns"]
	if !ok {
		t.Fatal("Internal test error")
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

	err = fs.Symlink("/teststat/test", "/teststat/link")

	fi, err = fs.Stat("/teststat/link")
	if err != nil {
		t.Error(err)
	}

	if fi.Name() != "test" {
		t.Error("Bad name")
	}
}

func TestLstat(t *testing.T) {
	err := fs.MkdirAll("/testlstat/test", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	err = fs.Symlink("/testlstat/test", "/testlstat/link")
	if err != nil {
		t.Error(err)
	}

	fi, err := fs.Lstat("/testlstat/link")
	if err != nil {
		t.Error(err)
	}

	if fi.Name() != "link" {
		t.Error("Bad name")
	}

	in, ok := fi.Sys().(*inode)
	if !ok {
		t.Fatal("Bad type")
	}

	if in.rel.name != "test" {
		t.Error("Bad rel name")
	}
}
