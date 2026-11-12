//go:build !windows

/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Kept in its own file so upstream changes in fuser.go never conflict.
*/

package fuser

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/deadc0de6/gocatcli/internal/log"

	"github.com/anacrolix/fuse"
)

// PrepareMountpoint creates the mountpoint if missing and cleans up
// a stale mount left by a previous crash or forced exit. The returned
// cleanup removes the directories it created, to be called after unmount.
func PrepareMountpoint(mountpoint string) (func(), error) {
	var fi, err = os.Stat(mountpoint)
	switch {
	case err == nil:
		if !fi.IsDir() {
			return noCleanup, fmt.Errorf("%s is not a directory", mountpoint)
		}
		// stat may still be served from the kernel cache on a dead
		// fuse mount, reading the directory tells for sure
		if isStaleMount(mountpoint) {
			return noCleanup, unmountStale(mountpoint)
		}
		return noCleanup, nil
	case os.IsNotExist(err):
		var created []string
		created, err = mkdirTracked(mountpoint)
		if err != nil {
			return noCleanup, fmt.Errorf("create mountpoint: %w", err)
		}
		log.Infof("created mountpoint %s", mountpoint)
		return func() { removeCreated(created) }, nil
	case errors.Is(err, syscall.ENOTCONN):
		return noCleanup, unmountStale(mountpoint)
	default:
		return noCleanup, err
	}
}

// isStaleMount returns true if the mountpoint is a fuse mount
// whose server is gone ("Transport endpoint is not connected")
func isStaleMount(mountpoint string) bool {
	var f, err = os.Open(mountpoint)
	if err == nil {
		_, err = f.Readdirnames(1)
		_ = f.Close()
	}
	return errors.Is(err, syscall.ENOTCONN)
}

func unmountStale(mountpoint string) error {
	log.Warnf("stale mount detected on %s, unmounting", mountpoint)
	var err = fuse.Unmount(mountpoint)
	if err != nil {
		return fmt.Errorf("unmount stale mount (try \"fusermount -uz %s\"): %w", mountpoint, err)
	}
	return nil
}
