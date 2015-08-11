package testfs

import (
	"os"
)

// attrs describes the basic attributes of a file or directory
type attrs struct {
	uid    uint16
	gid    uint16
	mode   os.FileMode
	xattrs *map[string]string
}

func (a attrs) chmod(mode os.FileMode) {
	a.mode = mode
}

func (a attrs) chown(uid, gid int) {
	a.uid = uint16(uid)
	a.gid = uint16(gid)
}
