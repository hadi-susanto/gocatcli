//go:build windows

/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Windows counterpart of mountpoint_ex.go: WinFsp mounts on a drive
letter or creates the mountpoint directory itself.
*/

package fuser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var driveLetterRe = regexp.MustCompile(`^[A-Za-z]:$`)

// PrepareMountpoint checks WinFsp is installed and the mountpoint is usable:
// a drive letter ("X:"), "*" (first free drive) or a directory that does
// not exist yet (its parents are created). The returned cleanup removes
// the parents it created, WinFsp removes the mountpoint directory itself.
func PrepareMountpoint(mountpoint string) (func(), error) {
	if !winfspInstalled() {
		return noCleanup, errWinFspMissing
	}

	if mountpoint == "*" || driveLetterRe.MatchString(mountpoint) {
		return noCleanup, nil
	}

	var _, err = os.Stat(mountpoint)
	if err == nil {
		return noCleanup, fmt.Errorf("%s already exists: on Windows mount on a drive letter (e.g. X:) "+
			"or on a directory that does not exist yet (WinFsp creates it)", mountpoint)
	}
	if !os.IsNotExist(err) {
		return noCleanup, err
	}

	var parent = filepath.Dir(filepath.Clean(mountpoint))
	var created []string
	created, err = mkdirTracked(parent)
	if err != nil {
		return noCleanup, fmt.Errorf("create mountpoint parent: %w", err)
	}
	return func() { removeCreated(created) }, nil
}
