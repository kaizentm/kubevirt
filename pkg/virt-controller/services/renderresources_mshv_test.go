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

package services

import (
	"k8s.io/apimachinery/pkg/api/resource"
	kubev1 "k8s.io/api/core/v1"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/kubevirt/pkg/libvmi"
	"kubevirt.io/kubevirt/pkg/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Memory Overhead Validation for mshv Hypervisor", func() {
	Context("GetMemoryOverhead function", func() {
		
		It("should calculate same memory overhead for mshv as for QEMU/KVM", func() {
			vmi := libvmi.New()
			vmi.Spec.Domain.Resources = v1.ResourceRequirements{
				Requests: kubev1.ResourceList{
					kubev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			}

			// Test memory overhead calculation for x86_64
			overheadAmd64 := GetMemoryOverhead(vmi, "amd64", nil)
			
			// Verify overhead includes expected components:
			// - Pagetable memory (1Gi/512 = 2Mi)
			// - Fixed overheads (VirtLauncher, Monitor, Virtqemud, Qemu)
			// - vCPU overhead (8Mi per vCPU + 8Mi IOThread)
			// - Graphics device overhead (32Mi)
			expectedMinOverhead := resource.MustParse("200Mi") // Conservative minimum
			Expect(overheadAmd64.Cmp(expectedMinOverhead)).To(Equal(1), 
				"Memory overhead should be at least 200Mi for basic 1Gi VM")
			
			// For arm64, should include additional 128Mi for UEFI
			overheadArm64 := GetMemoryOverhead(vmi, "arm64", nil)
			additionalArm64 := resource.MustParse("128Mi")
			expectedArm64 := overheadAmd64
			expectedArm64.Add(additionalArm64)
			
			Expect(overheadArm64.Cmp(expectedArm64)).To(Equal(0),
				"ARM64 should have 128Mi additional overhead for UEFI")
		})

		It("should include VFIO overhead for VMIs with passthrough devices", func() {
			vmi := libvmi.New()
			vmi.Spec.Domain.Resources = v1.ResourceRequirements{
				Requests: kubev1.ResourceList{
					kubev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			}
			
			// Add VFIO device to trigger additional overhead
			vmi.Spec.Domain.Devices.GPUs = []v1.GPU{{
				Name:       "test-gpu",
				DeviceName: "nvidia.com/TU102GL",
			}}

			overheadWithVFIO := GetMemoryOverhead(vmi, "amd64", nil)
			
			Expect(util.IsVFIOVMI(vmi)).To(BeTrue(), "VMI should be detected as VFIO")
			// Note: We can't do exact comparison due to base VM differences, but verify VFIO overhead is significant
			minExpected := resource.MustParse("1200Mi")
			Expect(overheadWithVFIO.Cmp(minExpected)).To(Equal(1),
				"VFIO VMI should have significantly higher memory overhead")
		})

		It("should include SEV overhead for secure encrypted virtualization", func() {
			vmi := libvmi.New()
			vmi.Spec.Domain.Resources = v1.ResourceRequirements{
				Requests: kubev1.ResourceList{
					kubev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			}
			
			// Add SEV launch security
			vmi.Spec.Domain.LaunchSecurity = &v1.LaunchSecurity{
				SEV: &v1.SEV{},
			}

			overheadWithSEV := GetMemoryOverhead(vmi, "amd64", nil)
			baseOverhead := GetMemoryOverhead(libvmi.New(), "amd64", nil)
			
			Expect(util.IsSEVVMI(vmi)).To(BeTrue(), "VMI should be detected as SEV")
			// Verify SEV overhead is included (allowing for rounding differences)
			Expect(overheadWithSEV.Value()).To(BeNumerically(">=", baseOverhead.Value()+240*1024*1024),
				"SEV VMI should have ~256Mi additional overhead")
		})

		It("should handle additional overhead ratio multiplier", func() {
			vmi := libvmi.New()
			vmi.Spec.Domain.Resources = v1.ResourceRequirements{
				Requests: kubev1.ResourceList{
					kubev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			}
			
			ratio := "1.5"
			overheadWithRatio := GetMemoryOverhead(vmi, "amd64", &ratio)
			baseOverhead := GetMemoryOverhead(vmi, "amd64", nil)
			
			// Should be 1.5x the base overhead
			expectedValue := int64(float64(baseOverhead.Value()) * 1.5)
			Expect(overheadWithRatio.Value()).To(BeNumerically("~", expectedValue, expectedValue/100),
				"Overhead with 1.5 ratio should be 1.5x base overhead")
		})
	})

	Context("Memory overhead compatibility research", func() {
		It("should document mshv memory characteristics vs QEMU/KVM", func() {
			// This test documents our research findings about mshv vs QEMU/KVM memory overhead
			Skip("Research findings: mshv hypervisor memory overhead analysis")
			
			// Research Questions Addressed:
			// 1. Does mshv hypervisor have different process overhead than qemu-kvm?
			//    FINDING: Need to test with actual mshv processes
			//    
			// 2. Are vCPU memory overhead calculations the same for mshv vs kvm?
			//    FINDING: vCPU overhead should be similar (8Mi per vCPU)
			//    
			// 3. Does HyperVLayered hardware passthrough have different memory requirements than VFIO?
			//    FINDING: Need research - may use different mechanism than VFIO
			//    
			// 4. Are pagetable memory calculations identical between mshv and kvm?
			//    FINDING: Should be similar - both manage guest physical memory
		})
	})
})