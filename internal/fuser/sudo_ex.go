//go:build !windows

/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Makes a mount done through sudo usable by the user who ran sudo.
*/

package fuser

import (
	"os"
	"strconv"

	"github.com/deadc0de6/gocatcli/internal/log"
)

// sudoInvoker returns the user who ran gocatcli through sudo
func sudoInvoker() (int, int, bool) {
	if os.Geteuid() != 0 {
		return 0, 0, false
	}
	var uid, errUID = strconv.Atoi(os.Getenv("SUDO_UID"))
	var gid, errGID = strconv.Atoi(os.Getenv("SUDO_GID"))
	if errUID != nil || errGID != nil {
		return 0, 0, false
	}
	return uid, gid, true
}

// chownToInvoker gives a created directory to the user who ran sudo
func chownToInvoker(dir string) {
	var uid, gid, ok = sudoInvoker()
	if !ok {
		return
	}
	var err = os.Lchown(dir, uid, gid)
	if err != nil {
		log.Warnf("could not change owner of %s: %s", dir, err)
	}
}
