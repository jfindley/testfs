package testfs

import (
	"os"
	"path"
)

func (t *TestFS) Mkdir(name string, perm os.FileMode) error {
	// Ensure the dir mode is set
	perm |= os.ModeDir

	dir, err := t.find(path.Dir(name))
	if err != nil {
		return err
	}

	if !dir.IsDir() {
		return os.ErrInvalid
	}

	// Create the directory
	return dir.new(path.Base(name), Uid, Gid, perm)
}

func (t *TestFS) MkdirAll(name string, perm os.FileMode) error {
	// Ensure the dir mode is set
	perm |= os.ModeDir

	terms, err := parsePath(name)
	if err != nil {
		return err
	}

	var dir *inode

	if name[0] == '/' {
		dir = &t.dirTree
	} else {
		dir = t.cwd
	}

	return dir.mkdirAll(terms, perm)
}

func (i *inode) mkdirAll(terms []string, perm os.FileMode) error {
	if len(terms) == 0 {
		return nil
	}

	err := i.new(terms[0], Uid, Gid, perm)
	if len(terms) == 1 {
		if os.IsExist(err) {
			return nil
		}
		return err
	}

	dir := i.children[terms[0]]

	switch {

	case err == nil:
		return dir.mkdirAll(terms[1:], perm)

	case os.IsExist(err):
		// If the child is not a directory, fail
		if !i.children[terms[0]].IsDir() {
			return err
		}
		// If it is a directory, just continue
		return dir.mkdirAll(terms[1:], perm)

	default:
		// Some other error
		return err

	}

}

func (t *TestFS) Chdir(dir string) error {

	d, err := t.find(dir)
	if err != nil {
		return err
	}

	if !d.IsDir() {
		return os.ErrInvalid
	}

	if !checkPerm(d, 'r', 'x') {
		return os.ErrPermission
	}

	t.cwd = d

	switch {

	case dir[0] == '/':
		t.cwdPath = dir

	case t.cwdPath[len(t.cwdPath)-1] == '/':
		t.cwdPath = t.cwdPath + dir

	default:
		t.cwdPath = t.cwdPath + "/" + dir

	}

	return nil
}

func (t *TestFS) Getwd() (dir string, err error) {
	return t.cwdPath, nil
}
