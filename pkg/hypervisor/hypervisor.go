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

package hypervisor

import v1 "kubevirt.io/api/core/v1"

// Hypervisor interface defines functions needed to tune the virt-launcher pod spec and the libvirt domain XML for a specific hypervisor
type Hypervisor interface {
	// GetK8sResourceName returns the name of the K8s resource representing the hypervisor
	GetK8sResourceName() string
}

func NewHypervisor(hypervisor string) Hypervisor {
	switch hypervisor {
	case v1.MshvL1vhHypervisorName:
		return &MshvL1vhHypervisor{}
	default:
		return &KVMHypervisor{}
	}
}
