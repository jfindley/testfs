package testfs

import (
	"errors"
	"os"
	"path"
)

func (t *TestFS) Mkdir(name string, perm os.FileMode) error {
	// Ensure the dir mode is set
	perm |= os.ModeDir

	d, err := t.findDentry(path.Dir(name))
	if err != nil {
		return err
	}

	// Create the directory
	_, err = t.create(d, path.Base(name), perm)
	return err
}

func (t *TestFS) MkdirAll(name string, perm os.FileMode) error {
	// Ensure the dir mode is set
	perm |= os.ModeDir

	terms, err := t.parsePath(name)
	if err != nil {
		return err
	}

	// Don't try to create root
	if terms == nil {
		return nil
	}

	dir := &t.dirTree

	for i := range terms {
		d := dir.lookup(terms[i])
		if d == nil {
			// Create the directory
			_, err = t.create(dir, terms[i], perm)
			if err != nil {
				return err
			}

			dir = dir.lookup(terms[i])
			if dir == nil {
				// Very, very unlikely race, catch it anyway
				return errors.New("Unexpected error")
			}
		} else {
			dir = d
		}
	}

	return nil
}

func (t *TestFS) Chdir(dir string) error {
	Uid = 0
	d, err := t.lookupPath(dir)
	if err != nil {
		return err
	}
	if t.lookupInode(d.inode).mode&os.ModeDir == 0 {
		return os.ErrInvalid
	}
	t.cwd = d
	if dir[0] == '/' {
		t.cwdPath = dir
	} else {
		t.cwdPath = t.cwdPath + "/" + dir
	}
	return nil
}

func (t *TestFS) Getwd() (dir string, err error) {
	return t.cwdPath, nil
}
