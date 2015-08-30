package testfs

import (
	"bytes"
	"os"
	"testing"
)

func TestTruncate(t *testing.T) {

	_, err := fs.Create("/testTruncate")
	if err != nil {
		t.Error(err)
	}

	f, err := fs.find("/testTruncate")
	if err != nil {
		t.Fatal(err)
	}

	f.data = []byte("test data")

	err = fs.Symlink("/testTruncate", "/testTruncateLink")
	if err != nil {
		t.Error(err)
	}

	err = fs.Truncate("/testTruncateLink", 4)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(f.data, []byte("test")) != 0 {
		t.Error("Bad data")
	}

	err = fs.Truncate("/testTruncate", 0)
	if err != nil {
		t.Error(err)
	}

	if len(f.data) != 0 {
		t.Error("Bad data")
	}

}

func TestCreate(t *testing.T) {

	f, err := fs.Create("/testCreate")
	if err != nil {
		t.Error(err)
	}

	f.Close()

	_, err = fs.find("/testCreate")
	if err != nil {
		t.Error(err)
	}

}

func TestOpen(t *testing.T) {
	f, err := fs.Open("/testOpen")
	if !os.IsNotExist(err) {
		t.Error("Bad error status")
	}

	f, err = fs.Create("/testOpen")
	if err != nil {
		t.Error(err)
	}
	f.Close()

	_, err = fs.find("/testOpen")
	if err != nil {
		t.Error(err)
	}
}

func TestOpenFile(t *testing.T) {
	_, err := fs.OpenFile("/testOpenFile", os.O_RDWR, 0)
	if !os.IsNotExist(err) {
		t.Error("Bad error status")
	}

	_, err = fs.OpenFile("/testOpenFile", os.O_RDWR|os.O_CREATE, os.FileMode(0664))
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/testOpenFile")
	if err != nil {
		t.Error(err)
	}

	_, err = fs.OpenFile("/testOpenFile", os.O_RDWR|os.O_CREATE|os.O_EXCL, os.FileMode(0664))
	if !os.IsExist(err) {
		t.Error("Bad error status")
	}
}

func TestFileChdir(t *testing.T) {
	f := file{}
	if f.Chdir().Error() != "Unsupported function" {
		t.Fail()
	}
}

func TestFileChmod(t *testing.T) {
	f, err := fs.OpenFile("/testFileChmod1", os.O_RDONLY|os.O_CREATE, os.FileMode(0664))
	if err != nil {
		t.Error(err)
	}

	err = f.Chmod(os.FileMode(1775))
	if !os.IsPermission(err) {
		t.Error("Bad error status")
	}

	f.Close()

	f, err = fs.OpenFile("/testFileChmod2", os.O_RDWR|os.O_CREATE, os.FileMode(0664))
	if err != nil {
		t.Error(err)
	}

	err = f.Chmod(os.FileMode(1775))
	if err != nil {
		t.Error(err)
	}
	f.Close()

	i, err := fs.find("/testFileChmod2")
	if err != nil {
		t.Error(err)
	}

	if i.mode != os.FileMode(1775) {
		t.Error("Bad mode")
	}
}

func TestFileChown(t *testing.T) {
	f, err := fs.OpenFile("/testFileChown", os.O_RDWR|os.O_CREATE, os.FileMode(0664))
	if err != nil {
		t.Error(err)
	}

	err = f.Chown(501, 500)
	if err != nil {
		t.Error(err)
	}
	f.Close()

	i, err := fs.find("/testFileChown")
	if err != nil {
		t.Error(err)
	}

	if i.uid != 501 || i.gid != 500 {
		t.Error("Bad ownership")
	}
}

func TestFileClose(t *testing.T) {
	f, err := fs.OpenFile("/testFileClose", os.O_RDWR|os.O_CREATE, os.FileMode(0664))
	if err != nil {
		t.Error(err)
	}

	f.Close()

	err = f.Chmod(0775)
	if err != os.ErrInvalid {
		t.Error("Bad error status")
	}

}
