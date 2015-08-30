package testfs

import (
	"bytes"
	"io"
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

func TestFileFd(t *testing.T) {
	f, err := fs.Create("/testFileFd")
	if err != nil {
		t.Error(err)
	}

	if f.Fd() == 0 {
		t.Error("Bad FD")
	}

	f.Close()
	if f.Fd() != 0 {
		t.Error("Bad FD")
	}
}

func TestFileName(t *testing.T) {
	f, err := fs.Create("/testFileName")
	if err != nil {
		t.Error(err)
	}

	if f.Name() != "testFileName" {
		t.Error("Bad name")
	}

	f.Close()
	if f.Name() != "" {
		t.Error("Bad name")
	}
}

func TestFileRead(t *testing.T) {
	f, err := fs.OpenFile("/testFileRead", os.O_RDWR|os.O_CREATE, os.FileMode(0664))
	if err != nil {
		t.Error(err)
	}

	data := []byte("short test data")

	f.(*file).inode.data = data

	buf := make([]byte, 20)

	n, err := f.Read(buf)
	if err != io.EOF {
		t.Error("Bad error status")
	}
	if n != 15 {
		t.Error("Bad output len")
	}

	if bytes.Compare(buf[:n], data) != 0 {
		t.Error("Bad data")
	}

	data = []byte("long test data......................................................................................")
	f.(*file).inode.data = data
	// Reset position
	f.(*file).pos = 0

	n, err = f.Read(buf)
	if err != nil {
		t.Error(err)
	}
	if n != 20 {
		t.Error("Bad output len")
	}

	if bytes.Compare(buf, data[:20]) != 0 {
		t.Error("Bad data")
	}

	// Test next read gets the next chunk
	n, err = f.Read(buf)
	if err != nil {
		t.Error(err)
	}
	if n != 20 {
		t.Error("Bad output len")
	}

	if bytes.Compare(buf, data[20:40]) != 0 {
		t.Error("Bad data", string(buf))
	}

}

func TestFileReadAt(t *testing.T) {
	f, err := fs.OpenFile("/testFileRead", os.O_RDWR|os.O_CREATE, os.FileMode(0664))
	if err != nil {
		t.Error(err)
	}

	data := []byte("short test data")

	f.(*file).inode.data = data

	buf := make([]byte, 20)

	_, err = f.ReadAt(buf, 100)
	if err == nil {
		t.Error("Bad error status")
	}
}

func TestFileReaddir(t *testing.T) {
	err := fs.MkdirAll("/testFileReaddir/1", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	err = fs.Mkdir("/testFileReaddir/3", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	err = fs.Mkdir("/testFileReaddir/2", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	err = fs.Mkdir("/testFileReaddir/4", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	f, err := fs.Open("/testFileReaddir")
	if err != nil {
		t.Error(err)
	}

	fi, err := f.Readdir(2)
	if err != nil {
		t.Error(err)
	}

	if len(fi) != 2 {
		t.Error("Bad result length")
	}
	if fi[0].Name() != "1" || fi[1].Name() != "2" {
		t.Error("Bad result content", fi[0].Name())
	}

	fi, err = f.Readdir(0)
	if err != nil {
		t.Error(err)
	}

	if len(fi) != 4 {
		t.Error("Bad result length")
	}

}

func TestFileReaddirnames(t *testing.T) {
	err := fs.MkdirAll("/testFileReaddirnames/1", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	err = fs.Mkdir("/testFileReaddirnames/3", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	err = fs.Mkdir("/testFileReaddirnames/2", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	err = fs.Mkdir("/testFileReaddirnames/4", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	f, err := fs.Open("/testFileReaddirnames")
	if err != nil {
		t.Error(err)
	}

	names, err := f.Readdirnames(2)
	if err != nil {
		t.Error(err)
	}

	if len(names) != 2 {
		t.Error("Bad result length")
	}
	if names[0] != "1" || names[1] != "2" {
		t.Error("Bad result content")
	}

	names, err = f.Readdirnames(0)
	if err != nil {
		t.Error(err)
	}

	if len(names) != 4 {
		t.Error("Bad result length")
	}

}

func TestFileSeek(t *testing.T) {
	f, err := fs.Create("/testFileSeek")
	if err != nil {
		t.Error(err)
	}

	data := []byte("long test data......................................................................................")
	f.(*file).inode.data = data

	p, err := f.Seek(1, 3)
	if err == nil {
		t.Error("Bad error status")
	}

	p, err = f.Seek(1, 2)
	if err == nil {
		t.Error("Bad error status")
	}

	p, err = f.Seek(-4, 2)
	if err != nil {
		t.Error(err)
	}

	if int(p) != len(data)-4 {
		t.Error("Bad position")
	}

	buf := make([]byte, 4)

	_, err = f.Read(buf)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(buf, []byte("....")) != 0 {
		t.Error("Bad data")
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		t.Error(err)
	}

	_, err = f.Read(buf)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(buf, []byte("long")) != 0 {
		t.Error("Bad data")
	}

	_, err = f.Seek(1, 1)
	if err != nil {
		t.Error(err)
	}

	_, err = f.Read(buf)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(buf, []byte("test")) != 0 {
		t.Error("Bad data")
	}

}

func TestFileStat(t *testing.T) {
	f, err := fs.Create("/testFileStat")
	if err != nil {
		t.Error(err)
	}

	data := []byte("short test data")
	f.(*file).inode.data = data

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}

	if fi.Size() != int64(len(data)) {
		t.Error("Bad size")
	}
}

func TestFileTruncate(t *testing.T) {
	f, err := fs.Create("/testFileTruncate")
	if err != nil {
		t.Error(err)
	}

	data := []byte("short test data")
	f.(*file).inode.data = data

	err = f.Truncate(4)
	if err != nil {
		t.Error(err)
	}

	if len(f.(*file).inode.data) != 4 {
		t.Error("Bad size")
	}
}

func TestFileWrite(t *testing.T) {
	f, err := fs.Create("/testFileWrite")
	if err != nil {
		t.Error(err)
	}

	n, err := f.Write([]byte("test data"))
	if err != nil {
		t.Error(err)
	}

	if n != 9 {
		t.Error("Bad length")
	}

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 9 {
		t.Error("Bad length")
	}

	n, err = f.Write([]byte("test data"))
	if err != nil {
		t.Error(err)
	}

	if n != 9 {
		t.Error("Bad length")
	}

	fi, err = f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 18 {
		t.Error("Bad length")
	}

	f.Seek(0, 0)
	n, err = f.Write([]byte("new stuff"))
	if err != nil {
		t.Error(err)
	}

	if n != 9 {
		t.Error("Bad length")
	}

	fi, err = f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 18 {
		t.Error("Bad length")
	}

	f.Seek(14, 0)
	n, err = f.Write([]byte("test data"))
	if err != nil {
		t.Error(err)
	}

	if n != 9 {
		t.Error("Bad length")
	}

	fi, err = f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 23 {
		t.Error("Bad length")
	}

}

func TestFileWriteAt(t *testing.T) {
	f, err := fs.Create("/testFileWriteAt")
	if err != nil {
		t.Error(err)
	}

	n, err := f.WriteAt([]byte("test data"), 100)
	if err != nil {
		t.Error(err)
	}

	if n != 9 {
		t.Error("Bad length")
	}

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 109 {
		t.Error("Bad length")
	}
}

func TestFileWriteString(t *testing.T) {
	f, err := fs.Create("/testFileWriteString")
	if err != nil {
		t.Error(err)
	}

	n, err := f.WriteString("test data")
	if err != nil {
		t.Error(err)
	}

	if n != 9 {
		t.Error("Bad length")
	}

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 9 {
		t.Error("Bad length")
	}
}
