#!/usr/bin/env bash
set -euo pipefail
workspaces=("$HOME/phi-lessons")
if [[ -n "${PHI_WORKSPACE:-}" ]]; then workspaces+=("$PHI_WORKSPACE"); fi
while [[ $# -gt 0 ]]; do
  case "$1" in
    --workspace)
      [[ $# -ge 2 && -n "$2" ]] || { echo 'Usage: bash uninstall.sh [--workspace <folder>]' >&2; exit 1; }
      workspaces+=("$2"); shift 2 ;;
    *) echo 'Usage: bash uninstall.sh [--workspace <folder>]' >&2; exit 1 ;;
  esac
done
install_dir="$HOME/.local/share/phi"
if [[ -L "$install_dir" || ( -e "$install_dir" && ! -f "$install_dir/.phi-install" ) ]]; then
  echo "Refusing to remove unmanaged installation: $install_dir" >&2
  exit 1
fi
lab_dirs=()
lab_count=0
home_dir=$(cd -- "$HOME" && pwd -P)
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
for workspace in "${workspaces[@]}"; do
  [[ -d "$workspace" && ! -L "$workspace" ]] || continue
  workspace=$(cd -- "$workspace" && pwd -P)
  for lab in "$workspace"/*; do
    [[ -d "$lab" && ! -L "$lab" && "$lab" != "$home_dir" && "$lab" != "$script_dir" ]] || continue
    marker=''
    for name in .phi.json .javaforphi.json; do
      if [[ -f "$lab/$name" && ! -L "$lab/$name" ]] && grep -Eq '"lesson"[[:space:]]*:[[:space:]]*"[^"[:space:]]+"' "$lab/$name"; then
        marker="$name"; break
      fi
    done
    [[ -n "$marker" ]] || continue
    duplicate=false
    if [[ "$lab_count" -gt 0 ]]; then
      for existing in "${lab_dirs[@]}"; do [[ "$existing" != "$lab" ]] || duplicate=true; done
    fi
    if ! "$duplicate"; then
      lab_dirs+=("$lab")
      lab_count=$((lab_count + 1))
    fi
  done
done
recovery_dir=''
if [[ "$lab_count" -gt 0 ]]; then
  echo 'Generated lab folders found (including your work and documents):'
  printf '  %s\n' "${lab_dirs[@]}"
  echo 'Removing them will move them to a recovery folder so you can restore them.'
  printf 'Remove these generated lab folders? [y/N] '
  answer=''
  IFS= read -r answer || true
  case "$answer" in
    y|Y|yes|YES|Yes)
      recovery_base="$HOME/.local/share/phi-lab-backups"
      [[ ! -L "$recovery_base" ]] || { echo "Refusing to use linked recovery directory: $recovery_base" >&2; exit 1; }
      mkdir -p "$recovery_base"
      recovery_dir=$(mktemp -d "$recovery_base/labs.XXXXXX")
      # Numbered subfolders avoid collisions between labs in different workspaces.
      index=0
      for lab in "${lab_dirs[@]}"; do
        [[ -d "$lab" && ! -L "$lab" ]] || { echo "Lab folder changed: $lab" >&2; exit 1; }
        index=$((index + 1))
        mkdir "$recovery_dir/$index"
        printf '%s\n' "$lab" > "$recovery_dir/$index/original-path.txt"
        mv -- "$lab" "$recovery_dir/$index/"
      done
      printf '\nLab folders removed from their workspaces. Restore them from: %s\n' "$recovery_dir"
      ;;
    *) printf '\nGenerated lab folders were kept.\n' ;;
  esac
fi
path_line='export PATH="$HOME/.local/share/phi/bin:$HOME/.local/share/phi/go/bin:$PATH" # phi installer'
linux_path_line='export PATH="$HOME/.local/bin:$HOME/.local/share/phi/go/bin:$PATH" # phi installer'
for profile in .profile .bashrc .bash_profile .zshrc .zprofile; do
  file="$HOME/$profile"
  if [[ -f "$file" ]] && { grep -Fqx "$path_line" "$file" || grep -Fqx "$linux_path_line" "$file"; }; then
    scratch=$(mktemp)
    awk -v line="$path_line" -v linux_line="$linux_path_line" '$0 != line && $0 != linux_line' "$file" > "$scratch"
    # Write through the existing file to preserve permissions and symlinks.
    cat "$scratch" > "$file"
    rm -f -- "$scratch"
  fi
done
fish_file="$HOME/.config/fish/conf.d/phi-installer.fish"
if [[ -f "$fish_file" ]] && grep -Fqx '# phi installer' "$fish_file"; then
  rm -f -- "$fish_file"
fi
if [[ -f "$install_dir/.phi-install" ]]; then
  public_phi="$HOME/.local/bin/phi"
  if [[ -L "$public_phi" && "$(readlink "$public_phi")" == "$install_dir/bin/phi" ]]; then
    rm -f -- "$public_phi"
  fi
  rm -rf -- "$install_dir"
fi
echo 'Removed installer-owned phi, Go, and PATH configuration. Open a new terminal.'
if [[ -z "$recovery_dir" ]]; then echo 'Your lessons were kept.'; fi
echo 'Your dependency cache, settings, and other Go installations were kept.'
