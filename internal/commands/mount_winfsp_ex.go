//go:build windows

/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Windows help for upstream's mount command.
*/

package commands

import (
	"github.com/deadc0de6/gocatcli/internal/fuser"
)

func init() {
	mountCmd.Short = "Mount catalog using WinFsp"
	mountCmd.Long = `Mount the catalog read-only (files are listed with their metadata, their content is empty).

<path> is either:
  - a drive letter, e.g. X:
  - * to use the first free drive letter
  - a directory that does not exist yet (WinFsp creates it)

Press Ctrl-C to unmount.

Requires WinFsp, install it from https://winfsp.dev/rel/
` + fuser.WinFspCredit
}
