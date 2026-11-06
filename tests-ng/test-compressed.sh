#!/usr/bin/env bash
# author: deadc0de6 (https://github.com/deadc0de6)
# Copyright (c) 2024, deadc0de6
#
# test compressed (gzip) catalogs
#

## start-test-cookie
set -eu -o errtrace -o pipefail
cur=$(cd "$(dirname "${0}")" && pwd)
bin="${cur}/../bin/gocatcli"
[ ! -e "${bin}" ] && echo "\"${bin}\" not found" && exit 1
# shellcheck disable=SC1091
source "${cur}"/helpers
## end-test-cookie

######################################
## the test

tmpd=$(mktemp -d --suffix='-gocatcli-tests' || mktemp -d)
clear_on_exit "${tmpd}"

src="${cur}/../internal"
out="${tmpd}/output.txt"

# $1: file path
is_gzip()
{
  [ "$(head -c2 "${1}" | od -An -tx1 | tr -d ' \n')" = "1f8b" ]
}

# $1: catalog path
count_entries()
{
  "${bin}" -c "${1}" ls -a -r | sed -e 's/\x1b\[[0-9;]*m//g' | sed '/^$/d' | wc -l | awk '{print $1}'
}

# reference plain catalog
# ===================================================
title ">>> index plain reference <<<"
plain="${tmpd}/plain.catalog"
"${bin}" index -a -C -c "${plain}" "${src}" internal
is_gzip "${plain}" && echo "plain catalog should not be compressed" && exit 1
expected=$(count_entries "${plain}")

for catalog in "${tmpd}/catalog.gz" "${tmpd}/catalog.toml.gz"; do
  # index
  # ===================================================
  title ">>> index compressed $(basename "${catalog}") <<<"
  "${bin}" index -a -C -c "${catalog}" "${src}" internal
  [ ! -e "${catalog}" ] && echo "catalog not created" && exit 1
  ! is_gzip "${catalog}" && echo "catalog not compressed" && exit 1

  # ls
  # ===================================================
  title ">>> ls compressed $(basename "${catalog}") <<<"
  cnt=$(count_entries "${catalog}")
  [ "${cnt}" != "${expected}" ] && echo "expecting ${expected} entries got ${cnt}" && exit 1

  # update and reload
  # ===================================================
  title ">>> tag compressed $(basename "${catalog}") <<<"
  "${bin}" -c "${catalog}" storage tag internal ztag > /dev/null
  ! is_gzip "${catalog}" && echo "catalog not compressed after update" && exit 1
  "${bin}" -c "${catalog}" storage list | sed -e 's/\x1b\[[0-9;]*m//g' > "${out}"
  grep -q "tags:ztag" "${out}" || (echo "tag not saved" && cat "${out}" && exit 1)

  # decompress and load as plain
  # ===================================================
  title ">>> gunzip $(basename "${catalog}") <<<"
  extracted="${catalog%.gz}"
  [ "${extracted}" = "${tmpd}/catalog" ] && extracted="${tmpd}/catalog.json"
  gunzip -c "${catalog}" > "${extracted}"
  cnt=$(count_entries "${extracted}")
  [ "${cnt}" != "${expected}" ] && echo "extracted: expecting ${expected} entries got ${cnt}" && exit 1
done

# compressed content with a non .gz name is still loaded
# ===================================================
title ">>> load renamed compressed catalog <<<"
renamed="${tmpd}/renamed.catalog"
cp "${tmpd}/catalog.gz" "${renamed}"
cnt=$(count_entries "${renamed}")
[ "${cnt}" != "${expected}" ] && echo "renamed: expecting ${expected} entries got ${cnt}" && exit 1

echo "test \"$(basename "${0}")\" success"
exit 0
