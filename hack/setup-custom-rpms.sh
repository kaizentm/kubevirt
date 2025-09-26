#!/usr/bin/env bash
#
# setup-custom-rpms.sh
#
# Purpose:
#   Reproduce the GitHub Actions step "Setup Custom libvirt & qemu RPMs" locally.
#   Spins up two HTTP repo container images (libvirt + qemu), validates accessibility,
#   generates a combined repo descriptor (YAML) and invokes `make rpm-deps` so that
#   subsequent image builds pull from the custom RPM repos.
#
# Requirements:
#   - bash, docker, curl, jq, make
#   - KubeVirt repo root as CWD (or set REPO_ROOT)
#
# Environment Variables (override as needed):
#   CUSTOM_LIBVIRT_IMAGE  (default: ghcr.io/${GITHUB_REPOSITORY:-your/repo}/libvirt-rpms:qemu-mshv)
#   CUSTOM_QEMU_IMAGE     (default: ghcr.io/${GITHUB_REPOSITORY:-your/repo}/qemu-rpms:qemu-mshv)
#   LIBVIRT_CONTAINER     (default: libvirt-rpms-http-server)
#   QEMU_CONTAINER        (default: qemu-rpms-http-server)
#   LIBVIRT_PORT          (default: 8080)  # host port mapped to container 80
#   QEMU_PORT             (default: 9090)  # host port mapped to container 80
#   OUTPUT_REPO_YAML      (default: custom-repo.yaml)
#   SINGLE_ARCH           (default: x86_64)
#   USE_CONTAINER_IP      (default: 1) If 1, baseurl uses container IP:80; if 0, uses localhost:$PORT
#   EXTRA_RPM_DEPS_ARGS   (optional) additional args passed to `make rpm-deps`
#   QUIET                 (default: 0) If 1, suppress non-error logs
#
# Generated Make invocation:
#   make CUSTOM_REPO=$OUTPUT_REPO_YAML LIBVIRT_VERSION=... QEMU_VERSION=... SINGLE_ARCH=$SINGLE_ARCH rpm-deps
#
# Example:
#   CUSTOM_LIBVIRT_IMAGE=myfork/libvirt-rpms:latest \
#   CUSTOM_QEMU_IMAGE=myfork/qemu-rpms:latest \
#   ./hack/setup-custom-rpms.sh
#
set -euo pipefail

log() {
  if [[ "${QUIET:-0}" != "1" ]]; then
    echo "[setup-custom-rpms] $*" >&2
  fi
}

err() { echo "[setup-custom-rpms][ERROR] $*" >&2; exit 1; }

usage() {
  grep '^#' "$0" | sed 's/^# \{0,1\}//'
  exit 0
}

if [[ ${1:-} == "-h" || ${1:-} == "--help" ]]; then usage; fi

: "${LIBVIRT_PORT:=8080}"
: "${QEMU_PORT:=9090}"
: "${OUTPUT_REPO_YAML:=custom-repo.yaml}"
: "${SINGLE_ARCH:=x86_64}"
: "${USE_CONTAINER_IP:=1}"
: "${LIBVIRT_CONTAINER:=libvirt-rpms-http-server}"
: "${QEMU_CONTAINER:=qemu-rpms-http-server}"

# Provide sensible defaults for images if not specified.
if [[ -z "${CUSTOM_LIBVIRT_IMAGE:-}" ]]; then
  CUSTOM_LIBVIRT_IMAGE="ghcr.io/${GITHUB_REPOSITORY:-kubevirt/unknown}/libvirt-rpms:qemu-mshv"
fi
if [[ -z "${CUSTOM_QEMU_IMAGE:-}" ]]; then
  CUSTOM_QEMU_IMAGE="ghcr.io/${GITHUB_REPOSITORY:-kubevirt/unknown}/qemu-rpms:qemu-mshv"
fi

command -v docker >/dev/null || err "docker not found in PATH"
command -v curl >/dev/null || err "curl not found in PATH"
command -v jq >/dev/null || log "jq not found; version extraction may be skipped"
command -v make >/dev/null || err "make not found in PATH"

REPO_ROOT=${REPO_ROOT:-$(pwd)}
[[ -f "$REPO_ROOT/Makefile" ]] || err "Run from KubeVirt repo root or set REPO_ROOT"

log "Using images:"
log "  libvirt: $CUSTOM_LIBVIRT_IMAGE"
log "  qemu:    $CUSTOM_QEMU_IMAGE"

log "Pulling images (may use cache)..."
docker pull "$CUSTOM_LIBVIRT_IMAGE" >/dev/null
docker pull "$CUSTOM_QEMU_IMAGE" >/dev/null

stop_container() {
  local name=$1
  if docker ps -a -q -f name="^${name}$" >/dev/null 2>&1; then
    docker rm -f "$name" >/dev/null 2>&1 || true
  fi
}

log "Ensuring old containers removed"
stop_container "$LIBVIRT_CONTAINER"
stop_container "$QEMU_CONTAINER"

log "Starting libvirt RPM HTTP server on host port $LIBVIRT_PORT"
docker run --rm -d \
  --name "$LIBVIRT_CONTAINER" \
  -p "$LIBVIRT_PORT:80" \
  "$CUSTOM_LIBVIRT_IMAGE" >/dev/null

log "Starting qemu RPM HTTP server on host port $QEMU_PORT"
docker run --rm -d \
  --name "$QEMU_CONTAINER" \
  -p "$QEMU_PORT:80" \
  "$CUSTOM_QEMU_IMAGE" >/dev/null

cleanup() {
  if [[ "${KEEP_CONTAINERS:-0}" != "1" ]]; then
    log "Stopping containers"
    docker rm -f "$LIBVIRT_CONTAINER" "$QEMU_CONTAINER" >/dev/null 2>&1 || true
  else
    log "KEEP_CONTAINERS=1 -> leaving repo servers running"
  fi
}
trap cleanup EXIT

wait_http() {
  local url=$1 name=$2 tries=30 delay=1
  for ((i=1;i<=tries;i++)); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      log "$name reachable at $url"
      return 0
    fi
    sleep $delay
  done
  err "Timed out waiting for $name at $url"
}

wait_http "http://localhost:$LIBVIRT_PORT/x86_64/repodata/repomd.xml" "libvirt repo"
wait_http "http://localhost:$QEMU_PORT/x86_64/repodata/repomd.xml" "qemu repo"

if [[ "$USE_CONTAINER_IP" == "1" ]]; then
  LIBVIRT_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$LIBVIRT_CONTAINER")
  QEMU_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$QEMU_CONTAINER")
  [[ -n "$LIBVIRT_IP" && -n "$QEMU_IP" ]] || err "Failed to obtain container IPs"
  BASEURL_LIBVIRT="http://$LIBVIRT_IP:80/x86_64/"
  BASEURL_QEMU="http://$QEMU_IP:80/x86_64/"
else
  BASEURL_LIBVIRT="http://localhost:$LIBVIRT_PORT/x86_64/"
  BASEURL_QEMU="http://localhost:$QEMU_PORT/x86_64/"
fi

log "Libvirt baseurl: $BASEURL_LIBVIRT"
log "QEMU baseurl:    $BASEURL_QEMU"

libvirt_build_info_url="http://localhost:$LIBVIRT_PORT/build-info.json"
qemu_build_info_url="http://localhost:$QEMU_PORT/build-info.json"

LIBVIRT_VERSION=${LIBVIRT_VERSION:-$(curl -fsS "$libvirt_build_info_url" | jq -r '.libvirt_version // empty' 2>/dev/null || true)}
QEMU_VERSION=${QEMU_VERSION:-$(curl -fsS "$qemu_build_info_url" | jq -r '.qemu_version // empty' 2>/dev/null || true)}

log "Detected libvirt version: ${LIBVIRT_VERSION:-<none>}"
log "Detected qemu version:    ${QEMU_VERSION:-<none>}"

cat >"$OUTPUT_REPO_YAML" <<EOF
repositories:
- arch: $SINGLE_ARCH
  baseurl: $BASEURL_LIBVIRT
  name: custom-libvirt
  gpgcheck: 0
  repo_gpgcheck: 0
- arch: $SINGLE_ARCH
  baseurl: $BASEURL_QEMU
  name: custom-qemu
  gpgcheck: 0
  repo_gpgcheck: 0
EOF

log "Generated repo descriptor: $OUTPUT_REPO_YAML"
[[ "${QUIET:-0}" == "1" ]] || cat "$OUTPUT_REPO_YAML"

MAKE_ARGS=(CUSTOM_REPO="$OUTPUT_REPO_YAML" SINGLE_ARCH="$SINGLE_ARCH")
[[ -n "${LIBVIRT_VERSION}" ]] && MAKE_ARGS+=(LIBVIRT_VERSION="$LIBVIRT_VERSION")
[[ -n "${QEMU_VERSION}" ]] && MAKE_ARGS+=(QEMU_VERSION="$QEMU_VERSION")

if [[ -n "${EXTRA_RPM_DEPS_ARGS:-}" ]]; then
  MAKE_ARGS+=($EXTRA_RPM_DEPS_ARGS)
fi

log "Invoking: make ${MAKE_ARGS[*]} rpm-deps"
make "${MAKE_ARGS[@]}" rpm-deps

log "rpm-deps completed successfully"
log "You can now run: make bazel-build-images"
