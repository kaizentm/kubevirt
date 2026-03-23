#!/bin/bash

set -e

while getopts q:l: flag; do
    case "${flag}" in
    q) QEMU_IMAGE=${OPTARG} ;;
    l) LIBVIRT_IMAGE=${OPTARG} ;;
    *)
        echo "Invalid option"
        exit 1
        ;;
    esac
done

if [ -z "$QEMU_IMAGE" ] || [ -z "$LIBVIRT_IMAGE" ]; then
    echo "Usage: $0 -q <QEMU_IMAGE> -l <LIBVIRT_IMAGE>"
    exit 1
fi

# Start docker ctr serving QEMU RPMs
docker run --rm -dit \
    --name qemu-rpms-http-server \
    -p 9090:80 \
    $QEMU_IMAGE
sleep 5

docker images --digests

QEMU_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' qemu-rpms-http-server)
echo "QEMU repo IP:    $QEMU_IP"

curl -f "http://localhost:9090/x86_64/repodata/repomd.xml" || {
    echo 'QEMU repo not accessible'
    exit 1
}

# Extract versions (tolerate missing build-info fields)
QEMU_VERSION=$(curl -s "http://localhost:9090/build-info.json" | jq -r '.qemu_version // empty') || true
echo "Detected qemu version:    ${QEMU_VERSION:-<none>}"

# Build combined repo descriptor so rpm-deps sees both
cat >custom-repo.yaml <<EOF
repositories:
- arch: x86_64
  baseurl: http://$QEMU_IP:80/x86_64/
  name: custom-qemu
  gpgcheck: 0
  repo_gpgcheck: 0
EOF

echo "Combined custom-repo.yaml:"
cat custom-repo.yaml

make CUSTOM_REPO=custom-repo.yaml LIBVIRT_VERSION="0:11.10.0-12.el9" QEMU_VERSION="$QEMU_VERSION" SINGLE_ARCH="x86_64" rpm-deps

echo "rpm-deps completed with custom libvirt & qemu"

make bazel-build-images
