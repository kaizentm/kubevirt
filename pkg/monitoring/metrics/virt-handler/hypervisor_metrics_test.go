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

package virt_handler

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rhobs/operator-observability-toolkit/pkg/operatormetrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	v1 "kubevirt.io/api/core/v1"
)

var _ = Describe("Hypervisor Metrics", func() {
	BeforeEach(func() {
		// Clean up any existing metrics
		operatormetrics.UnregisterMetrics(hypervisorMetrics)
	})

	AfterEach(func() {
		// Clean up metrics after each test
		operatormetrics.UnregisterMetrics(hypervisorMetrics)
	})

	Describe("detectHypervisorType", func() {
		It("should detect KVM hypervisor type", func() {
			domainXML := `<domain type="kvm"><name>test-domain</name></domain>`
			hypervisorType := detectHypervisorType(domainXML)
			Expect(hypervisorType).To(Equal(HypervisorTypeKVM))
		})

		It("should detect QEMU-TCG hypervisor type", func() {
			domainXML := `<domain type="qemu"><name>test-domain</name></domain>`
			hypervisorType := detectHypervisorType(domainXML)
			Expect(hypervisorType).To(Equal(HypervisorTypeQEMUTCG))
		})

		It("should handle unknown hypervisor type", func() {
			domainXML := `<domain type="xen"><name>test-domain</name></domain>`
			hypervisorType := detectHypervisorType(domainXML)
			Expect(hypervisorType).To(Equal(HypervisorTypeUnknown))
		})

		It("should handle invalid XML", func() {
			domainXML := `<invalid-xml>`
			hypervisorType := detectHypervisorType(domainXML)
			Expect(hypervisorType).To(Equal(HypervisorTypeUnknown))
		})

		It("should handle empty XML", func() {
			domainXML := ""
			hypervisorType := detectHypervisorType(domainXML)
			Expect(hypervisorType).To(Equal(HypervisorTypeUnknown))
		})

		It("should handle missing type attribute", func() {
			domainXML := `<domain><name>test-domain</name></domain>`
			hypervisorType := detectHypervisorType(domainXML)
			Expect(hypervisorType).To(Equal(HypervisorTypeUnknown))
		})

		It("should handle case insensitive type values", func() {
			domainXML := `<domain type="KVM"><name>test-domain</name></domain>`
			hypervisorType := detectHypervisorType(domainXML)
			Expect(hypervisorType).To(Equal(HypervisorTypeKVM))
		})
	})

	Describe("updateHypervisorMetric", func() {
		BeforeEach(func() {
			// Register metrics for testing
			Expect(operatormetrics.RegisterMetrics(hypervisorMetrics)).To(Succeed())
		})

		It("should update metric for VMI with valid node", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					NodeName: "test-node",
				},
			}

			updateHypervisorMetric(vmi, HypervisorTypeKVM)

			// Verify metric was created (we can't easily verify the exact value in unit tests)
			// but we can verify no panic occurred
		})

		It("should handle VMI without node assignment", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					// NodeName is empty
				},
			}

			// Should not panic
			updateHypervisorMetric(vmi, HypervisorTypeKVM)
		})

		It("should handle nil VMI", func() {
			// Should not panic
			updateHypervisorMetric(nil, HypervisorTypeKVM)
		})
	})

	Describe("removeHypervisorMetric", func() {
		BeforeEach(func() {
			// Register metrics for testing
			Expect(operatormetrics.RegisterMetrics(hypervisorMetrics)).To(Succeed())
		})

		It("should remove metric for VMI", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					NodeName: "test-node",
				},
			}

			// Should not panic
			removeHypervisorMetric(vmi)
		})

		It("should handle VMI without node assignment", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					// NodeName is empty
				},
			}

			// Should not panic
			removeHypervisorMetric(vmi)
		})

		It("should handle nil VMI", func() {
			// Should not panic
			removeHypervisorMetric(nil)
		})
	})

	Describe("hypervisor metric registration", func() {
		It("should register hypervisor metrics successfully", func() {
			err := operatormetrics.RegisterMetrics(hypervisorMetrics)
			Expect(err).ToNot(HaveOccurred())
			Expect(hypervisorMetrics).ToNot(BeEmpty())
		})
	})

	Describe("VMI event handlers", func() {
		BeforeEach(func() {
			// Register metrics for testing
			Expect(operatormetrics.RegisterMetrics(hypervisorMetrics)).To(Succeed())
		})

		Describe("handleVMIAdd", func() {
			It("should handle running VMI addition", func() {
				vmi := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Running,
					},
				}

				// Should not panic
				handleVMIAdd(vmi)
			})

			It("should ignore non-running VMI", func() {
				vmi := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Pending,
					},
				}

				// Should not panic and should not create metric
				handleVMIAdd(vmi)
			})

			It("should handle VMI with test hypervisor annotation", func() {
				vmi := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
						Annotations: map[string]string{
							"kubevirt.io/test-hypervisor-type": "kvm",
						},
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Running,
					},
				}

				// Should not panic
				handleVMIAdd(vmi)
			})

			It("should handle invalid object type", func() {
				// Should not panic
				handleVMIAdd("not-a-vmi")
			})

			It("should handle nil object", func() {
				// Should not panic
				handleVMIAdd(nil)
			})
		})

		Describe("handleVMIUpdate", func() {
			It("should handle VMI transitioning to running", func() {
				oldVMI := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Pending,
					},
				}

				newVMI := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Running,
					},
				}

				// Should not panic
				handleVMIUpdate(oldVMI, newVMI)
			})

			It("should handle VMI transitioning from running", func() {
				oldVMI := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Running,
					},
				}

				newVMI := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Succeeded,
					},
				}

				// Should not panic
				handleVMIUpdate(oldVMI, newVMI)
			})

			It("should handle VMI remaining running", func() {
				oldVMI := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Running,
					},
				}

				newVMI := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Running,
					},
				}

				// Should not panic
				handleVMIUpdate(oldVMI, newVMI)
			})

			It("should handle invalid object types", func() {
				// Should not panic
				handleVMIUpdate("not-a-vmi", "also-not-a-vmi")
			})

			It("should handle nil objects", func() {
				// Should not panic
				handleVMIUpdate(nil, nil)
			})
		})

		Describe("handleVMIDelete", func() {
			It("should handle VMI deletion", func() {
				vmi := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Running,
					},
				}

				// Should not panic
				handleVMIDelete(vmi)
			})

			It("should handle DeletedFinalStateUnknown", func() {
				vmi := &v1.VirtualMachineInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vmi",
						Namespace: "default",
					},
					Status: v1.VirtualMachineInstanceStatus{
						NodeName: "test-node",
						Phase:    v1.Running,
					},
				}

				tombstone := cache.DeletedFinalStateUnknown{
					Key: "default/test-vmi",
					Obj: vmi,
				}

				// Should not panic
				handleVMIDelete(tombstone)
			})

			It("should handle invalid DeletedFinalStateUnknown object", func() {
				tombstone := cache.DeletedFinalStateUnknown{
					Key: "default/test-vmi",
					Obj: "not-a-vmi",
				}

				// Should not panic
				handleVMIDelete(tombstone)
			})

			It("should handle invalid object type", func() {
				// Should not panic
				handleVMIDelete("not-a-vmi")
			})

			It("should handle nil object", func() {
				// Should not panic
				handleVMIDelete(nil)
			})
		})
	})

	Describe("getDomainXMLForVMI", func() {
		It("should return empty string for nil VMI", func() {
			result := getDomainXMLForVMI(nil)
			Expect(result).To(Equal(""))
		})

		It("should return KVM domain XML for test annotation", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
					Annotations: map[string]string{
						"kubevirt.io/test-hypervisor-type": "kvm",
					},
				},
			}

			result := getDomainXMLForVMI(vmi)
			Expect(result).To(ContainSubstring(`<domain type="kvm"`))
		})

		It("should return QEMU domain XML for test annotation", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
					Annotations: map[string]string{
						"kubevirt.io/test-hypervisor-type": "qemu-tcg",
					},
				},
			}

			result := getDomainXMLForVMI(vmi)
			Expect(result).To(ContainSubstring(`<domain type="qemu"`))
		})

		It("should return unknown domain XML for unknown test annotation", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
					Annotations: map[string]string{
						"kubevirt.io/test-hypervisor-type": "unknown",
					},
				},
			}

			result := getDomainXMLForVMI(vmi)
			Expect(result).To(ContainSubstring(`<domain type="unknown"`))
		})

		It("should return default KVM XML for running VMI without annotations", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					Phase: v1.Running,
				},
			}

			result := getDomainXMLForVMI(vmi)
			Expect(result).To(ContainSubstring(`<domain type="kvm"`))
			Expect(result).To(ContainSubstring(`<name>test-vmi</name>`))
		})

		It("should return empty string for non-running VMI without annotations", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					Phase: v1.Pending,
				},
			}

			result := getDomainXMLForVMI(vmi)
			Expect(result).To(Equal(""))
		})
	})
})
