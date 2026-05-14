#!/bin/sh
set -eu

REPO="${REPO:-shadracnicholas/better-hn}"
BINARY="hn"

log() { printf '%s\n' "$*" >&2; }
err() { printf 'error: %s\n' "$*" >&2; exit 1; }

need() {
	command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

need curl
need tar
need uname
need mktemp

detect_os() {
	os=$(uname -s | tr '[:upper:]' '[:lower:]')
	case "$os" in
		linux) echo linux ;;
		darwin) echo darwin ;;
		*) err "unsupported OS: $os" ;;
	esac
}

detect_arch() {
	arch=$(uname -m)
	case "$arch" in
		x86_64|amd64) echo amd64 ;;
		aarch64|arm64) echo arm64 ;;
		*) err "unsupported architecture: $arch" ;;
	esac
}

resolve_version() {
	if [ -n "${VERSION:-}" ]; then
		echo "$VERSION"
		return
	fi
	tag=$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
		"https://github.com/${REPO}/releases/latest" | sed 's|.*/tag/||')
	[ -n "$tag" ] || err "could not determine latest version"
	echo "$tag"
}

pick_install_dir() {
	if [ -n "${INSTALL_DIR:-}" ]; then
		echo "$INSTALL_DIR"
		return
	fi
	if [ -w /usr/local/bin ] 2>/dev/null; then
		echo /usr/local/bin
	else
		echo "$HOME/.local/bin"
	fi
}

main() {
	os=$(detect_os)
	arch=$(detect_arch)
	version=$(resolve_version)
	install_dir=$(pick_install_dir)

	archive="${BINARY}_${version}_${os}_${arch}.tar.gz"
	url="https://github.com/${REPO}/releases/download/${version}/${archive}"
	sums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

	log "installing ${BINARY} ${version} (${os}/${arch}) -> ${install_dir}"

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT

	log "downloading ${url}"
	curl -fsSL "$url" -o "${tmp}/${archive}" \
		|| err "failed to download ${url}"

	if curl -fsSL "$sums_url" -o "${tmp}/checksums.txt" 2>/dev/null; then
		log "verifying checksum"
		expected=$(grep " ${archive}\$" "${tmp}/checksums.txt" | awk '{print $1}')
		if [ -n "$expected" ]; then
			if command -v sha256sum >/dev/null 2>&1; then
				actual=$(sha256sum "${tmp}/${archive}" | awk '{print $1}')
			else
				actual=$(shasum -a 256 "${tmp}/${archive}" | awk '{print $1}')
			fi
			[ "$expected" = "$actual" ] \
				|| err "checksum mismatch: expected $expected, got $actual"
		else
			log "warning: no checksum entry for ${archive}, skipping verification"
		fi
	else
		log "warning: checksums.txt not available, skipping verification"
	fi

	tar -xzf "${tmp}/${archive}" -C "$tmp"
	src="${tmp}/${BINARY}_${version}_${os}_${arch}/${BINARY}"
	[ -f "$src" ] || err "binary not found in archive: $src"

	mkdir -p "$install_dir"
	dest="${install_dir}/${BINARY}"
	if ! mv "$src" "$dest" 2>/dev/null; then
		log "elevating with sudo to write ${dest}"
		sudo mv "$src" "$dest"
		sudo chmod +x "$dest"
	else
		chmod +x "$dest"
	fi

	log "installed ${dest}"

	case ":${PATH}:" in
		*:"${install_dir}":*) ;;
		*)
			log ""
			log "note: ${install_dir} is not on your PATH"
			log "add this to your shell profile:"
			log "    export PATH=\"${install_dir}:\$PATH\""
			;;
	esac
}

main "$@"
