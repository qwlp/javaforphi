#!/usr/bin/env bash
set -euo pipefail

# Run from a checkout; the script locates the source independently of cwd.
source_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
install_dir="$HOME/.local/share/phi"
if [[ -L "$install_dir" || ( -e "$install_dir" && ! -f "$install_dir/.phi-install" ) ]]; then
  echo "Refusing to overwrite unmanaged installation: $install_dir" >&2
  exit 1
fi
case "$(uname -s)" in
  Linux) platform=linux ;;
  Darwin) platform=darwin ;;
  *) echo 'Use install.ps1 on Windows.' >&2; exit 1 ;;
esac
public_bin="$install_dir/bin"
if [[ "$platform" == linux ]]; then
  public_bin="$HOME/.local/bin"
  public_phi="$public_bin/phi"
  if [[ -e "$public_phi" || -L "$public_phi" ]]; then
    if [[ ! -L "$public_phi" || "$(readlink "$public_phi")" != "$install_dir/bin/phi" ]]; then
      echo "Refusing to overwrite an existing unmanaged command: $public_phi" >&2
      exit 1
    fi
  fi
fi
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo 'Supported architectures: x86-64 and ARM64.' >&2; exit 1 ;;
esac
stage=$(mktemp -d)
trap 'rm -rf -- "$stage"' EXIT
minimum_go=$(awk '$1 == "go" { print $2; exit }' "$source_dir/go.mod")
[[ "$minimum_go" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]] || { echo 'Cannot read the required Go version from go.mod.' >&2; exit 1; }
go_command=''
for candidate in "$(command -v go || true)" "$install_dir/go/bin/go"; do
  [[ -n "$candidate" && -x "$candidate" ]] || continue
  installed_version=$(GOTOOLCHAIN=local "$candidate" version 2>/dev/null | awk '{print $3}' || true)
  if [[ "$installed_version" =~ ^go[0-9]+\.[0-9]+(\.[0-9]+)?$ ]] && awk -v current="${installed_version#go}" -v minimum="$minimum_go" 'BEGIN {
    split(current, c, "."); split(minimum, m, ".")
    for (i = 1; i <= 3; i++) {
      if (c[i]+0 > m[i]+0) exit 0
      if (c[i]+0 < m[i]+0) exit 1
    }
    exit 0
  }'; then
    go_command="$candidate"
    echo "Using existing $installed_version: $go_command"
    break
  fi
done
if [[ -z "$go_command" ]]; then
echo "Go $minimum_go or newer was not found; installing a private SDK."
download() {
  curl --fail --location --silent --show-error --connect-timeout 15 --max-time 600 --retry 2 "$1" -o "$2"
}
# The default API returns stable releases, newest first. These fields contain
# plain filenames and hex digests; split objects without requiring jq or Python.
download 'https://go.dev/dl/?mode=json' "$stage/releases.json"
archive_info=$(tr '\n' ' ' < "$stage/releases.json" | tr '{}' '\n\n' | awk -F '"' -v suffix=".$platform-$arch.tar.gz" '
  {
    filename = ""; digest = ""
    for (i = 2; i <= NF; i += 2) {
      if ($i == "filename") filename = $(i + 2)
      if ($i == "sha256") digest = $(i + 2)
    }
    if (filename != "" && substr(filename, length(filename) - length(suffix) + 1) == suffix) {
      print filename, digest; exit
    }
  }')
read -r filename digest <<< "$archive_info"
if [[ ! "$filename" =~ ^go[0-9]+\.[0-9]+(\.[0-9]+)?\.$platform-$arch\.tar\.gz$ || ! "$digest" =~ ^[0-9a-f]{64}$ ]]; then
  echo "No valid stable Go archive found for $platform/$arch." >&2
  exit 1
fi
go_version=${filename%%.$platform-*}
go_version=${go_version#go}
echo "Downloading Go $go_version for $platform/$arch..."
download "https://go.dev/dl/$filename" "$stage/go.tar.gz"
if command -v sha256sum >/dev/null; then
  actual=$(sha256sum "$stage/go.tar.gz" | awk '{print $1}')
else
  actual=$(shasum -a 256 "$stage/go.tar.gz" | awk '{print $1}')
fi
[[ "$actual" == "$digest" ]] || { echo 'Go checksum verification failed.' >&2; exit 1; }
tar -xzf "$stage/go.tar.gz" -C "$stage"
go_command="$stage/go/bin/go"
fi
echo 'Building phi...'
(cd "$source_dir" &&
  if [[ -d "$stage/go" ]]; then export GOROOT="$stage/go"; fi
  GOTOOLCHAIN=local CGO_ENABLED=0 "$go_command" build -trimpath -o "$stage/phi" .)
mkdir -p "$install_dir/bin"
touch "$install_dir/.phi-install"
printf '%s\n' "$source_dir" > "$install_dir/source-dir"
# Replace the private SDK only when a new one was required.
if [[ -d "$stage/go" ]]; then
  rm -rf -- "$install_dir/go"
  mv "$stage/go" "$install_dir/go"
fi
mv "$stage/phi" "$install_dir/bin/phi"
if [[ "$platform" == linux ]]; then
  mkdir -p "$public_bin"
  ln -sfn "$install_dir/bin/phi" "$public_bin/phi"
fi
path_line='export PATH="$HOME/.local/share/phi/bin:$HOME/.local/share/phi/go/bin:$PATH" # phi installer'
old_path_line="$path_line"
if [[ "$platform" == linux ]]; then
  path_line='export PATH="$HOME/.local/bin:$HOME/.local/share/phi/go/bin:$PATH" # phi installer'
fi
for profile in .profile .bashrc .bash_profile .zshrc .zprofile; do
  if [[ "$path_line" != "$old_path_line" ]] && grep -Fqx "$old_path_line" "$HOME/$profile" 2>/dev/null; then
    awk -v line="$old_path_line" '$0 != line' "$HOME/$profile" > "$stage/profile"
    cat "$stage/profile" > "$HOME/$profile"
  fi
  if ! grep -Fqx "$path_line" "$HOME/$profile" 2>/dev/null; then
    printf '\n%s\n' "$path_line" >> "$HOME/$profile"
  fi
done
fish_dir="$HOME/.config/fish/conf.d"
fish_file="$fish_dir/phi-installer.fish"
if [[ -e "$fish_file" ]] && ! grep -Fq '# phi installer' "$fish_file"; then
  echo "Skipping existing unmanaged fish configuration: $fish_file"
else
  mkdir -p "$fish_dir"
  if [[ "$platform" == linux ]]; then
    fish_path='set -gx PATH "$HOME/.local/bin" "$HOME/.local/share/phi/go/bin" $PATH'
  else
    fish_path='set -gx PATH "$HOME/.local/share/phi/bin" "$HOME/.local/share/phi/go/bin" $PATH'
  fi
  printf '%s\n' '# phi installer' "$fish_path" > "$fish_file"
fi
"$install_dir/bin/phi" version
echo 'Installed phi. Open a new terminal, then run: phi doctor'
echo "For this terminal: ${path_line% # phi installer}"
