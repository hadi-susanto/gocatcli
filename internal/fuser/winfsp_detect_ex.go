//go:build windows

/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
Detects WinFsp (https://github.com/winfsp/winfsp), the user mode file
system layer used to mount catalogs on Windows.
*/

package fuser

import (
	"errors"
	"path/filepath"
	"runtime"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// WinFspCredit is the notice required by the WinFsp FLOSS exception
const WinFspCredit = "WinFsp - Windows File System Proxy, Copyright (C) Bill Zissimopoulos (https://github.com/winfsp/winfsp)"

var errWinFspMissing = errors.New("WinFsp is required to mount a catalog on Windows, " +
	"install it from https://winfsp.dev/rel/ and try again\n" + WinFspCredit)

// winfspDLLName returns the WinFsp DLL matching the current architecture
func winfspDLLName() string {
	switch runtime.GOARCH {
	case "arm64":
		return "winfsp-a64.dll"
	case "386":
		return "winfsp-x86.dll"
	default:
		return "winfsp-x64.dll"
	}
}

// winfspInstalled mirrors cgofuse's DLL lookup: by name first,
// then through the install directory found in the registry
func winfspInstalled() bool {
	var name = winfspDLLName()
	if tryLoadDLL(name) {
		return true
	}

	var key, err = registry.OpenKey(registry.LOCAL_MACHINE, `Software\WinFsp`,
		registry.QUERY_VALUE|registry.WOW64_32KEY)
	if err != nil {
		return false
	}
	defer key.Close()

	var dir string
	dir, _, err = key.GetStringValue("InstallDir")
	if err != nil || dir == "" {
		return false
	}
	return tryLoadDLL(filepath.Join(dir, "bin", name))
}

func tryLoadDLL(path string) bool {
	var dll, err = syscall.LoadDLL(path)
	if err != nil {
		return false
	}
	_ = dll.Release()
	return true
}
