package testfs

import (
	"bytes"
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

	_, err := fs.Create("/testCreate")
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/testCreate")
	if err != nil {
		t.Error(err)
	}

}

func TestOpen(t *testing.T) {

}

func TestOpenFile(t *testing.T) {

}
