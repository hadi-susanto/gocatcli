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
// not exist yet (its parents are created)
func PrepareMountpoint(mountpoint string) error {
	if !winfspInstalled() {
		return errWinFspMissing
	}

	if mountpoint == "*" || driveLetterRe.MatchString(mountpoint) {
		return nil
	}

	var _, err = os.Stat(mountpoint)
	if err == nil {
		return fmt.Errorf("%s already exists: on Windows mount on a drive letter (e.g. X:) "+
			"or on a directory that does not exist yet (WinFsp creates it)", mountpoint)
	}
	if !os.IsNotExist(err) {
		return err
	}

	var parent = filepath.Dir(filepath.Clean(mountpoint))
	err = os.MkdirAll(parent, 0o755)
	if err != nil {
		return fmt.Errorf("create mountpoint parent: %w", err)
	}
	return nil
}
