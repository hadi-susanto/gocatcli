# gocatcli-ex

Fork additions of [hadi-susanto/gocatcli-ex](https://github.com/hadi-susanto/gocatcli-ex)
(`main-ex` branch) on top of [deadc0de6/gocatcli](https://github.com/deadc0de6/gocatcli).
Everything in the upstream [README](README.md) still applies.

* [Installation](#installation)
* [Compressed catalog](#compressed-catalog)
* [Rename a storage](#rename-a-storage)
* [Mountpoint handling](#mountpoint-handling)
* [Mount on Windows](#mount-on-windows)
* [Versioning and releases](#versioning-and-releases)
* [Credits](#credits)

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

On Linux/macOS (Windows: see [Mount on Windows](#mount-on-windows)).

Since upstream 1.3.3, `mount` unmounts cleanly on Ctrl-C, `SIGTERM` or `SIGHUP`.
If the mount is busy and cannot be unmounted, it keeps serving and a second
Ctrl-C forces the exit, which leaves a stale mount behind.

Before mounting, `gocatcli-ex mount` also:

* creates the mountpoint directory (and its parents) if it does not exist
* refuses a mountpoint that exists but is not a directory
* detects a stale mount left by a crash or a forced exit
  ("Transport endpoint is not connected"), even when the kernel still serves
  the mountpoint from its cache, and unmounts it; if that fails, it suggests
  `fusermount -uz <mountpoint>`

# Mount on Windows

Upstream `gocatcli` has no Windows mount. `gocatcli-ex` mounts catalogs on Windows
through [WinFsp](https://winfsp.dev), which must be installed first
(get the installer from the [WinFsp releases](https://winfsp.dev/rel/)).
Without it, `mount` stops with an error explaining how to install it.

```bash
> gocatcli-ex mount X:                 # on a drive letter
> gocatcli-ex mount *                  # on the first free drive letter
> gocatcli-ex mount C:\mnt\catalog     # on a directory that does not exist yet
```

* the mount is read-only; files show their size, date and mode, but their content is empty
  (a catalog only stores metadata)
* a directory mountpoint must not exist, WinFsp creates it (missing parents are created)
* press Ctrl-C to unmount
* on Linux/macOS mount still uses FUSE, as upstream

# Versioning and releases

`gocatcli-ex --version` prints `x.y.z-<commit>`, where `x.y.z` is the upstream version.

Releases are tagged `v<x.y.z>-ex-YY-MM-DD` (`.2`, `.3`... for extra releases the same day):
```bash
$ make -f Makefile.ex release   # tag + push, GitHub Actions publishes the release
```

# Credits

The Windows mount is built on:

* **WinFsp - Windows File System Proxy, Copyright (C) Bill Zissimopoulos**
  ([github.com/winfsp/winfsp](https://github.com/winfsp/winfsp)),
  GPLv3 with a FLOSS exception. WinFsp is not bundled: install it from
  [winfsp.dev](https://winfsp.dev/rel/).
* **cgofuse**, Copyright (c) 2017-2022 Bill Zissimopoulos
  ([github.com/winfsp/cgofuse](https://github.com/winfsp/cgofuse)), MIT license.
