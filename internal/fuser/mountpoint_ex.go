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
// a stale mount left by a previous crash or forced exit
func PrepareMountpoint(mountpoint string) error {
	var fi, err = os.Stat(mountpoint)
	switch {
	case err == nil:
		if !fi.IsDir() {
			return fmt.Errorf("%s is not a directory", mountpoint)
		}
		// stat may still be served from the kernel cache on a dead
		// fuse mount, reading the directory tells for sure
		if isStaleMount(mountpoint) {
			return unmountStale(mountpoint)
		}
		return nil
	case os.IsNotExist(err):
		err = os.MkdirAll(mountpoint, 0o755)
		if err != nil {
			return fmt.Errorf("create mountpoint: %w", err)
		}
		log.Infof("created mountpoint %s", mountpoint)
		return nil
	case errors.Is(err, syscall.ENOTCONN):
		return unmountStale(mountpoint)
	default:
		return err
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
