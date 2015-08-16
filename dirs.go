package testfs

import (
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
		d, err := t.lookupPath(terms[:i+1])
		if os.IsNotExist(err) {

			// Create the directory
			_, err = t.create(dir, terms[i], perm)
			if err != nil {
				return err
			}

			dir, err = t.lookupPath(terms[:i+1])
			if err != nil {
				// This should not happen
				return err
			}

		} else if err != nil {
			return err
		} else {
			dir = d
		}
	}

	return nil
}

func (t *TestFS) Chdir(dir string) error {
	terms, err := t.parsePath(dir)
	if err != nil {
		return err
	}
	_, err = t.lookupPath(terms)
	if err != nil {
		return err
	}
	if dir[0:1] == "/" {
		t.cwd = dir
	} else {
		t.cwd = t.cwd + "/" + dir
	}
	return nil
}
