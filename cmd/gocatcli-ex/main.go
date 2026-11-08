/*
Fork-only entry point (main-ex branch of hadi-susanto/gocatcli-ex).
Builds the gocatcli-ex binary, see Makefile.ex.
*/

package main

import (
	"os"
	"runtime/debug"

	"github.com/deadc0de6/gocatcli/internal/commands"
	"github.com/deadc0de6/gocatcli/internal/log"
)

const binName = "gocatcli-ex"

// commit is set at build time with
// -ldflags "-X main.commit=<short-sha>"
var commit = ""

func main() {
	if commit == "" {
		commit = vcsCommit()
	}
	commands.ApplyEx(binName, commit)

	var err = commands.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

// vcsCommit is the fallback when built without ldflags: it reads the
// git revision embedded by `go build` in a git checkout (Go >= 1.18).
func vcsCommit() string {
	var info, ok = debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	var rev string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return "dev"
	}
	if len(rev) > 7 {
		rev = rev[:7]
	}
	if dirty {
		rev += "-dirty"
	}
	return rev
}
