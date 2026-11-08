#!/usr/bin/env bash
# Fork-only test (main-ex branch of hadi-susanto/gocatcli-ex)
#
# test storage rename command
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

catalog="${tmpd}/catalog"
out="${tmpd}/output.txt"

# index
# ===================================================
title ">>> index <<<"
"${bin}" index -a -C -c "${catalog}" "${cur}/../internal" internal
"${bin}" index -a -C -c "${catalog}" "${cur}/../tests-ng" testsng
[ ! -e "${catalog}" ] && echo "catalog not created" && exit 1

# ===================================================
title ">>> test rename <<<"
"${bin}" --debug -c "${catalog}" storage rename testsng renamed | sed -e 's/\x1b\[[0-9;]*m//g' > "${out}"
cat "${out}"
cnt=$(wc -l "${out}" | awk '{print $1}')
[ "${cnt}" != "2" ] && echo "expecting 2 lines (${cnt})" && exit 1
grep -q "^storage renamed" "${out}" || (echo "storage not renamed" && exit 1)
grep -q "^storage testsng" "${out}" && echo "old name still present" && exit 1
# children still reachable under the new name
cnt=$("${bin}" -c "${catalog}" ls renamed | sed -e 's/\x1b\[[0-9;]*m//g' | sed '/^$/d' | wc -l | awk '{print $1}')
[ "${cnt}" = "0" ] && echo "no children under renamed storage" && exit 1

# ===================================================
title ">>> test rename collision <<<"
if "${bin}" -c "${catalog}" storage rename renamed internal > /dev/null 2>&1; then
  echo "rename to an existing name should fail" && exit 1
fi
if "${bin}" -c "${catalog}" storage rename doesnotexist foo > /dev/null 2>&1; then
  echo "rename of a missing storage should fail" && exit 1
fi

# ===================================================
title ">>> test rename back <<<"
"${bin}" -c "${catalog}" storage rename renamed testsng > /dev/null
"${bin}" -c "${catalog}" storage list | sed -e 's/\x1b\[[0-9;]*m//g' | grep -q "^storage testsng" || (echo "rename back failed" && exit 1)

echo "test $(basename "${0}") OK!"
exit 0
