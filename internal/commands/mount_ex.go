/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Hooks into upstream's mount command without editing mount.go.
*/

package commands

import (
	"github.com/deadc0de6/gocatcli/internal/fuser"

	"github.com/spf13/cobra"
)

// wrap upstream's mount command: prepare the mountpoint before mounting
// and remove the directories created for it once unmounted
func init() {
	var upstreamRun = mountCmd.RunE
	mountCmd.RunE = func(cmd *cobra.Command, args []string) error {
		var cleanup, err = fuser.PrepareMountpoint(args[0])
		if err != nil {
			return err
		}
		defer cleanup()
		return upstreamRun(cmd, args)
	}
}
