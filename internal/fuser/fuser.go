/*
author: deadc0de6 (https://github.com/deadc0de6)
Copyright (c) 2024, deadc0de6
*/

package fuser

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/deadc0de6/gocatcli/internal/log"
	"github.com/deadc0de6/gocatcli/internal/tree"

	"github.com/anacrolix/fuse"
	"github.com/anacrolix/fuse/fs"
)

var (
	logPath = "/tmp/gocatcli-fuser.log"
)

// FS fuse filesystem
type FS struct {
	mountPoint string
	theTree    *tree.Tree
	root       *FuseDir
	debugMode  bool
}

// Root returns the root dir handle
func (f *FS) Root() (fs.Node, error) {
	rd := &FuseDir{
		fs:      f,
		theTree: f.theTree,
	}
	f.root = rd
	return rd, nil
}

// prepareMountpoint creates the mountpoint if missing
// and cleans up a stale mount left by a previous crash
func prepareMountpoint(mountpoint string) error {
	fi, err := os.Stat(mountpoint)
	switch {
	case err == nil:
		if !fi.IsDir() {
			return fmt.Errorf("%s is not a directory", mountpoint)
		}
		return nil
	case os.IsNotExist(err):
		if err := os.MkdirAll(mountpoint, 0o755); err != nil {
			return fmt.Errorf("create mountpoint: %w", err)
		}
		log.Infof("created mountpoint %s", mountpoint)
		return nil
	case errors.Is(err, syscall.ENOTCONN):
		log.Warnf("stale mount detected on %s, unmounting", mountpoint)
		if err := fuse.Unmount(mountpoint); err != nil {
			return fmt.Errorf("unmount stale mount (try \"fusermount -uz %s\"): %w", mountpoint, err)
		}
		return nil
	default:
		return err
	}
}

// unmountOnSignal unmounts the mountpoint on SIGINT/SIGTERM/SIGHUP,
// which makes fs.Serve return so the process exits cleanly
func unmountOnSignal(mountpoint string) func() {
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		if _, ok := <-sigs; !ok {
			return
		}
		log.Infof("unmounting %s", mountpoint)
		if err := fuse.Unmount(mountpoint); err != nil {
			log.Errorf("unmount failed (%v), mountpoint busy? send the signal again (Ctrl-C) to force a lazy unmount", err)
			if _, ok := <-sigs; !ok {
				return
			}
			if err := exec.Command("fusermount", "-u", "-z", mountpoint).Run(); err != nil {
				log.Errorf("lazy unmount failed: %v", err)
			}
		}
	}()
	return func() {
		signal.Stop(sigs)
		close(sigs)
	}
}

// Mount mount the tree
func Mount(theTree *tree.Tree, mountpoint string, debug bool) error {
	if err := prepareMountpoint(mountpoint); err != nil {
		return err
	}

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("gocatcli"),
		fuse.Subtype("gocatcli"),
	)
	if err != nil {
		return err
	}
	defer func() {
		// anacrolix/fuse leaves Conn.backendState nil on linux,
		// making Close() panic; the process is exiting anyway
		defer func() {
			if r := recover(); r != nil {
				log.Debugf("fuse close: %v", r)
			}
		}()
		err := c.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	myFS := &FS{
		mountPoint: mountpoint,
		theTree:    theTree,
		debugMode:  debug,
	}
	stop := unmountOnSignal(mountpoint)
	defer stop()
	err = fs.Serve(c, myFS)
	return err
}
