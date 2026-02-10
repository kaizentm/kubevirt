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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "kubevirt.io/api/core/v1"
)

var _ = Describe("Test hypervisor registration", func() {
	var testHypervisor *mockHypervisor

	BeforeEach(func() {
		testHypervisor = &mockHypervisor{}
	})

	It("should register and retrieve hypervisor", func() {
		RegisterHypervisor("test", testHypervisor)
		retrieved := NewLauncherHypervisorResources("test")
		Expect(retrieved).To(Equal(testHypervisor))
	})

	It("should use default hypervisor for empty string", func() {
		defaultHypervisor := &mockHypervisor{}
		RegisterHypervisor(DefaultHypervisor, defaultHypervisor)
		retrieved := NewLauncherHypervisorResources("")
		Expect(retrieved).To(Equal(defaultHypervisor))
	})

	It("should panic when hypervisor is not registered", func() {
		UnregisterHypervisor("nonexistent")
		Expect(func() {
			NewLauncherHypervisorResources("nonexistent")
		}).To(Panic())
	})

	It("should override existing hypervisor on re-registration", func() {
		firstHypervisor := &mockHypervisor{}
		secondHypervisor := &mockHypervisor{}
		RegisterHypervisor("override-test", firstHypervisor)
		RegisterHypervisor("override-test", secondHypervisor)
		retrieved := NewLauncherHypervisorResources("override-test")
		Expect(retrieved).To(Equal(secondHypervisor))
	})

	It("should unregister hypervisor", func() {
		RegisterHypervisor("to-remove", testHypervisor)
		UnregisterHypervisor("to-remove")
		Expect(func() {
			NewLauncherHypervisorResources("to-remove")
		}).To(Panic())
	})
})

type mockHypervisor struct{}

func (m *mockHypervisor) GetHypervisorDevice() string {
	return "mock"
}

func (m *mockHypervisor) GetMemoryOverhead(vmi *v1.VirtualMachineInstance, arch string, additionalOverheadRatio *string) resource.Quantity {
	return resource.Quantity{}
}
