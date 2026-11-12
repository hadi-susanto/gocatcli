/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Mountpoint directories created by gocatcli are tracked so that only
those are removed after unmount, like a removable drive.
*/

package fuser

import (
	"os"
	"path/filepath"

	"github.com/deadc0de6/gocatcli/internal/log"
)

// mkdirTracked creates dir and its missing parents, returning the
// created directories from the topmost to dir itself. Nothing is left
// behind on failure.
func mkdirTracked(dir string) ([]string, error) {
	var missing []string
	var cur = filepath.Clean(dir)
	for {
		var _, err = os.Lstat(cur)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		missing = append(missing, cur)
		var parent = filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}

	var created = make([]string, 0, len(missing))
	for i := len(missing) - 1; i >= 0; i-- {
		var err = os.Mkdir(missing[i], 0o755)
		if err != nil {
			removeCreated(created)
			return nil, err
		}
		created = append(created, missing[i])
	}
	return created, nil
}

// removeCreated removes directories returned by mkdirTracked, deepest
// first. Only empty directories are removed, it stops at the first
// one that cannot be removed (its parents are then still needed).
func removeCreated(created []string) {
	for i := len(created) - 1; i >= 0; i-- {
		var err = os.Remove(created[i])
		if err != nil {
			log.Warnf("could not remove %s: %s", created[i], err)
			return
		}
		log.Infof("removed mountpoint directory %s", created[i])
	}
}
