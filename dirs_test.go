package testfs

import (
	"code.google.com/p/go-uuid/uuid"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMkdir(t *testing.T) {
	err := fs.Mkdir("/testmkdir", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/testmkdir")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkMkdir(b *testing.B) {
	fs = NewTestFS()
	fs = NewTestFS()
	for n := 0; n < b.N; n++ {
		err := fs.Mkdir("/"+uuid.New(), os.FileMode(0755))
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParallelMkdir(b *testing.B) {
	fs = NewTestFS()
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
	err := fs.MkdirAll("/test/mkdir/all", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	_, err = fs.find("/test/mkdir/all")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkMkdirAll(b *testing.B) {
	fs = NewTestFS()
	path := strings.Repeat("/bench/mkdir/all", 4)

	for n := 0; n < b.N; n++ {
		err := fs.MkdirAll("/benchmkdirall"+strconv.Itoa(n)+path, os.FileMode(0755))
		if err != nil {
			b.Error(err)
		}
	}
}

func TestChdir(t *testing.T) {
	err := fs.Chdir("/testchdir")
	if !os.IsNotExist(err) {
		t.Error("Bad error code")
	}
	if fs.cwd.name != "/" {
		t.Error("Wrong working dir")
	}

	err = fs.MkdirAll("/testchdir/test", os.FileMode(0777))
	if err != nil {
		t.Error(err)
	}

	err = fs.Chdir("/testchdir")
	if err != nil {
		t.Error(err)
	}
	if fs.cwd.name != "testchdir" {
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
	dir, err := fs.Getwd()
	if err != nil {
		t.Error(err)
	}
	if dir != fs.cwdPath {
		t.Error("Bad WD")
	}

	err = fs.MkdirAll("/testgetwd/test", os.FileMode(0777))
	if err != nil {
		t.Error(err)
	}

	fs.Chdir("/testgetwd/test")
	dir, err = fs.Getwd()
	if err != nil {
		t.Error(err)
	}
	if dir != "/testgetwd/test" {
		t.Error("Bad WD")
	}
}
