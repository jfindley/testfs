package testfs

import (
	"os"
	"strings"
	"testing"
)

func TestParsePath(t *testing.T) {
	fs := NewTestFS()
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

	terms, err = parsePath("/tmp/test")
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

	terms, err = parsePath("/tmp/./test/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = parsePath("/tmp/test/../test/")
	if err != nil {
		t.Error("Parse failure")
	}
	if len(terms) != 2 || terms[0] != "tmp" || terms[1] != "test" {
		t.Error("Parse failure")
	}

	terms, err = parsePath("/tmp/../../test/")
	if err != os.ErrNotExist {
		t.Error("Parse failure")
	}

}

func BenchmarkParsePath(b *testing.B) {
	path := "/test/path/with/five/elements"

	for n := 0; n < b.N; n++ {
		_, err := parsePath(path)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestCheckPerm(t *testing.T) {
	fs := NewTestFS()

	fs.dirTree.new("2", Uid, 666, os.FileMode(0000))
	fs.dirTree.new("3", 666, Gid, os.FileMode(0000))
	fs.dirTree.new("4", 666, 666, os.FileMode(0000))
	fs.dirTree.new("5", Uid, 666, os.FileMode(0700))
	fs.dirTree.new("6", 666, Gid, os.FileMode(0070))
	fs.dirTree.new("7", 666, 666, os.FileMode(0007))

	// Check failures
	if i, _ := fs.find("/2"); checkPerm(i, 'r') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/2"); checkPerm(i, 'w') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/2"); checkPerm(i, 'x') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/3"); checkPerm(i, 'r') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/3"); checkPerm(i, 'w') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/3"); checkPerm(i, 'x') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/4"); checkPerm(i, 'r') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/4"); checkPerm(i, 'w') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/4"); checkPerm(i, 'x') {
		t.Error("Permission check failed")
	}

	// Check success
	if i, _ := fs.find("/5"); !checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/6"); !checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}
	if i, _ := fs.find("/7"); !checkPerm(i, 'r', 'w', 'x') {
		t.Error("Permission check failed")
	}

}

func BenchmarkCheckPerm(b *testing.B) {
	Uid = 0
	fs := NewTestFS()
	err := fs.dirTree.new("2", Uid, Gid, os.FileMode(0644))
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		if !checkPerm(fs.dirTree.children["2"], 'r', 'w') {
			b.Error("Permission check failed")
		}
	}
}

func TestFind(t *testing.T) {
	fs := NewTestFS()

	_, err := fs.find("/tmp/test")
	if !os.IsNotExist(err) {
		t.Error("Bad error status")
	}

	err = fs.MkdirAll("/tmp/test", os.FileMode(0755))
	if err != nil {
		t.Error(err)
	}

	in, err := fs.find("/tmp/test")
	if err != nil {
		t.Error(err)
	}
	if in.name != "test" {
		t.Error("Bad name")
	}
}

func BenchmarkFind(b *testing.B) {
	Uid = 0
	path := strings.Repeat("/testpath", 50)

	fs := NewTestFS()
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
