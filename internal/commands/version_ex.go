/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Kept in its own file so upstream changes in root.go never conflict.
*/

package commands

// ApplyEx turns the shared commands into the gocatcli-ex flavour:
// the upstream x.y.z version gets the commit appended ("1.3.2-b0482f5")
// and the command is named after the fork binary.
// It must be called before Execute.
func ApplyEx(binName string, commit string) {
	if commit != "" {
		version += "-" + commit
	}
	// rootCmd captured version at package init, refresh it
	rootCmd.Version = version
	rootCmd.Use = binName
}
