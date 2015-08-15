package testfs

import (
	"os"
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
