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

	err := fs.Mkdir("/tmp", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	_, err = fs.lookupPath([]string{"tmp"})
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkMkdirDeep(b *testing.B) {
	fs := NewTestFS()

	path := strings.Repeat("/tmp", b.N)

	for n := 0; n < b.N; n++ {
		err := fs.Mkdir(path[:4*(n+1)], os.FileMode(0755))
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkMkdirWide(b *testing.B) {
	fs := NewTestFS()

	for n := 0; n < b.N; n++ {
		err := fs.Mkdir("/"+strconv.Itoa(n), os.FileMode(0755))
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParallelMkdirWide(b *testing.B) {
	fs := NewTestFS()

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

	_, err = fs.lookupPath([]string{"test", "path", "foo"})
	if err != nil {
		t.Error(err)
	}
}
