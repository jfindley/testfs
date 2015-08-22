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

	newDir, newFile := path.Split(newname)

	tar, err := t.find(oldname)
	if err != nil {
		return err
	}

	// No hardlinks to directories
	if tar.mode&os.ModeDir == os.ModeDir {
		return os.ErrInvalid
	}

	dir, err := t.find(newDir)
	if err != nil {
		return err
	}

	dir.mu.Lock()
	defer dir.mu.Unlock()

	if _, ok := dir.children[newFile]; ok {
		return os.ErrExist
	}

	dir.children[newFile] = tar
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

func (t *TestFS) Remove(name string) error {

	dir, file := path.Split(name)

	d, err := t.find(dir)
	if err != nil {
		return err
	}

	if !checkPerm(d, 'r', 'w', 'x') {
		return os.ErrPermission
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := d.lookup([]string{file})
	if err != nil {
		return err
	}

	if !checkPerm(f, 'w') {
		return os.ErrPermission
	}

	unlink(f)
	delete(d.children, file)
	return nil

}

func (t *TestFS) RemoveAll(name string) error {

	dir, file := path.Split(name)

	d, err := t.find(dir)
	if err != nil {
		return err
	}

	if !checkPerm(d, 'r', 'w', 'x') {
		return os.ErrPermission
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := d.lookup([]string{file})
	if err != nil {
		return err
	}

	if !checkPerm(f, 'w') {
		return os.ErrPermission
	}

	unlinkall(f)
	delete(d.children, file)
	return nil
}

func (t *TestFS) Rename(oldpath, newpath string) error {

	newDir, newFile := path.Split(newpath)
	oldDir, oldFile := path.Split(oldpath)

	srcDir, err := t.find(oldDir)
	if err != nil {
		return err
	}

	if !checkPerm(srcDir, 'r', 'w', 'x') {
		return os.ErrPermission
	}

	srcDir.mu.Lock()
	defer srcDir.mu.Unlock()

	src, err := srcDir.lookup([]string{oldFile})
	if err != nil {
		return err
	}

	dstDir, err := t.find(newDir)
	if err != nil {
		return err
	}

	if !checkPerm(dstDir, 'r', 'w', 'x') {
		return os.ErrPermission
	}

	if srcDir != dstDir {
		dstDir.mu.Lock()
		defer dstDir.mu.Unlock()
	}

	_, err = dstDir.lookup([]string{newFile})
	if !os.IsNotExist(err) {
		return os.ErrExist
	}

	dstDir.children[newFile] = src
	delete(srcDir.children, oldFile)

	return nil
}

func (t *TestFS) Symlink(oldname, newname string) error {

	newDir, newFile := path.Split(newname)

	dst, err := t.find(oldname)
	if err != nil {
		return err
	}

	srcDir, err := t.find(path.Dir(newname))
	if err != nil {
		return err
	}

	if !checkPerm(srcDir, 'r', 'w', 'x') {
		return os.ErrPermission
	}

	_, err = srcDir.lookup([]string{newDir})
	if !os.IsNotExist(err) {
		return os.ErrExist
	}

	err = srcDir.new(newFile, Uid, Gid, os.FileMode(0777)|os.ModeSymlink)
	if err != nil {
		return err
	}

	srcDir.children[newFile].rel = dst
	srcDir.children[newFile].relName = oldname

	return nil
}

func (t *TestFS) Lstat(path string) (os.FileInfo, error) {
	return nil, nil
}

func (t *TestFS) Stat(path string) (os.FileInfo, error) {
	return nil, nil
}

func unlink(in *inode) {
	in.mu.Lock()
	defer in.mu.Unlock()

	in.linkCount--

	if in.linkCount == 0 {
		in = nil
	}
	return
}

func unlinkall(in *inode) {
	in.mu.Lock()
	if in.mode&os.ModeDir == 0 {
		in.mu.Unlock()
		unlink(in)
		return
	}
	for _, child := range in.children {
		unlinkall(child)
	}
	in.mu.Unlock()
	unlink(in)
	return
}
