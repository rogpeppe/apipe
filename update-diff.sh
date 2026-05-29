#!/bin/bash

# update-diff.sh — vendor golang.org/x/tools/internal/diff into internal/diff.
#
# Usage:
#   ./update-diff.sh              # fetches latest version
#   ./update-diff.sh v0.45.0      # fetches a specific version

# NOTE: We could just use github.com/akedrou/textdiff which
# provides the same API as the internal diff package,
# but we don't want to depend on some random potentially
# malicious upstream so the vendor the internal package instead.

set -eu

MODULE="golang.org/x/tools"
VERSION="${1:-latest}"

# Resolve the version and download.
resolved=$(go list -m -f '{{.Version}}' "${MODULE}@${VERSION}")
echo "Using ${MODULE}@${resolved}"
modcache=$(go env GOMODCACHE)
moddir="${modcache}/${MODULE}@${resolved}"
srcdir="${moddir}/internal/diff"
if [ ! -d "$srcdir" ]; then
	echo "error: ${srcdir} not found; run 'go mod download ${MODULE}@${resolved}'" >&2
	exit 1
fi

# Our module path, read from go.mod.
ourmod=$(go list -m)
destdir="internal/diff"

rm -rf "${destdir}"
mkdir -p "${destdir}/lcs"

# Copy non-test Go source files.
for dir in "" "/lcs"; do
	for f in "${srcdir}${dir}"/*.go; do
		base=$(basename "$f")
		case "$base" in
		*_test.go) continue ;;
		export_test.go) continue ;;
		esac
		cp "$f" "${destdir}${dir}/${base}"
	done
done

# Rewrite the import path for the lcs subpackage.
find "${destdir}" -name '*.go' -exec \
	sed -i "s|\"${MODULE}/internal/diff/lcs\"|\"${ourmod}/internal/diff/lcs\"|g" {} +

# Record provenance.
cat > "${destdir}/VERSION" <<EOF
Vendored from ${MODULE}@${resolved}
EOF

echo "Vendored ${MODULE}/internal/diff@${resolved} into ${destdir}/"
go build ./internal/diff/...
echo "Build OK."
