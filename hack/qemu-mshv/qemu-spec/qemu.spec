Name:           qemu
Version:        10.1.50.mshv.v5
Release:        1%{?dist}
Summary:        QEMU with MSHV patch

License:        GPLv2+
URL:            https://www.qemu.org
Source0:        qemu-%{version}.tar.xz

Patch1: 0001-Initial-redhat-build.patch
Patch2: 0002-Enable-disable-devices-for-RHEL.patch
Patch3: 0003-Machine-type-related-general-changes.patch
Patch4: 0004-meson-temporarily-disable-Wunused-function.patch
Patch5: 0005-Remove-upstream-machine-type-versions-for-aarch64-s3.patch
Patch6: 0006-Adapt-versioned-machine-type-macros-for-RHEL.patch
Patch7: 0007-Increase-deletion-schedule-to-3-releases.patch
Patch8: 0008-Add-downstream-aarch64-versioned-virt-machine-types.patch
Patch9: 0009-Add-downstream-ppc64-versioned-spapr-machine-types.patch
Patch10: 0010-Add-downstream-s390x-versioned-s390-ccw-virtio-machi.patch
Patch11: 0011-Add-downstream-x86_64-versioned-pc-q35-machine-types.patch
Patch12: 0012-Revert-meson-temporarily-disable-Wunused-function.patch
Patch13: 0013-Enable-make-check.patch
Patch14: 0014-vfio-cap-number-of-devices-that-can-be-assigned.patch
Patch15: 0015-Add-support-statement-to-help-output.patch
Patch16: 0016-Use-qemu-kvm-in-documentation-instead-of-qemu-system.patch
Patch17: 0017-qcow2-Deprecation-warning-when-opening-v2-images-rw.patch
Patch18: 0018-redhat-allow-5-level-paging-for-TDX-VMs.patch
Patch19: 0019-Add-upstream-compat-bits.patch
Patch20: 0020-Revert-hw-s390x-s390-virtio-ccw-Remove-the-deprecate.patch
Patch21: 0021-Revert-hw-s390x-s390-virtio-ccw-Remove-the-deprecate.patch
Patch22: 0022-Revert-hw-s390x-s390-virtio-ccw-Remove-the-deprecate.patch
Patch23: 0023-redhat-Fix-rhel7.6.0-machine-type.patch
Patch24: 0024-redhat-Compatibility-handling-for-the-s390-ccw-virti.patch
Patch25: 0025-redhat-Add-new-s390-ccw-virtio-rhel9.8.0-machine-typ.patch
Patch26: 0026-hw-core-machine-rhel-machine-types-compat-fix.patch
Patch27: 0027-arm-rhel-machine-type-compat-fix.patch
Patch28: 0028-target-i386-add-compatibility-property-for-arch_capa.patch
Patch29: 0029-target-i386-add-compatibility-property-for-pdcm-feat.patch
Patch30: 0030-arm-create-new-virt-machine-type-for-rhel-9.8.patch
# For RHEL-119369 - [rhel9] Backport "arm/kvm: report registers we failed to set"
Patch31: kvm-arm-kvm-report-registers-we-failed-to-set.patch
# For RHEL-73009 - [IBM 9.8 FEAT] KVM: Implement Control Program Identification (qemu)
Patch32: kvm-qapi-machine-s390x-add-QAPI-event-SCLP_CPI_INFO_AVAI.patch
# For RHEL-73009 - [IBM 9.8 FEAT] KVM: Implement Control Program Identification (qemu)
Patch33: kvm-tests-functional-add-tests-for-SCLP-event-CPI.patch
# For RHEL-122919 - [RHEL 9.8] Windows 11 VM fails to boot up with ramfb='on' with QEMU 10.1
Patch34: kvm-vfio-rename-field-to-num_initial_regions.patch
# For RHEL-122919 - [RHEL 9.8] Windows 11 VM fails to boot up with ramfb='on' with QEMU 10.1
Patch35: kvm-vfio-only-check-region-info-cache-for-initial-region.patch
# For RHEL-105902 - Add new -rhel9.8.0 machine type to qemu-kvm [x86_64]
Patch36: kvm-x86-create-new-pc-q35-machine-type-for-rhel-9.8.patch
# For RHEL-120127 - CVE-2025-11234 qemu-kvm: VNC WebSocket handshake use-after-free [rhel-9.8]
Patch37: kvm-io-move-websock-resource-release-to-close-method.patch
# For RHEL-120127 - CVE-2025-11234 qemu-kvm: VNC WebSocket handshake use-after-free [rhel-9.8]
Patch38: kvm-io-fix-use-after-free-in-websocket-handshake-code.patch
# For RHEL-126593 - [RHEL 9.8] VFIO migration using multifd should be disabled by default
Patch39: kvm-vfio-Disable-VFIO-migration-with-MultiFD-support.patch
# For RHEL-126693 - [RHEL 9]snp guest fail to boot with hugepage
Patch40: kvm-ram-block-attributes-fix-interaction-with-hugetlb-me.patch
# For RHEL-126693 - [RHEL 9]snp guest fail to boot with hugepage
Patch41: kvm-ram-block-attributes-Unify-the-retrieval-of-the-bloc.patch
# For RHEL-129949 - [rhel9] Fix the typo under vfio-pci device's enable-migration option
Patch42: kvm-Fix-the-typo-of-vfio-pci-device-s-enable-migration-o.patch
# For RHEL-133008 - Assertion failure on drain with iothread and I/O load [rhel-9]
Patch43: kvm-block-backend-Fix-race-when-resuming-queued-requests.patch
# For RHEL-133303 - The VM hit io error when do S3-PR integration on the pass-through  failover multipath device [rhel-9]
Patch44: kvm-file-posix-Handle-suspended-dm-multipath-better-for-.patch
# For RHEL-131144 - qemu crash after hot-unplug disk from the multifunction enabled bus [RHEL.9.8]
Patch45: kvm-pcie_sriov-make-pcie_sriov_pf_exit-safe-on-non-SR-IO.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch46: kvm-accel-Add-Meson-and-config-support-for-MSHV-accelera.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch47: kvm-target-i386-emulate-Allow-instruction-decoding-from-.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch48: kvm-target-i386-mshv-Add-x86-decoder-emu-implementation.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch49: kvm-hw-intc-Generalize-APIC-helper-names-from-kvm_-to-ac.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch50: kvm-include-hw-hyperv-Add-MSHV-ABI-header-definitions.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch51: kvm-linux-headers-linux-Add-mshv.h-headers.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch52: kvm-accel-mshv-Add-accelerator-skeleton.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch53: kvm-accel-mshv-Register-memory-region-listeners.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch54: kvm-accel-mshv-Initialize-VM-partition.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch55: kvm-accel-mshv-Add-vCPU-creation-and-execution-loop.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch56: kvm-treewide-rename-qemu_wait_io_event-qemu_wait_io_even.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch57: kvm-accel-mshv-Add-vCPU-signal-handling.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch58: kvm-target-i386-mshv-Add-CPU-create-and-remove-logic.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch59: kvm-target-i386-mshv-Implement-mshv_store_regs.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch60: kvm-target-i386-mshv-Implement-mshv_get_standard_regs.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch61: kvm-target-i386-mshv-Implement-mshv_get_special_regs.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch62: kvm-target-i386-mshv-Implement-mshv_arch_put_registers.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch63: kvm-target-i386-mshv-Set-local-interrupt-controller-stat.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch64: kvm-target-i386-mshv-Register-CPUID-entries-with-MSHV.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch65: kvm-target-i386-mshv-Register-MSRs-with-MSHV.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch66: kvm-target-i386-mshv-Integrate-x86-instruction-decoder-e.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch67: kvm-target-i386-mshv-Write-MSRs-to-the-hypervisor.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch68: kvm-target-i386-mshv-Implement-mshv_vcpu_run.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch69: kvm-accel-mshv-Handle-overlapping-mem-mappings.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch70: kvm-qapi-accel-Allow-to-query-mshv-capabilities.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch71: kvm-target-i386-mshv-Use-preallocated-page-for-hvcall.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch72: kvm-docs-Add-mshv-to-documentation.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch73: kvm-MAINTAINERS-Add-maintainers-for-mshv-accelerator.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch74: kvm-accel-mshv-initialize-thread-name.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch75: kvm-accel-mshv-use-return-value-of-handle_pio_str_read.patch
# For RHEL-132193 - [rhel 9.8]L1VH qemu downstream initial merge RHEL9
Patch76: kvm-monitor-generalize-query-mshv-info-mshv-to-query-acc.patch
# For RHEL-138240 - QEMU crashes when stopping source VM during live migration [rhel-9]
Patch77: kvm-block-Fix-BDS-use-after-free-during-shutdown.patch
# For RHEL-132989 - Expose block limits of block nodes in QMP and qemu-img [rhel-9]
Patch78: kvm-block-Improve-comments-in-BlockLimits.patch
# For RHEL-132989 - Expose block limits of block nodes in QMP and qemu-img [rhel-9]
Patch79: kvm-block-Expose-block-limits-for-images-in-QMP.patch
# For RHEL-132989 - Expose block limits of block nodes in QMP and qemu-img [rhel-9]
Patch80: kvm-qemu-img-info-Optionally-show-block-limits.patch
# For RHEL-132989 - Expose block limits of block nodes in QMP and qemu-img [rhel-9]
Patch81: kvm-qemu-img-info-Add-cache-mode-option.patch
# For RHEL-140187 - Intel IOMMU VM freezes: "call_irq_handler: 3.37 No irq handler for vector"[rhel-9.8]
Patch82: kvm-hw-intc-ioapic-Fix-ACCEL_KERNEL_GSI_IRQFD_POSSIBLE-t.patch
# For RHEL-139057 - [qemu, rhel-9] increase default TSEG size
Patch83: kvm-q35-increase-default-tseg-size.patch
# For RHEL-147422 - virtiofs: processes become stuck in request_wait_answer on virtiofs mounts
Patch84: kvm-vhost-user-make-vhost_set_vring_file-synchronous.patch
# For RHEL-130620 - VM crashes during boot when virtio device is attached through vfio_ccw [rhel-9]
Patch85: kvm-hw-s390x-Fix-a-possible-crash-with-passed-through-vi.patch
# For RHEL-67115 - [network-storage][rbd][core-dump]installation of guest failed sometimes with multiqueue enabled[rhel9.6]
Patch86: kvm-rbd-Run-co-BH-CB-in-the-coroutine-s-AioContext.patch
# For RHEL-67115 - [network-storage][rbd][core-dump]installation of guest failed sometimes with multiqueue enabled[rhel9.6]
Patch87: kvm-curl-Fix-coroutine-waking.patch
# For RHEL-67115 - [network-storage][rbd][core-dump]installation of guest failed sometimes with multiqueue enabled[rhel9.6]
Patch88: kvm-block-io-Take-reqs_lock-for-tracked_requests.patch
# For RHEL-67115 - [network-storage][rbd][core-dump]installation of guest failed sometimes with multiqueue enabled[rhel9.6]
Patch89: kvm-qcow2-Re-initialize-lock-in-invalidate_cache.patch
# For RHEL-67115 - [network-storage][rbd][core-dump]installation of guest failed sometimes with multiqueue enabled[rhel9.6]
Patch90: kvm-qcow2-Fix-cache_clean_timer.patch
# For RHEL-149396 - Migrate SCSI PR state and preempt reservation upon live migration [rhel-9]
Patch91: kvm-scsi-generalize-scsi_SG_IO_FROM_DEV-to-scsi_SG_IO.patch
# For RHEL-149396 - Migrate SCSI PR state and preempt reservation upon live migration [rhel-9]
Patch92: kvm-scsi-add-error-reporting-to-scsi_SG_IO.patch
# For RHEL-149396 - Migrate SCSI PR state and preempt reservation upon live migration [rhel-9]
Patch93: kvm-scsi-track-SCSI-reservation-state-for-live-migration.patch
# For RHEL-149396 - Migrate SCSI PR state and preempt reservation upon live migration [rhel-9]
Patch94: kvm-scsi-save-load-SCSI-reservation-state.patch
# For RHEL-149396 - Migrate SCSI PR state and preempt reservation upon live migration [rhel-9]
Patch95: kvm-docs-add-SCSI-migrate-pr-documentation.patch
# For RHEL-151679 - [rhel-9.8] Regression in BLOCK_IO_ERROR event delivery with (w|r)error setting of 'stop' or 'enospc' due to event rate limiting
Patch96: kvm-block-Never-drop-BLOCK_IO_ERROR-with-action-stop-for.patch
# For HyperV-Direct backend
Patch97:         revert-overlapping-mem-mappings.patch


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

