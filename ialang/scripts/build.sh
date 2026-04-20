#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

OUTPUT="${1:-${ROOT_DIR}/bin/ialang}"
TARGET="${2:-./cmd/ialang}"

mkdir -p "$(dirname "${OUTPUT}")"

cd "${ROOT_DIR}"
go build -o "${OUTPUT}" "${TARGET}"

# upx
upx "${OUTPUT}"

echo "Build succeeded: ${OUTPUT}"
