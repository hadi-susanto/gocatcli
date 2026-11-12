/*
Fork-only file (main-ex branch of hadi-susanto/gocatcli-ex).
author: Hadi Susanto Lesmana (https://github.com/hadi-susanto)
*/

package fuser

import (
	"os"
	"path/filepath"
	"testing"
)

func exists(t *testing.T, path string) bool {
	t.Helper()
	var _, err = os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return err == nil
}

func TestMkdirTrackedCreatesOnlyMissing(t *testing.T) {
	var base = t.TempDir()
	var keep = filepath.Join(base, "keep")
	var err = os.Mkdir(keep, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	var target = filepath.Join(keep, "a", "b", "c")

	var created []string
	created, err = mkdirTracked(target)
	if err != nil {
		t.Fatal(err)
	}
	var want = []string{
		filepath.Join(keep, "a"),
		filepath.Join(keep, "a", "b"),
		target,
	}
	if len(created) != len(want) {
		t.Fatalf("created %v, want %v", created, want)
	}
	for i := range want {
		if created[i] != want[i] {
			t.Fatalf("created %v, want %v", created, want)
		}
	}

	removeCreated(created)
	if exists(t, filepath.Join(keep, "a")) {
		t.Fatal("created directories not removed")
	}
	if !exists(t, keep) {
		t.Fatal("pre-existing directory removed")
	}
}

func TestMkdirTrackedExisting(t *testing.T) {
	var base = t.TempDir()
	var created, err = mkdirTracked(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 0 {
		t.Fatalf("created %v for an existing directory", created)
	}
	removeCreated(created)
	if !exists(t, base) {
		t.Fatal("existing directory removed")
	}
}

func TestRemoveCreatedKeepsNonEmpty(t *testing.T) {
	var base = t.TempDir()
	var target = filepath.Join(base, "a", "b")
	var created, err = mkdirTracked(target)
	if err != nil {
		t.Fatal(err)
	}
	var other = filepath.Join(base, "a", "other")
	if err = os.WriteFile(other, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	removeCreated(created)
	if exists(t, target) {
		t.Fatal("empty created directory not removed")
	}
	if !exists(t, other) {
		t.Fatal("non-empty created directory content removed")
	}
}

func TestMkdirTrackedFileInPath(t *testing.T) {
	var base = t.TempDir()
	var file = filepath.Join(base, "file")
	var err = os.WriteFile(file, nil, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	var created []string
	created, err = mkdirTracked(filepath.Join(file, "x", "y"))
	if err == nil {
		t.Fatalf("expected an error, created %v", created)
	}
	if len(created) != 0 {
		t.Fatalf("created %v on error", created)
	}
}
