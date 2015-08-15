package testfs

import (
	"code.google.com/p/go-uuid/uuid"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestParsePath(t *testing.T) {
	fs := NewTestFS()
	tmp := dentry{inode: 2}
	tmp.children = make(map[string]dentry)
	test := dentry{inode: 3}

	fs.dirTree.children["tmp"] = tmp
	fs.dirTree.children["tmp"].children["test"] = test

	terms, err := fs.parsePath("/")
	if err != nil || terms != nil {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 1 || terms[0] != "tmp" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 1 || terms[0] != "tmp" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/test")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/test//")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/./test/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/test/../test/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = fs.parsePath("/tmp/../../test/")
	if err != os.ErrNotExist {
		t.Error("Parse failure")
	}

	fs.cwd = "/tmp"
	terms, err = fs.parsePath("test")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

}

func BenchmarkParsePath(b *testing.B) {
	path := "/test/path/with/five/elements"

	fs := NewTestFS()

	for n := 0; n < b.N; n++ {
		_, err := fs.parsePath(path)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestLookupPath(t *testing.T) {
	fs := NewTestFS()
	tmp := dentry{inode: 2}
	tmp.children = make(map[string]dentry)
	test := dentry{inode: 3}

	fs.dirTree.children["tmp"] = tmp
	fs.dirTree.children["tmp"].children["test"] = test

	i, err := fs.lookupPath(nil)
	if err != nil {
		t.Error(err)
	}
	if i.inode != 1 {
		t.Error("Wrong root inode number")
	}

	i, err = fs.lookupPath([]string{"tmp"})
	if err != nil {
		t.Error(err)
	}
	if i.inode != 2 {
		t.Error("Wrong inode number")
	}

	i, err = fs.lookupPath([]string{"tmp", "test"})
	if err != nil {
		t.Error(err)
	}
	if i.inode != 3 {
		t.Error("Wrong inode number")
	}

}

func BenchmarkLookupPath(b *testing.B) {
	path := "/test/path/with/five/elements"

	fs := NewTestFS()
	err := fs.MkdirAll(path, os.FileMode(0775))
	if err != nil {
		b.Error(err)
	}

	terms, err := fs.parsePath(path)
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		_, err = fs.lookupPath(terms)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestCheckPerm(t *testing.T) {
	user := newInode(Uid, 666, os.FileMode(0000))
	group := newInode(666, Gid, os.FileMode(0000))
	all := newInode(666, 666, os.FileMode(0000))

	// Check failures
	if checkPerm(user, 'r') {
		t.Error("Permission check failed")
	}
	if checkPerm(user, 'w') {
		t.Error("Permission check failed")
	}
	if checkPerm(user, 'x') {
		t.Error("Permission check failed")
	}
	if checkPerm(group, 'r') {
		t.Error("Permission check failed")
	}
	if checkPerm(group, 'w') {
		t.Error("Permission check failed")
	}
	if checkPerm(group, 'x') {
		t.Error("Permission check failed")
	}
	if checkPerm(all, 'r') {
		t.Error("Permission check failed")
	}
	if checkPerm(all, 'w') {
		t.Error("Permission check failed")
	}
	if checkPerm(all, 'x') {
		t.Error("Permission check failed")
	}

	user.mode = os.FileMode(0700)
	group.mode = os.FileMode(0070)
	all.mode = os.FileMode(0007)

	// Check success
	if !checkPerm(user, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	if !checkPerm(group, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	if !checkPerm(all, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}

}

func BenchmarkCheckPerm(b *testing.B) {
	i := newInode(Uid, Gid, os.FileMode(0644))
	for n := 0; n < b.N; n++ {
		if !checkPerm(i, 'r', 'w') {
			b.Error("Permission check failed")
		}
	}
}

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
