Name:           qemu
Version:        10.1.50.mshv.v5
Release:        1%{?dist}
Summary:        QEMU with MSHV patch

License:        GPLv2+
URL:            https://www.qemu.org
Source0:        qemu-%{version}.tar.xz
Patch0:         revert-overlapping-mem-mappings.patch


BuildRequires: gcc 
BuildRequires: make 
BuildRequires: meson 
BuildRequires: ninja-build 
BuildRequires: glib2-devel 
BuildRequires: pixman-devel 
BuildRequires: zlib-devel 
BuildRequires: libfdt-devel 
BuildRequires: libaio-devel 
BuildRequires: liburing-devel 
BuildRequires: libseccomp-devel 
BuildRequires: libcap-ng-devel 
BuildRequires: nettle-devel 
BuildRequires: gnutls-devel 
BuildRequires: libgcrypt-devel 
BuildRequires: numactl-devel 
BuildRequires: libxml2-devel 
BuildRequires: usbredir-devel 
BuildRequires: libusb1-devel 
BuildRequires: libepoxy-devel 
BuildRequires: libattr-devel 
BuildRequires: python3 
BuildRequires: python3-setuptools 
BuildRequires: python3-tomli 
BuildRequires: pkgconfig 
BuildRequires: bzip2 
BuildRequires: xz 
BuildRequires: findutils


BuildArch:      x86_64

%description
Packaging the latest (at time of writing) version of qemu that is not available from Red Hat repository

%prep
%setup -q -n %{name}-%{version}
%patch 0 -p1



%build
./configure \
    --target-list=x86_64-softmmu \
    --disable-xen \
    --disable-vnc-jpeg \
    --enable-mshv \
    --disable-gtk \
    --disable-libiscsi
cd build
make -j

%install
export DESTDIR=$RPM_BUILD_ROOT
make install

%clean
rm -rf $RPM_BUILD_ROOT

# --- Files sections ---
%files
%defattr(-,root,root,-)
%license COPYING COPYING.LIB
%doc README.rst
/usr/local/bin/
/usr/local/libexec/
/usr/local/share
/usr/local/include

