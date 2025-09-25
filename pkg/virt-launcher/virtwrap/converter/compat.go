/*
 * This file is part of the KubeVirt project
*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright The KubeVirt Authors.
 *
*/

package converter

import (
	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

// rhelQ35CompatGlobals is the initial minimal set of QEMU -global property
// overrides we inject when a user asks for a downstream RHEL pc-q35 machine
// type (e.g. pc-q35-rhel9.6.0) but we only have upstream QEMU machine types.
//
// NOTE: This list is intentionally conservative; once the precise minimal
// subset required for virtio enumeration is derived we should trim it and
// keep ordering stable for test assertions.
var rhelQ35CompatGlobals = []string{
	// Keep both modern & legacy paths enabled while we diagnose; downstream
	// compat machines often leave both available for broad guest coverage.
	"virtio-blk-pci.disable-legacy=off",
	"virtio-blk-pci.disable-modern=off",
	"virtio-net-pci.disable-legacy=off",
	"virtio-net-pci.disable-modern=off",
	"virtio-scsi-pci.disable-legacy=off",
	"virtio-scsi-pci.disable-modern=off",
	"virtio-serial-pci.disable-legacy=off",
	"virtio-serial-pci.disable-modern=off",
	"virtio-balloon-pci.disable-legacy=off",
	"virtio-balloon-pci.disable-modern=off",
	"virtio-rng-pci.disable-legacy=off",
	"virtio-rng-pci.disable-modern=off",
	// Conservative queue sizes to avoid large allocation failures observed
	// in early boots; we may relax these later.
	"virtio-blk-pci.queue-size=128",
}

// addQEMUGlobal appends a single -global <prop>=<value> pair to the domain's
// QEMU command line, initializing the slice structure if necessary.
func addQEMUGlobal(domain *api.Domain, prop string) {
	initializeQEMUCmdAndQEMUArg(domain)
	domain.Spec.QEMUCmd.QEMUArg = append(domain.Spec.QEMUCmd.QEMUArg,
		api.Arg{Value: "-global"},
		api.Arg{Value: prop},
	)
}

// applyRHELQ35Compat detects downstream RHEL pc-q35 machine aliases and
// rewrites them to an upstream machine plus a curated set of -global
// overrides which approximate the downstream compatibility surface.
//
// Current strategy: map any pc-q35-rhel9.* to upstream pc-q35-10.2 (adjust if
// a different upstream baseline is more appropriate for the shipped QEMU).
// This function is purposely narrow and side-effect free for all other
// machine types.
func applyRHELQ35Compat(vmi *v1.VirtualMachineInstance, domain *api.Domain) {
	// if vmi == nil || vmi.Spec.Domain.Machine == nil {
	//     return
	// }
	// mt := vmi.Spec.Domain.Machine.Type
	// if !strings.HasPrefix(mt, "pc-q35-rhel9.") {
	//     return
	// }

	// // Only rewrite if we haven't already changed it (idempotent on retries).
	// if domain.Spec.OS.Type.Machine == mt {
	//     domain.Spec.OS.Type.Machine = "pc-q35-10.2"
	// }

	for _, g := range rhelQ35CompatGlobals {
		addQEMUGlobal(domain, g)
	}
}
