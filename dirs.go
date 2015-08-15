package testfs

import (
	"os"
)

func (t *TestFS) Mkdir(name string, perm os.FileMode) error {
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

	d, err := t.lookupPath(terms[:len(terms)-1])
	if err != nil {
		return err
	}

	d.Lock()
	defer d.Unlock()

	// Fail if dir exists
	if d.lookup(terms[len(terms)-1]) != nil {
		return os.ErrExist
	}

	// Create the directory
	i := t.newInum()
	d.children[terms[len(terms)-1]] = *newDentry(i)
	t.files[i] = newInode(Uid, Gid, perm)

	return nil
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

	dir := t.dirTree

	for i := range terms {
		d, err := t.lookupPath(terms[:i+1])
		if os.IsNotExist(err) {

			// Create the directory
			newInum := t.newInum()

			dir.children[terms[i]] = *newDentry(newInum)
			dir = dir.children[terms[i]]
			t.files[newInum] = newInode(Uid, Gid, perm)

		} else if err != nil {
			return err
		} else {
			dir = *d
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
