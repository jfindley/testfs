package testfs

import (
	"os"
	"path"
)

func (t *TestFS) Chmod(name string, mode os.FileMode) error {
	in, err := t.find(name)
	if err != nil {
		return err
	}

	t.files[in].mu.Lock()
	defer t.files[in].mu.Unlock()

	// Blank out existing permission bits
	this := t.files[in]
	this.mode = this.mode >> 10
	this.mode = this.mode << 10

	this.mode = this.mode &^ os.ModeSetuid
	this.mode = this.mode &^ os.ModeSetgid
	this.mode = this.mode &^ os.ModeSticky

	// Set new permission bits
	this.mode |= mode

	t.files[in] = this

	return nil
}

func (t *TestFS) Chown(name string, uid, gid int) error {
	in, err := t.find(name)
	if err != nil {
		return err
	}

	t.files[in].mu.Lock()
	defer t.files[in].mu.Unlock()

	this := t.files[in]
	this.uid = uint16(uid)
	this.gid = uint16(gid)

	t.files[in] = this

	return nil
}

func (t *TestFS) Link(oldname, newname string) error {
	in, err := t.find(oldname)
	if err != nil {
		return err
	}
	dir, err := t.findDentry(path.Dir(newname))
	if err != nil {
		return err
	}

	rel := t.lookupInode(in)
	if rel == nil {
		return os.ErrNotExist
	}

	if rel.mode&os.ModeDir != 0 {
		// No hardlinks to directories
		return os.ErrInvalid
	}

	err = dir.newDentry(in, path.Base(newname))
	if err != nil {
		return err
	}
	rel.linkCount++

	return nil
}

func (t *TestFS) Getwd() (dir string, err error) {
	return t.cwd, nil
}

func (t *TestFS) Readlink(name string) (string, error) {
	in, err := t.find(name)
	if err != nil {
		return "", err
	}

	link := t.lookupInode(in)
	if link == nil {
		return "", os.ErrNotExist
	}

	if link.mode&os.ModeSymlink == 0 || link.rel == "" {
		return "", os.ErrInvalid
	}

	return link.rel, nil

}

func (t *TestFS) Remove(name string) error {

	dir, err := t.findDentry(path.Dir(name))
	if err != nil {
		return err
	}

	dir.mu.Lock()
	defer dir.mu.Unlock()

	file := dir.lookup(path.Base(name))
	if file == nil {
		return os.ErrNotExist
	}

	err = t.unlink(file.inode)
	if err != nil {
		return err
	}

	delete(dir.children, path.Base(name))
	return nil

}

func (t *TestFS) unlink(in inum) error {
	t.Lock()
	defer t.Unlock()
	for i := range t.files {

		if t.files[i].id == in {

			if !checkPerm(&t.files[i], 'w') {
				return os.ErrPermission
			}

			if t.files[i].linkCount <= 1 {
				t.files[i] = t.files[len(t.files)-1]
				t.files = t.files[:len(t.files)-1]

			} else {
				t.files[i].linkCount--
			}

			return nil
		}
	}

	return os.ErrNotExist
}
