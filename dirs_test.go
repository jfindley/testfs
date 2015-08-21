package testfs

import (
	"code.google.com/p/go-uuid/uuid"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMkdir(t *testing.T) {
	fs := NewTestFS()
	Uid = 0

	err := fs.Mkdir("/tmp", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/tmp")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkMkdir(b *testing.B) {
	fs := NewTestFS()
	Uid = 0

	for n := 0; n < b.N; n++ {
		err := fs.Mkdir("/"+strconv.Itoa(n), os.FileMode(0755))
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParallelMkdir(b *testing.B) {
	fs := NewTestFS()
	Uid = 0

	b.RunParallel(func(pb *testing.PB) {

		for pb.Next() {
			err := fs.Mkdir("/"+uuid.New(), os.FileMode(0755))
			if err != nil {
				b.Error(err)
			}

		}
	})
}

func TestMkdirAll(t *testing.T) {
	fs := NewTestFS()

	err := fs.MkdirAll("/test/path/foo", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/test/path/foo")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkMkdirAll(b *testing.B) {
	fs := NewTestFS()
	Uid = 0

	path := strings.Repeat("/tmp", 4)

	for n := 0; n < b.N; n++ {
		err := fs.MkdirAll("/"+strconv.Itoa(n)+path, os.FileMode(0755))
		if err != nil {
			b.Error(err)
		}
	}
}

func TestChdir(t *testing.T) {
	fs := NewTestFS()

	err := fs.Chdir("/tmp")
	if !os.IsNotExist(err) {
		t.Error("Bad error code")
	}
	if fs.cwd.name != "/" {
		t.Error("Wrong working dir")
	}

	err = fs.MkdirAll("/tmp/test", os.FileMode(0777))
	if err != nil {
		t.Error(err)
	}

	err = fs.Chdir("/tmp")
	if err != nil {
		t.Error(err)
	}
	if fs.cwd.name != "tmp" {
		t.Error("Wrong working dir")
	}

	err = fs.Chdir("test")
	if err != nil {
		t.Error(err)
	}
	if fs.cwd.name != "test" {
		t.Error("Wrong working dir")
	}
}

func TestGetwd(t *testing.T) {
	fs := NewTestFS()
	dir, err := fs.Getwd()
	if err != nil {
		t.Error(err)
	}
	if dir != "/" {
		t.Error("Bad WD")
	}

	err = fs.MkdirAll("/tmp/test", os.FileMode(0777))
	if err != nil {
		t.Error(err)
	}

	fs.Chdir("/tmp/test")
	dir, err = fs.Getwd()
	if err != nil {
		t.Error(err)
	}
	if dir != "/tmp/test" {
		t.Error("Bad WD")
	}
}
