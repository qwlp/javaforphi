#!/usr/bin/env bash
set -euo pipefail
source_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
command -v git >/dev/null || { echo 'Updating requires Git and a cloned checkout.' >&2; exit 1; }
if [[ "$(git -C "$source_dir" rev-parse --show-toplevel 2>/dev/null || true)" != "$source_dir" ]]; then
  echo 'Updating requires a cloned checkout. For a ZIP download, download the new source and run install.sh.' >&2
  exit 1
fi
if [[ -n "$(git -C "$source_dir" status --porcelain)" ]]; then
  echo 'Commit or stash your checkout changes before updating.' >&2
  exit 1
fi
git -C "$source_dir" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' >/dev/null || {
  echo 'Your current branch needs an upstream before updating.' >&2; exit 1;
}
echo 'Updating the checkout from its upstream...'
git -C "$source_dir" pull --ff-only
bash "$source_dir/install.sh"
echo 'phi is updated.'
