#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

while getopts r:t:v: flag; do
    case "${flag}" in
    r) QEMU_REPO=${OPTARG} ;;
    t) QEMU_TAG=${OPTARG} ;;
    v) QEMU_VERSION=${OPTARG} ;;
    *)
        echo "Invalid option"
        exit 1
        ;;
    esac
done

if [ -z "$QEMU_REPO" ] || [ -z "$QEMU_TAG" ]; then
    echo "Usage: $0 -r <QEMU_REPO> -t <QEMU_TAG> [-v <QEMU_VERSION>]"
    echo "Note: If QEMU_VERSION is not provided, it will derivable from QEMU_TAG."
    echo "It will be assumed that QEMU_TAG is of the form v<QEMU_VERSION>, with hyphens replaced by dots."
    exit 1
fi

if [ -z "$QEMU_VERSION" ]; then
    # Derive QEMU_VERSION from QEMU_TAG
    QEMU_VERSION=${QEMU_TAG#v}
    QEMU_VERSION=${QEMU_VERSION//-/.}
fi

# Fetch QEMU source code from upstream
rm -rf ./qemu-rpm-build
mkdir -p ./qemu-rpm-build
cd ./qemu-rpm-build

git clone -b ${QEMU_TAG} ${QEMU_REPO} qemu-${QEMU_VERSION}

# Create tarball of QEMU source code
tar -cf qemu-${QEMU_VERSION}.tar.xz \
    qemu-${QEMU_VERSION}
rm -rf qemu-${QEMU_VERSION}

# Copy spec file and related files
cp -r $SCRIPT_DIR/qemu-spec/* .

sed -i "s/Version:.*$/Version: ${QEMU_VERSION}/" specs/qemu-kvm.spec

docker rm -f qemu-build

docker run -td \
    --name qemu-build \
    -v $(pwd):/qemu-src \
    registry.gitlab.com/libvirt/libvirt/ci-centos-stream-9

# Build qemu RPM
docker exec -w /qemu-src qemu-build bash -c "
  set -ex
  mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
  cp specs/qemu-kvm.spec ~/rpmbuild/SPECS/
  cp sources/* ~/rpmbuild/SOURCES/
  cp qemu-${QEMU_VERSION}.tar.xz ~/rpmbuild/SOURCES/
  cd ~/rpmbuild/SPECS
  dnf update -y
  dnf -y install createrepo
  dnf builddep -y qemu-kvm.spec
  rpmbuild -ba qemu-kvm.spec
  cd ~/rpmbuild/RPMS
  createrepo --general-compress-type=gz --checksum=sha256 x86_64
"

cd ../

docker cp qemu-build:/root/rpmbuild/RPMS ./rpms-qemu

cat >./rpms-qemu/build-info.json <<EOF
{
<<<<<<< Updated upstream
  "qemu_version": "0:${QEMU_VERSION}-100.el9"
=======
  "qemu_version": "17:${QEMU_VERSION}-100.el9"
>>>>>>> Stashed changes
}
EOF
