# gocatcli-ex

Fork additions of [hadi-susanto/gocatcli-ex](https://github.com/hadi-susanto/gocatcli-ex)
(`main-ex` branch) on top of [deadc0de6/gocatcli](https://github.com/deadc0de6/gocatcli).
Everything in the upstream [README](README.md) still applies.

* [Installation](#installation)
* [Compressed catalog](#compressed-catalog)
* [Rename a storage](#rename-a-storage)
* [Mountpoint handling](#mountpoint-handling)
* [Versioning and releases](#versioning-and-releases)

# Installation

Pick a binary in the [releases](https://github.com/hadi-susanto/gocatcli-ex/releases)

Or compile it yourself
```bash
$ make -f Makefile.ex
$ ./bin/gocatcli-ex --help
```

`gocatcli-ex` uses the same catalog format, default catalog name and
`GOCATCLI_` environment variables as `gocatcli`, so both can be installed side by side.

# Compressed catalog

Catalogs of large hierarchies can get big. If the catalog path ends with `.gz`,
the catalog is saved gzip-compressed (`json` or `toml` is picked from the extension
before `.gz`, for example `my.catalog.gz` or `my.toml.gz`).
Compressed catalogs are detected from their content when loading, so they are
read transparently whatever their name.

```bash
$ gocatcli-ex index -c gocatcli.catalog.gz <some-path>
$ gocatcli-ex -c gocatcli.catalog.gz ls
## or
$ export GOCATCLI_CATALOG=gocatcli.catalog.gz
```

An existing catalog can be compressed with `gzip gocatcli.catalog` and
decompressed with `gunzip gocatcli.catalog.gz`.
Note that compressed catalogs cannot be diffed when versioned with git,
and cannot be read by upstream `gocatcli`.

# Rename a storage

```bash
$ gocatcli-ex storage rename <storage-name> <new-name>
```

Entries stay attached to the storage, only its name changes.
Renaming to an existing storage name is refused.

# Mountpoint handling

On top of upstream's clean unmount on Ctrl-C, `gocatcli-ex mount`:

* creates the mountpoint directory if it does not exist
* detects a stale mount left by a crash or a forced exit
  ("Transport endpoint is not connected") and unmounts it before mounting

# Versioning and releases

`gocatcli-ex --version` prints `x.y.z-<commit>`, where `x.y.z` is the upstream version.

Releases are tagged `v<x.y.z>-ex-YY-MM-DD` (`.2`, `.3`... for extra releases the same day):
```bash
$ make -f Makefile.ex release   # tag + push, GitHub Actions publishes the release
```
