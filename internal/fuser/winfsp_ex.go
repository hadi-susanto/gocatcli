//go:build windows

/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Windows mount backend: serves the catalog through WinFsp
(https://github.com/winfsp/winfsp) using cgofuse
(https://github.com/winfsp/cgofuse), built without cgo.
*/

package fuser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deadc0de6/gocatcli/internal/helpers"
	"github.com/deadc0de6/gocatcli/internal/log"
	"github.com/deadc0de6/gocatcli/internal/node"
	"github.com/deadc0de6/gocatcli/internal/tree"

	"github.com/winfsp/cgofuse/fuse"
)

const dirSize = 512

var winfspLogPath = filepath.Join(os.TempDir(), "gocatcli-fuser.log")

// winfspFS read-only view of the catalog: storages at the root,
// their directories and files below; files have no content
type winfspFS struct {
	fuse.FileSystemBase
	theTree   *tree.Tree
	debugMode bool
	mountTime time.Time
}

func (f *winfspFS) debug(format string, a ...interface{}) {
	if f.debugMode {
		log.ToFile(winfspLogPath, fmt.Sprintf(format, a...))
	}
}

// isVisible archived files are skipped, as with the fuse backend
func isVisible(n node.Node) bool {
	switch n.GetType() {
	case node.FileTypeStorage, node.FileTypeDir, node.FileTypeFile, node.FileTypeArchive:
		return true
	}
	return false
}

func isDirNode(n node.Node) bool {
	return n.GetType() == node.FileTypeStorage || n.GetType() == node.FileTypeDir
}

// resolve returns the node at path, nil for the root;
// names are matched literally (no glob)
func (f *winfspFS) resolve(path string) (node.Node, bool) {
	var trimmed = strings.Trim(path, "/")
	if trimmed == "" {
		return nil, true
	}

	var parts = strings.Split(trimmed, "/")
	var storage = f.theTree.GetStorageByName(parts[0])
	if storage == nil {
		return nil, false
	}

	var current node.Node = storage
	for _, part := range parts[1:] {
		var child, found = current.GetDirectChildren()[part]
		if !found || child == nil || !isVisible(child) {
			return nil, false
		}
		current = child
	}
	return current, true
}

func (f *winfspFS) fillRootStat(stat *fuse.Stat_t) {
	var ts = fuse.NewTimespec(f.mountTime)
	stat.Mode = fuse.S_IFDIR | 0o755
	stat.Nlink = 2
	stat.Ino = 1
	stat.Size = dirSize
	stat.Atim = ts
	stat.Mtim = ts
	stat.Ctim = ts
	stat.Birthtim = ts
}

func fillNodeStat(n node.Node, stat *fuse.Stat_t) {
	var ts = fuse.NewTimespec(time.Unix(n.GetMAccess(), 0))
	var perm = uint32(helpers.ModeStrToInt(n.GetMode()))
	if isDirNode(n) {
		stat.Mode = fuse.S_IFDIR | perm
		stat.Nlink = 2
		stat.Size = dirSize
	} else {
		stat.Mode = fuse.S_IFREG | perm
		stat.Nlink = 1
		stat.Size = int64(n.GetSize())
	}
	stat.Ino = helpers.HashString64(n.GetPath())
	stat.Atim = ts
	stat.Mtim = ts
	stat.Ctim = ts
	stat.Birthtim = ts
}

// Statfs volume information
func (f *winfspFS) Statfs(_ string, stat *fuse.Statfs_t) int {
	stat.Bsize = 4096
	stat.Frsize = 4096
	stat.Namemax = 255
	return 0
}

// Getattr file and directory attributes
func (f *winfspFS) Getattr(path string, stat *fuse.Stat_t, _ uint64) int {
	f.debug("getattr %s", path)
	var n, found = f.resolve(path)
	if !found {
		return -fuse.ENOENT
	}
	if n == nil {
		f.fillRootStat(stat)
		return 0
	}
	fillNodeStat(n, stat)
	return 0
}

// Opendir only checks the directory exists
func (f *winfspFS) Opendir(path string) (int, uint64) {
	f.debug("opendir %s", path)
	var n, found = f.resolve(path)
	if !found {
		return -fuse.ENOENT, ^uint64(0)
	}
	if n != nil && !isDirNode(n) {
		return -fuse.ENOTDIR, ^uint64(0)
	}
	return 0, 0
}

// Releasedir nothing to release
func (f *winfspFS) Releasedir(_ string, _ uint64) int {
	return 0
}

// Readdir lists storages at the root, children below
func (f *winfspFS) Readdir(path string, fill func(name string, stat *fuse.Stat_t, ofst int64) bool, _ int64, _ uint64) int {
	f.debug("readdir %s", path)
	var n, found = f.resolve(path)
	if !found {
		return -fuse.ENOENT
	}

	fill(".", nil, 0)
	fill("..", nil, 0)

	if n == nil {
		for _, storage := range f.theTree.GetStorages() {
			var stat fuse.Stat_t
			fillNodeStat(storage, &stat)
			if !fill(storage.GetName(), &stat, 0) {
				break
			}
		}
		return 0
	}

	if !isDirNode(n) {
		return -fuse.ENOTDIR
	}
	for _, child := range n.GetSortedDirectChildren() {
		if !isVisible(child) {
			continue
		}
		var stat fuse.Stat_t
		fillNodeStat(child, &stat)
		if !fill(child.GetName(), &stat, 0) {
			break
		}
	}
	return 0
}

// Open read-only access to existing files
func (f *winfspFS) Open(path string, flags int) (int, uint64) {
	f.debug("open %s flags %#x", path, flags)
	var n, found = f.resolve(path)
	if !found {
		return -fuse.ENOENT, ^uint64(0)
	}
	if n == nil || isDirNode(n) {
		return -fuse.EISDIR, ^uint64(0)
	}
	if flags&fuse.O_ACCMODE != fuse.O_RDONLY {
		return -fuse.EROFS, ^uint64(0)
	}
	return 0, 0
}

// Read the catalog stores metadata only: files are empty
func (f *winfspFS) Read(_ string, _ []byte, _ int64, _ uint64) int {
	return 0
}

// Release nothing to release
func (f *winfspFS) Release(_ string, _ uint64) int {
	return 0
}

// Mount mount the tree through WinFsp, blocks until unmounted (Ctrl-C)
func Mount(theTree *tree.Tree, mountpoint string, debug bool) error {
	if !winfspInstalled() {
		return errWinFspMissing
	}

	var myFS = &winfspFS{
		theTree:   theTree,
		debugMode: debug,
		mountTime: time.Now(),
	}
	var host = fuse.NewFileSystemHost(myFS)
	host.SetCapReaddirPlus(true)

	// uid/gid -1: owned by the user running gocatcli-ex
	var opts = []string{"-o", "uid=-1,gid=-1,volname=gocatcli"}
	log.Infof("mounting on %s, press ctrl-c to unmount", mountpoint)
	if !host.Mount(mountpoint, opts) {
		return fmt.Errorf("could not mount on %s (see WinFsp output above)", mountpoint)
	}
	return nil
}
