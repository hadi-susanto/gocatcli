//go:build windows

/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Windows counterpart of sudo_ex.go: there is no sudo.
*/

package fuser

func chownToInvoker(_ string) {}
