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

	k8sv1 "k8s.io/api/core/v1"

	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/testutils"
	"kubevirt.io/kubevirt/tests/decorators"
)

var _ = Describe("[sig-compute]Hypervisor", decorators.SigCompute, func() {

	Context("GetHypervisorContextByName", func() {
		It("should return KVM hypervisor context when name is KVM", func() {
			ctx, err := GetHypervisorContextByName(KVM)
			Expect(err).ToNot(HaveOccurred())
			Expect(ctx).ToNot(BeNil())
			Expect(ctx.Name).To(Equal(KVM))
			Expect(ctx.DevicePath).To(Equal("/dev/kvm"))
			Expect(ctx.DomainType).To(Equal("kvm"))
		})

		It("should return HyperV hypervisor context when name is HYPERVLAYERED", func() {
			ctx, err := GetHypervisorContextByName(HYPERVLAYERED)
			Expect(err).ToNot(HaveOccurred())
			Expect(ctx).ToNot(BeNil())
			Expect(ctx.Name).To(Equal(HYPERVLAYERED))
			Expect(ctx.DevicePath).To(Equal("/dev/mshv"))
			Expect(ctx.DomainType).To(Equal("hyperv"))
		})

		It("should return error when hypervisor name is not found", func() {
			ctx, err := GetHypervisorContextByName("unknown")
			Expect(err).To(HaveOccurred())
			Expect(ctx).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("hypervisor context not found for name: unknown"))
		})

		It("should return error for empty hypervisor name", func() {
			ctx, err := GetHypervisorContextByName("")
			Expect(err).To(HaveOccurred())
			Expect(ctx).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("hypervisor context not found for name:"))
		})
	})

	Context("GetCurrentHypervisorContext", decorators.SigCompute, func() {
		It("should return KVM context when HyperV is not enabled", decorators.Conformance, func() {
			clusterConfig, _, _ := testutils.NewFakeClusterConfigUsingKVConfig(&v1.KubeVirtConfiguration{
				DeveloperConfiguration: &v1.DeveloperConfiguration{
					FeatureGates: []string{},
				},
			})

			ctx := GetCurrentHypervisorContext(clusterConfig)
			Expect(ctx).ToNot(BeNil())
			Expect(ctx.Name).To(Equal(KVM))
			Expect(ctx.DevicePath).To(Equal("/dev/kvm"))
			Expect(ctx.DomainType).To(Equal("kvm"))
		})

		It("should return HyperV context when HyperV is enabled", decorators.HyperVLayered, func() {
			clusterConfig, _, _ := testutils.NewFakeClusterConfigUsingKVConfig(&v1.KubeVirtConfiguration{
				DeveloperConfiguration: &v1.DeveloperConfiguration{
					FeatureGates: []string{"HyperVLayered"},
				},
			})

			ctx := GetCurrentHypervisorContext(clusterConfig)
			Expect(ctx).ToNot(BeNil())
			Expect(ctx.Name).To(Equal(HYPERVLAYERED))
			Expect(ctx.DevicePath).To(Equal("/dev/mshv"))
			Expect(ctx.DomainType).To(Equal("hyperv"))
		})

		It("should not cache the hypervisor context and re-evaluate each call", func() {
			clusterConfig, _, _ := testutils.NewFakeClusterConfigUsingKVConfig(&v1.KubeVirtConfiguration{
				DeveloperConfiguration: &v1.DeveloperConfiguration{
					FeatureGates: []string{},
				},
			})

			ctx1 := GetCurrentHypervisorContext(clusterConfig)
			ctx2 := GetCurrentHypervisorContext(clusterConfig)

			// Since there's no caching, contexts should have the same values but not be identical objects
			Expect(ctx1).ToNot(BeIdenticalTo(ctx2))
			Expect(ctx1.Name).To(Equal(ctx2.Name))
			Expect(ctx1.Name).To(Equal(KVM))
		})
	})

	Context("HypervisorContext.K8sResourceName", func() {
		It("should return correct Kubernetes resource name for KVM", func() {
			ctx := &HypervisorContext{
				Name:       KVM,
				DevicePath: "/dev/kvm",
				DomainType: "kvm",
			}

			resourceName := ctx.K8sResourceName()
			expected := k8sv1.ResourceName("devices.kubevirt.io/kvm")
			Expect(resourceName).To(Equal(expected))
		})

		It("should return correct Kubernetes resource name for HyperV", func() {
			ctx := &HypervisorContext{
				Name:       HYPERVLAYERED,
				DevicePath: "/dev/mshv",
				DomainType: "hyperv",
			}

			resourceName := ctx.K8sResourceName()
			expected := k8sv1.ResourceName("devices.kubevirt.io/mshv")
			Expect(resourceName).To(Equal(expected))
		})

		It("should handle custom hypervisor names", func() {
			ctx := &HypervisorContext{
				Name:       "custom-hypervisor",
				DevicePath: "/dev/custom",
				DomainType: "custom",
			}

			resourceName := ctx.K8sResourceName()
			expected := k8sv1.ResourceName("devices.kubevirt.io/custom-hypervisor")
			Expect(resourceName).To(Equal(expected))
		})
	})

	Context("hypervisor constants and contexts", func() {
		It("should have the correct constant values", func() {
			Expect(KVM).To(Equal("kvm"))
			Expect(HYPERVLAYERED).To(Equal("mshv"))
		})

		It("should have predefined hypervisor contexts initialized", func() {
			Expect(hypervisorContexts).ToNot(BeNil())
			Expect(hypervisorContexts).To(HaveLen(2))

			kvmCtx, exists := hypervisorContexts[KVM]
			Expect(exists).To(BeTrue())
			Expect(kvmCtx.Name).To(Equal(KVM))
			Expect(kvmCtx.DevicePath).To(Equal("/dev/kvm"))
			Expect(kvmCtx.DomainType).To(Equal("kvm"))

			hypervCtx, exists := hypervisorContexts[HYPERVLAYERED]
			Expect(exists).To(BeTrue())
			Expect(hypervCtx.Name).To(Equal(HYPERVLAYERED))
			Expect(hypervCtx.DevicePath).To(Equal("/dev/mshv"))
			Expect(hypervCtx.DomainType).To(Equal("hyperv"))
		})
	})
})
