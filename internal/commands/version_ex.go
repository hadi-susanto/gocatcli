/*
Fork-only file (main-ex branch): marks builds of hadi-susanto/gocatcli.
Kept in its own file so upstream version bumps in root.go never conflict.
*/

package commands

import "strings"

const versionSuffix = "-ex"

func init() {
	if !strings.HasSuffix(version, versionSuffix) {
		version += versionSuffix
	}
	// rootCmd.Version was captured from version at package var init, before init() ran
	rootCmd.Version = version
}
