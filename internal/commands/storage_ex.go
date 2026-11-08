/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Adds "storage rename" without editing upstream's storage.go.
*/

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	storageRenameCmd = &cobra.Command{
		Use:    "rename <storage-name> <new-name>",
		Short:  "Rename a storage",
		Args:   cobra.ExactArgs(2),
		PreRun: preRun(true),
		RunE:   storageRename,
	}
)

func init() {
	storageCmd.AddCommand(storageRenameCmd)
	storageRenameCmd.PersistentFlags().BoolVarP(&storageOptIndent, "indent", "I", true, "do not indent json")
}

// storageRename renames a storage; children reference
// their storage by ID so only the name changes
func storageRename(_ *cobra.Command, args []string) error {
	var name = args[0]
	var newName = args[1]

	if len(newName) < 1 {
		return fmt.Errorf("new storage name cannot be empty")
	}

	var storage = rootTree.GetStorageByName(name)
	if storage == nil {
		return fmt.Errorf("no such storage %s", name)
	}
	if name == newName {
		return nil
	}
	if rootTree.GetStorageByName(newName) != nil {
		return fmt.Errorf("storage %s already exists", newName)
	}

	storage.Name = newName
	var err = storageSave()
	listStorages()
	return err
}
