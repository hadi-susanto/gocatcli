#!/usr/bin/env bash
# author: Hadi Susanto Lesmana (https://github.com/hadi-susanto)
# Copyright (c) 2026, Hadi Susanto Lesmana
#
# test nav start directory (storage path differs from storage name)
#

## start-test-cookie
set -eu -o errtrace -o pipefail
cur=$(cd "$(dirname "${0}")" && pwd)
bin="${cur}/../bin/gocatcli"
[ ! -e "${bin}" ] && echo "\"${bin}\" not found" && exit 1
# shellcheck disable=SC1091
source "${cur}"/helpers
## end-test-cookie

# nav is interactive and needs a tty, use util-linux script(1)
if ! script --version 2>/dev/null | grep -q util-linux; then
  echo "util-linux script not found, skipping $(basename "${0}")"
  exit 0
fi

######################################
## the test

tmpd=$(mktemp -d --suffix='-gocatcli-tests' || mktemp -d)
tmpx=$(mktemp -d --suffix='-gocatcli-tests' || mktemp -d)
clear_on_exit "${tmpd}"
clear_on_exit "${tmpx}"

catalog="${tmpd}/catalog"
out="${tmpd}/output.txt"

# create hierarchy
top="${tmpx}/top"
mkdir -p "${top}/sub"
echo "top-file" > "${top}/top-file"
echo "sub-file" > "${top}/sub/sub-file"

# index with a storage name different from the indexed path
# ===================================================
title ">>> index dir <<<"
"${bin}" index -a -C --debug -c "${catalog}" "${top}" mystorage
[ ! -e "${catalog}" ] && echo "catalog not created" && exit 1

# run nav in a pseudo-terminal, quit after a second and dump the screen
run_nav() {
  local cmd
  cmd="$(printf '%q ' "${bin}" --nocolor -c "${catalog}" nav "$@")"
  (sleep 1; printf 'q') \
    | TERM=xterm timeout 10 script -qec "stty cols 200 rows 20; ${cmd}" /dev/null \
    | sed -e 's/\x1b\[[0-9;?]*[a-zA-Z]//g' -e 's/\x1b]8;;\x1b\\//g' -e 's/\x1b(B//g' \
    | tr -s ' ' > "${out}"
  cat_file "${out}"
}

# nav without path on a single storage starts inside the storage
# ===================================================
title ">>> nav single storage without path <<<"
run_nav
grep -q 'top-file' "${out}" || (echo "storage content not shown" && exit 1)
grep -q 'Current Location: mystorage' "${out}" || (echo "wrong title" && exit 1)
grep -q 'mystorage/top-file' "${out}" && echo "entry shown with full path" && exit 1

# nav with a directory path starts inside that directory
# ===================================================
title ">>> nav with directory path <<<"
run_nav "mystorage/sub"
grep -q 'sub-file' "${out}" || (echo "directory content not shown" && exit 1)
grep -q 'top-file' "${out}" && echo "storage content shown instead of directory" && exit 1
grep -q 'Current Location: mystorage/sub' "${out}" || (echo "wrong title" && exit 1)
grep -q 'mystorage/sub/sub-file' "${out}" && echo "entry shown with full path" && exit 1

echo "test $(basename "${0}") OK!"
exit 0
