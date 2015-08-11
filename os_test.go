package testfs

import (
	"os"
	"testing"
)

// We just test the interface works and basic operations succeed.
// Most testing is done in the os package itself.
func TestOSFS(t *testing.T) {
	fs := NewOSFS()

	f, err := fs.Create(os.TempDir() + "/testfile")
	if err != nil {
		t.Error(err)
	}

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	if fi.Size() != 0 {
		t.Error("Bad filesize")
	}

	err = fs.Remove(os.TempDir() + "/testfile")
	if err != nil {
		t.Error(err)
	}
}
