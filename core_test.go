package testfs

import (
	"flag"
	"os"
	"strings"
	"testing"
)

var fs *TestFS

func TestMain(m *testing.M) {
	flag.Parse()

	fs = NewTestFS(0,0)
	Uid = 0
	Gid = 0

	os.Exit(m.Run())
}

func TestParsePath(t *testing.T) {
	fs.MkdirAll("/tmp/test", os.FileMode(0777)|os.ModeDir)

	terms, err := parsePath("/")
	if err != nil || terms != nil {
		t.Error("Parse failure")
	}

	terms, err = parsePath("/tmp")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 1 || terms[0] != "tmp" {
		t.Error("Parse failure")
	}

	terms, err = parsePath("/tmp/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 1 || terms[0] != "tmp" {
		t.Error("Parse failure")
	}

	terms, err = parsePath("/tmp//test")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = parsePath("/tmp/test//")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = parsePath("/tmp///./test/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

}

func BenchmarkParsePath(b *testing.B) {
	fs = NewTestFS(0,0)
	path := "/test/path/with/five/elements"

	for n := 0; n < b.N; n++ {
		_, err := parsePath(path)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestCheckPerm(t *testing.T) {

	fs.dirTree.new("2", 100, 0, os.FileMode(0000))
	fs.dirTree.new("3", 0, 200, os.FileMode(0000))
	fs.dirTree.new("4", 0, 0, os.FileMode(0000))
	fs.dirTree.new("5", 100, 0, os.FileMode(0700))
	fs.dirTree.new("6", 0, 200, os.FileMode(0070))
	fs.dirTree.new("7", 0, 0, os.FileMode(0007))

	Uid = 100
	Gid = 200

	// Check failures
	i, err := fs.find("/2")
	if !os.IsPermission(err) {
		t.Error(err)
	}
	i, err = fs.find("/2")
	if !os.IsPermission(err) {
		t.Error(err)
	}
	i, err = fs.find("/2")
	if !os.IsPermission(err) {
		t.Error(err)
	}
	i, err = fs.find("/3")
	if !os.IsPermission(err) {
		t.Error(err)
	}
	i, err = fs.find("/3")
	if !os.IsPermission(err) {
		t.Error(err)
	}
	i, err = fs.find("/3")
	if !os.IsPermission(err) {
		t.Error(err)
	}
	i, err = fs.find("/4")
	if !os.IsPermission(err) {
		t.Error(err)
	}
	i, err = fs.find("/4")
	if !os.IsPermission(err) {
		t.Error(err)
	}
	i, err = fs.find("/4")
	if !os.IsPermission(err) {
		t.Error(err)
	}

	// Check success
	i, err = fs.find("/5")
	if err != nil {
		t.Error(err)
	}
	if !checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	i, err = fs.find("/6")
	if err != nil {
		t.Error(err)
	}
	if !checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	i, err = fs.find("/7")
	if err != nil {
		t.Error(err)
	}
	if !checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}

	Uid = 0
	Gid = 0

}

func BenchmarkCheckPerm(b *testing.B) {
	fs = NewTestFS(0,0)
	err := fs.dirTree.new("benchcheckperm", Uid, Gid, os.FileMode(0644))
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		if !checkPerm(fs.dirTree.children["benchcheckperm"], 'r', 'w') {
			b.Error("Permission check failed")
		}
	}
}

func TestFind(t *testing.T) {
	_, err := fs.find("/tmp/test/find")
	if !os.IsNotExist(err) {
		t.Error("Bad error status")
	}

	err = fs.MkdirAll("/tmp/test/find", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	in, err := fs.find("/tmp/test/find")
	if err != nil {
		t.Error(err)
	}
	if in.name != "find" {
		t.Error("Bad name")
	}

	in, err = fs.find("/tmp//test/../test/./find")
	if err != nil {
		t.Error(err)
	}
	if in.name != "find" {
		t.Error("Bad name")
	}
}

func BenchmarkFind(b *testing.B) {
	path := strings.Repeat("/testpath", 50)

	err := fs.MkdirAll(path, os.FileMode(0775))
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		_, err = fs.find(path)
		if err != nil {
			b.Error(err)
		}
	}
}
