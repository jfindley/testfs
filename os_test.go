package testfs

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// We just test the interface works and basic operations succeed.
// Most testing is done in the os package itself.
func TestOSFS(t *testing.T) {
	var testfs FileSystem
	testfs = NewOSFS()

	err := testfs.MkdirAll(os.TempDir(), os.FileMode(1777))
	if err != nil {
		t.Error(err)
	}

	f, err := testfs.Create(os.TempDir() + "/testfile")
	if err != nil {
		t.Error(err)
	}

	data := []byte("test")

	f.Write(data)

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 4 {
		t.Error("Bad filesize")
	}

	buf := make([]byte, 5)

	n, err := f.ReadAt(buf, 0)
	if err != io.EOF {
		t.Error("Bad error status")
	}

	if n != 4 {
		t.Error("Bad data length")
	}

	if bytes.Compare(data, buf[:n]) != 0 {
		t.Error("Bad file contents")
	}

	err = testfs.Remove(os.TempDir() + "/testfile")
	if err != nil {
		t.Error(err)
	}
}

// Make sure that TestFS works in the same way as OSFS.
func TestTestFS(t *testing.T) {
	var testfs FileSystem
	testfs = NewTestFS(0,0)

	err := testfs.MkdirAll(os.TempDir(), os.FileMode(1777))
	if err != nil {
		t.Error(err)
	}

	f, err := testfs.Create(os.TempDir() + "/testfile")
	if err != nil {
		t.Error(err)
	}

	data := []byte("test")

	f.Write(data)

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 4 {
		t.Error("Bad filesize")
	}

	buf := make([]byte, 5)

	n, err := f.ReadAt(buf, 0)
	if err != io.EOF {
		t.Error("Bad error status")
	}

	if n != 4 {
		t.Error("Bad data length")
	}

	if bytes.Compare(data, buf[:n]) != 0 {
		t.Error("Bad file contents")
	}

	err = testfs.Remove(os.TempDir() + "/testfile")
	if err != nil {
		t.Error(err)
	}
}
