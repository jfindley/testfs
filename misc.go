package testfs

import (
	"os"
	"path"
)

func (t *TestFS) Chmod(name string, mode os.FileMode) error {
	f, err := t.find(name)
	if err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Blank out existing permission bits
	this := f.mode
	this = this >> 10
	this = this << 10

	this = this &^ os.ModeSetuid
	this = this &^ os.ModeSetgid
	this = this &^ os.ModeSticky

	// Set new permission bits
	this |= mode

	f.mode = this

	return nil
}

func (t *TestFS) Chown(name string, uid, gid int) error {
	f, err := t.find(name)
	if err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	this := f
	this.uid = uint16(uid)
	this.gid = uint16(gid)

	f = this

	return nil
}

func (t *TestFS) Link(oldname, newname string) error {
	tar, err := t.find(oldname)
	if err != nil {
		return err
	}

	// No hardlinks to directories
	if tar.mode&os.ModeDir == os.ModeDir {
		return os.ErrInvalid
	}

	dir, err := t.find(path.Dir(newname))
	if err != nil {
		return err
	}

	dir.mu.Lock()
	defer dir.mu.Unlock()

	if _, ok := dir.children[path.Base(newname)]; ok {
		return os.ErrExist
	}

	dir.children[path.Base(newname)] = tar
	tar.linkCount++

	return nil
}

func (t *TestFS) Readlink(name string) (string, error) {
	f, err := t.find(name)
	if err != nil {
		return "", err
	}

	if f.mode&os.ModeSymlink == 0 || f.relName == "" {
		return "", os.ErrInvalid
	}

	return f.relName, nil

}

func unlink(in *inode) {
	in.mu.Lock()
	defer in.mu.Unlock()

	if in.linkCount <= 1 {
		in = nil
	} else {
		in.linkCount--
	}
	return
}

func (t *TestFS) Remove(name string) error {

	d, err := t.find(path.Dir(name))
	if err != nil {
		return err
	}

	if !checkPerm(d, 'r', 'w', 'x') {
		return os.ErrPermission
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := d.lookup([]string{path.Base(name)})
	if err != nil {
		return err
	}

	if !checkPerm(f, 'w') {
		return os.ErrPermission
	}

	unlink(f)
	delete(d.children, f.name)
	return nil

}

func (t *TestFS) RemoveAll(path string) error {
	return nil
}

func (t *TestFS) Rename(oldpath, newpath string) error {
	return nil
}

func (t *TestFS) Symlink(oldname, newname string) error {
	return nil
}

func (t *TestFS) Lstat(path string) (os.FileInfo, error) {
	return nil, nil
}

func (t *TestFS) Stat(path string) (os.FileInfo, error) {
	return nil, nil
}
