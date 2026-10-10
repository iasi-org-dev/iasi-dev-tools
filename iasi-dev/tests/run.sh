#!/usr/bin/env bash
set -euo pipefail

# Ejecutar desde cualquier directorio, sin mover tests al arbol de trabajo.
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
cp -R "$root/src" "$tmp/src"
while IFS= read -r -d '' file; do
  relative="${file#"$root/tests/"}"
  mkdir -p -- "$tmp/src/$(dirname -- "$relative")"
  cp -- "$file" "$tmp/src/$relative"
done < <(find "$root/tests" -type f -name '*_test.go' -print0)
cd "$tmp/src"
go test ./... "$@"
