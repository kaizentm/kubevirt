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

import (
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "kubevirt.io/api/core/v1"
)

type LauncherHypervisorResources interface {
	GetHypervisorDevice() string
	GetMemoryOverhead(vmi *v1.VirtualMachineInstance, arch string, additionalOverheadRatio *string) resource.Quantity
}

const (
	DefaultHypervisor = "kvm"
)

var hypervisors = map[string]LauncherHypervisorResources{}

// RegisterHypervisor adds a given hypervisor to the hypervisor list.
// In case the hypervisor already exists (based on its name), it overrides the
// existing hypervisor.
func RegisterHypervisor(name string, hypervisor LauncherHypervisorResources) {
	hypervisors[name] = hypervisor
}

// UnregisterHypervisor removes a hypervisor from the hypervisor list.
func UnregisterHypervisor(name string) {
	delete(hypervisors, name)
}

// NewLauncherHypervisorResources returns the LauncherHypervisorResources instance
// for the given hypervisor name. If the hypervisor is not registered or the name
// is empty, it returns the default hypervisor.
func NewLauncherHypervisorResources(hypervisor string) LauncherHypervisorResources {
	if hypervisor == "" {
		hypervisor = DefaultHypervisor
	}

	h, exist := hypervisors[hypervisor]
	if !exist {
		panic("hypervisor " + hypervisor + " is not registered")
	}

	return h
}
