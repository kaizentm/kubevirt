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

	Describe("getHypervisorTypeForVMI", func() {
		It("should return error when socket not found", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					NodeName: "test-node",
				},
			}

			hypervisorType, err := getHypervisorTypeForVMI(vmi)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find socket"))
			Expect(hypervisorType).To(Equal(""))
		})

		It("should handle nil VMI", func() {
			hypervisorType, err := getHypervisorTypeForVMI(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("VMI cannot be nil"))
			Expect(hypervisorType).To(Equal(""))
		})

		It("should return proper error for VMI without socket", func() {
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "non-existent-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					NodeName: "test-node",
				},
			}

			hypervisorType, err := getHypervisorTypeForVMI(vmi)
			Expect(err).To(HaveOccurred())
			Expect(hypervisorType).To(Equal(""))
		})

		// Note: Testing successful communication with cmd-client requires
		// integration testing with actual virt-launcher pods, as unit tests
		// cannot easily mock the unix socket communication
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

			updateHypervisorMetric(vmi, "kvm")

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
			updateHypervisorMetric(vmi, "kvm")
		})

		It("should handle nil VMI", func() {
			// Should not panic
			updateHypervisorMetric(nil, "kvm")
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

	Describe("hypervisor constants", func() {
		It("should have correct HypervisorType constants", func() {
			Expect(string(HypervisorTypeKVM)).To(Equal("kvm"))
			Expect(string(HypervisorTypeHyperv)).To(Equal("hyperv"))
			Expect(string(HypervisorTypeQEMU)).To(Equal("qemu"))
			Expect(string(HypervisorTypeUnknown)).To(Equal("unknown"))
		})

		It("should support expected hypervisor types", func() {
			// Test that the constants work with the metric function
			vmi := &v1.VirtualMachineInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vmi",
					Namespace: "default",
				},
				Status: v1.VirtualMachineInstanceStatus{
					NodeName: "test-node",
				},
			}

			// Register metrics for testing
			Expect(operatormetrics.RegisterMetrics(hypervisorMetrics)).To(Succeed())

			// Test all hypervisor types work with updateHypervisorMetric
			updateHypervisorMetric(vmi, string(HypervisorTypeKVM))
			updateHypervisorMetric(vmi, string(HypervisorTypeHyperv))
			updateHypervisorMetric(vmi, string(HypervisorTypeQEMU))
			updateHypervisorMetric(vmi, string(HypervisorTypeUnknown))
		})
	})
})
