#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

while getopts r:v: flag; do
    case "${flag}" in
    r) QEMU_REPO=${OPTARG} ;;
    v) QEMU_VERSION=${OPTARG} ;;
    *)
        echo "Invalid option"
        exit 1
        ;;
    esac
done

if [ -z "$QEMU_REPO" ] || [ -z "$QEMU_VERSION" ]; then
    echo "Usage: $0 -r <QEMU_REPO> -v <QEMU_VERSION>"
    exit 1
fi

# Fetch QEMU source code from upstream
rm -rf ./qemu-rpm-build
mkdir -p ./qemu-rpm-build
cd ./qemu-rpm-build

# RPM spec compatible version of QEMU version
# 1. Replace hyphens with dot
QEMU_SPEC_VERSION=${QEMU_VERSION//-/.}
git clone -b v${QEMU_VERSION} ${QEMU_REPO} qemu-${QEMU_SPEC_VERSION}

# Create tarball of QEMU source code
tar -cf qemu-${QEMU_SPEC_VERSION}.tar.xz \
    qemu-${QEMU_SPEC_VERSION}
rm -rf qemu-${QEMU_SPEC_VERSION}

# Copy spec file and related files
cp $SCRIPT_DIR/qemu-spec/* .

sed -i "s/Version:.*$/Version: ${QEMU_SPEC_VERSION}/" qemu.spec

docker rm -f qemu-build

docker run -td \
    --name qemu-build \
    -v $(pwd):/qemu-src \
    registry.gitlab.com/libvirt/libvirt/ci-centos-stream-9

# Build qemu RPM
docker exec -w /qemu-src qemu-build bash -c "
  set -ex
  mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
  cp qemu.spec ~/rpmbuild/SPECS
  cp *.patch ~/rpmbuild/SOURCES/
  cp qemu-${QEMU_SPEC_VERSION}.tar.xz ~/rpmbuild/SOURCES/
  cd ~/rpmbuild/SPECS
  dnf update -y
  dnf -y install createrepo
  dnf builddep -y qemu.spec
  rpmbuild -ba qemu.spec
  cd ~/rpmbuild/RPMS
  createrepo --general-compress-type=gz --checksum=sha256 x86_64
"

cd ../

docker cp qemu-build:/root/rpmbuild/RPMS ./rpms-qemu

cat >./rpms-qemu/build-info.json <<EOF
{
  "qemu_version": "0:${QEMU_SPEC_VERSION}-1.el9"
}
EOF
