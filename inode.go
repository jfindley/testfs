package testfs

import (
	"os"
	"path"
	"time"
)

// Methods to implement os.FileInfo
func (i *inode) Name() string {
	return i.name
}

func (i *inode) Size() int64 {
	return int64(len(i.data))
}

func (i *inode) Mode() os.FileMode {
	return i.mode
}

func (i *inode) ModTime() time.Time {
	return i.mtime
}

func (i *inode) IsDir() bool {
	if i.mode&os.ModeDir == 0 {
		return false
	}
	return true
}

func (i *inode) Sys() interface{} {
	return &Stat_t{
		Name:     i.name,
		Uid:      i.uid,
		Gid:      i.gid,
		Mode:     i.mode,
		Xattrs:   i.xattrs,
		Mtime:    i.mtime,
		Linkname: i.relName,
	}
}

func (i *inode) chmod(mode os.FileMode) error {
	if !checkPerm(i, 'r') {
		return os.ErrPermission
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	// Blank out existing permission bits
	perm := i.mode
	perm = perm >> 10
	perm = perm << 10

	perm = perm &^ os.ModeSetuid
	perm = perm &^ os.ModeSetgid
	perm = perm &^ os.ModeSticky

	// Set new permission bits
	perm |= mode

	i.mode = perm
	return nil
}

func (i *inode) chown(uid, gid int) error {
	if !checkPerm(i, 'r') {
		return os.ErrPermission
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	i.uid = uint16(uid)
	i.gid = uint16(gid)
	return nil
}

func (t *TestFS) Chmod(name string, mode os.FileMode) error {
	f, err := t.find(name)
	if err != nil {
		return err
	}

	return f.chmod(mode)
}

func (t *TestFS) Chown(name string, uid, gid int) error {
	f, err := t.find(name)
	if err != nil {
		return err
	}

	return f.chown(uid, gid)
}

func (t *TestFS) Link(oldname, newname string) error {

	newDir, newFile := path.Split(newname)

	tar, err := t.find(oldname)
	if err != nil {
		return err
	}

	// No hardlinks to directories
	if tar.IsDir() {
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
	dir.mtime = time.Now()
	tar.linkCount++

	return nil
}

func (t *TestFS) Readlink(name string) (string, error) {
	dir, file := path.Split(name)

	d, err := t.find(dir)
	if err != nil {
		return "", err
	}

	f, err := d.lookupSymlink(file)

	return f.relName, err

}

func (t *TestFS) Remove(name string) error {
	return t.rmHelper(name, unlink)

}

func (t *TestFS) RemoveAll(name string) error {
	return t.rmHelper(name, unlinkall)
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

	srcDir.mtime = time.Now()

	if srcDir != dstDir {
		dstDir.mu.Lock()
		defer dstDir.mu.Unlock()

		dstDir.mtime = time.Now()
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

	srcDir.mtime = time.Now()

	return nil
}

func (t *TestFS) Lstat(name string) (os.FileInfo, error) {
	dir, file := path.Split(name)

	d, err := t.find(dir)
	if err != nil {
		return nil, err
	}

	return d.lookupSymlink(file)
}

func (t *TestFS) Stat(path string) (os.FileInfo, error) {
	return t.find(path)
}

func unlink(in *inode) {
	in.mu.Lock()
	defer in.mu.Unlock()

	in.mtime = time.Now()

	in.linkCount--

	if in.linkCount == 0 {
		in = nil
	}
	return
}

func unlinkall(in *inode) {
	in.mu.Lock()

	in.mtime = time.Now()

	if !in.IsDir() {
		in.mu.Unlock()
		unlink(in)
		return
	}

	for name, child := range in.children {
		if name == ".." {
			continue
		}
		unlinkall(child)
	}

	in.mu.Unlock()
	unlink(in)
	return
}

func (t *TestFS) rmHelper(name string, unlinkfunc func(*inode)) error {
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

	unlinkfunc(f)
	delete(d.children, file)
	d.mtime = time.Now()
	return nil
}
