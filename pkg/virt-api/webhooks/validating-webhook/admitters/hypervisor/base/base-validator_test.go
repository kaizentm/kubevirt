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

package base_validator

import (
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"kubevirt.io/client-go/api"
	"kubevirt.io/client-go/kubecli"

	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/testutils"
	"kubevirt.io/kubevirt/pkg/virt-config/featuregate"
)

var _ = Describe("Fine-grained validation of VMI specs", func() {
	config, crdInformer, kvStore := testutils.NewFakeClusterConfigUsingKVConfig(&v1.KubeVirtConfiguration{})
	var (
		namespaceInformer cache.SharedIndexInformer
		virtClient        *kubecli.MockKubevirtClient
		k8sClient         *k8sfake.Clientset
		baseValidator     *BaseValidator
	)

	enableFeatureGate := func(featureGates ...string) {
		kv := testutils.GetFakeKubeVirtClusterConfig(kvStore)
		if kv.Spec.Configuration.DeveloperConfiguration == nil {
			kv.Spec.Configuration.DeveloperConfiguration = &v1.DeveloperConfiguration{}
		}
		if kv.Spec.Configuration.DeveloperConfiguration.FeatureGates == nil {
			kv.Spec.Configuration.DeveloperConfiguration.FeatureGates = featureGates
		} else {
			kv.Spec.Configuration.DeveloperConfiguration.FeatureGates = append(kv.Spec.Configuration.DeveloperConfiguration.FeatureGates, featureGates...)
		}
		testutils.UpdateFakeKubeVirtClusterConfig(kvStore, kv)
	}
	disableFeatureGates := func() {
		kv := testutils.GetFakeKubeVirtClusterConfig(kvStore)
		if kv.Spec.Configuration.DeveloperConfiguration != nil {
			kv.Spec.Configuration.DeveloperConfiguration.FeatureGates = make([]string, 0)
		}
		testutils.UpdateFakeKubeVirtClusterConfig(kvStore, kv)
	}

	BeforeEach(func() {
		namespaceInformer, _ = testutils.NewFakeInformerFor(&k8sv1.Namespace{})
		ns1 := &k8sv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns1",
			},
		}
		ns2 := &k8sv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns2",
			},
		}
		ns3 := &k8sv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns3",
			},
		}
		Expect(namespaceInformer.GetStore().Add(ns1)).To(Succeed())
		Expect(namespaceInformer.GetStore().Add(ns2)).To(Succeed())
		Expect(namespaceInformer.GetStore().Add(ns3)).To(Succeed())

		ctrl := gomock.NewController(GinkgoT())
		k8sClient = k8sfake.NewSimpleClientset()
		virtClient = kubecli.NewMockKubevirtClient(ctrl)
		baseValidator = &BaseValidator{}

		const kubeVirtNamespace = "kubevirt"
		virtClient.EXPECT().AuthorizationV1().Return(k8sClient.AuthorizationV1()).AnyTimes()
	})

	AfterEach(func() {
		disableFeatureGates()
	})

	Context("with Volume", func() {

		BeforeEach(func() {
			enableFeatureGate(featuregate.HostDiskGate)
		})

		DescribeTable("should accept valid volumes",
			func(volumeSource v1.VolumeSource) {
				vmi := api.NewMinimalVMI("testvmi")
				vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
					Name:         "testvolume",
					VolumeSource: volumeSource,
				})

				testutils.AddDataVolumeAPI(crdInformer)
				causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
				Expect(causes).To(BeEmpty())
			},
			Entry("with pvc volume source", v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{}}),
			Entry("with cloud-init volume source", v1.VolumeSource{CloudInitNoCloud: &v1.CloudInitNoCloudSource{UserData: "fake", NetworkData: "fake"}}),
			Entry("with containerDisk volume source", v1.VolumeSource{ContainerDisk: testutils.NewFakeContainerDiskSource()}),
			Entry("with ephemeral volume source", v1.VolumeSource{Ephemeral: &v1.EphemeralVolumeSource{}}),
			Entry("with emptyDisk volume source", v1.VolumeSource{EmptyDisk: &v1.EmptyDiskSource{}}),
			Entry("with dataVolume volume source", v1.VolumeSource{DataVolume: &v1.DataVolumeSource{Name: "fake"}}),
			Entry("with hostDisk volume source", v1.VolumeSource{HostDisk: &v1.HostDisk{Path: "fake", Type: v1.HostDiskExistsOrCreate}}),
			Entry("with configMap volume source", v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: k8sv1.LocalObjectReference{Name: "fake"}}}),
			Entry("with secret volume source", v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "fake"}}),
			Entry("with serviceAccount volume source", v1.VolumeSource{ServiceAccount: &v1.ServiceAccountVolumeSource{ServiceAccountName: "fake"}}),
		)
		It("should allow create a vm using a DataVolume when cdi doesnt exist", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name:         "testvolume",
				VolumeSource: v1.VolumeSource{DataVolume: &v1.DataVolumeSource{Name: "fake"}},
			})

			testutils.RemoveDataVolumeAPI(crdInformer)
			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(BeEmpty())
		})
		It("should reject DataVolume when DataVolume name is not set", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name:         "testvolume",
				VolumeSource: v1.VolumeSource{DataVolume: &v1.DataVolumeSource{Name: ""}},
			})

			testutils.AddDataVolumeAPI(crdInformer)
			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(string(causes[0].Type)).To(Equal("FieldValueRequired"))
			Expect(causes[0].Field).To(Equal("fake[0].name"))
			Expect(causes[0].Message).To(Equal("DataVolume 'name' must be set"))
		})
		It("should reject volume with no volume source set", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testvolume",
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0]"))
		})
		It("should reject volume with multiple volume sources set", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testvolume",
				VolumeSource: v1.VolumeSource{
					ContainerDisk:         testutils.NewFakeContainerDiskSource(),
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0]"))
		})
		It("should reject volumes with duplicate names", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testvolume",
				VolumeSource: v1.VolumeSource{
					ContainerDisk: testutils.NewFakeContainerDiskSource(),
				},
			})
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testvolume",
				VolumeSource: v1.VolumeSource{
					ContainerDisk: testutils.NewFakeContainerDiskSource(),
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[1].name"))
		})

		DescribeTable("should verify cloud-init userdata length", func(userDataLen int, expectedErrors int, base64Encode bool) {
			vmi := api.NewMinimalVMI("testvmi")

			// generate fake userdata
			userdata := ""
			for i := 0; i < userDataLen; i++ {
				userdata = fmt.Sprintf("%sa", userdata)
			}

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{VolumeSource: v1.VolumeSource{CloudInitNoCloud: &v1.CloudInitNoCloudSource{}}})

			if base64Encode {
				vmi.Spec.Volumes[0].VolumeSource.CloudInitNoCloud.UserDataBase64 = base64.StdEncoding.EncodeToString([]byte(userdata))
			} else {
				vmi.Spec.Volumes[0].VolumeSource.CloudInitNoCloud.UserData = userdata
			}

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(expectedErrors))
			for _, cause := range causes {
				Expect(cause.Field).To(ContainSubstring("fake[0].cloudInitNoCloud"))
			}
		},
			Entry("should accept userdata under max limit", 10, 0, false),
			Entry("should accept userdata equal max limit", cloudInitUserMaxLen, 0, false),
			Entry("should reject userdata greater than max limit", cloudInitUserMaxLen+1, 1, false),
			Entry("should accept userdata base64 under max limit", 10, 0, true),
			Entry("should accept userdata base64 equal max limit", cloudInitUserMaxLen, 0, true),
			Entry("should reject userdata base64 greater than max limit", cloudInitUserMaxLen+1, 1, true),
		)

		DescribeTable("should verify cloud-init networkdata length", func(networkDataLen int, expectedErrors int, base64Encode bool) {
			vmi := api.NewMinimalVMI("testvmi")

			// generate fake networkdata
			networkdata := ""
			for i := 0; i < networkDataLen; i++ {
				networkdata = fmt.Sprintf("%sa", networkdata)
			}

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{VolumeSource: v1.VolumeSource{CloudInitNoCloud: &v1.CloudInitNoCloudSource{}}})
			vmi.Spec.Volumes[0].VolumeSource.CloudInitNoCloud.UserData = "#config"

			if base64Encode {
				vmi.Spec.Volumes[0].VolumeSource.CloudInitNoCloud.NetworkDataBase64 = base64.StdEncoding.EncodeToString([]byte(networkdata))
			} else {
				vmi.Spec.Volumes[0].VolumeSource.CloudInitNoCloud.NetworkData = networkdata
			}

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(expectedErrors))
			for _, cause := range causes {
				Expect(cause.Field).To(ContainSubstring("fake[0].cloudInitNoCloud"))
			}
		},
			Entry("should accept networkdata under max limit", 10, 0, false),
			Entry("should accept networkdata equal max limit", cloudInitNetworkMaxLen, 0, false),
			Entry("should reject networkdata greater than max limit", cloudInitNetworkMaxLen+1, 1, false),
			Entry("should accept networkdata base64 under max limit", 10, 0, true),
			Entry("should accept networkdata base64 equal max limit", cloudInitNetworkMaxLen, 0, true),
			Entry("should reject networkdata base64 greater than max limit", cloudInitNetworkMaxLen+1, 1, true),
		)

		It("should reject cloud-init with invalid base64 userdata", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					CloudInitNoCloud: &v1.CloudInitNoCloudSource{
						UserDataBase64: "#######garbage******",
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].cloudInitNoCloud.userDataBase64"))
		})

		It("should reject cloud-init with invalid base64 networkdata", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					CloudInitNoCloud: &v1.CloudInitNoCloudSource{
						UserData:          "fake",
						NetworkDataBase64: "#######garbage******",
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].cloudInitNoCloud.networkDataBase64"))
		})

		It("should reject cloud-init with multiple userdata sources", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					CloudInitNoCloud: &v1.CloudInitNoCloudSource{
						UserData: "fake",
						UserDataSecretRef: &k8sv1.LocalObjectReference{
							Name: "fake",
						},
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].cloudInitNoCloud"))
		})

		It("should reject cloud-init with multiple networkdata sources", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					CloudInitNoCloud: &v1.CloudInitNoCloudSource{
						UserData:    "fake",
						NetworkData: "fake",
						NetworkDataSecretRef: &k8sv1.LocalObjectReference{
							Name: "fake",
						},
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].cloudInitNoCloud"))
		})

		It("should reject hostDisk without required parameters", func() {
			vmi := api.NewMinimalVMI("testvmi")
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					HostDisk: &v1.HostDisk{},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(2))
			Expect(causes[0].Field).To(Equal("fake[0].hostDisk.path"))
			Expect(causes[1].Field).To(Equal("fake[0].hostDisk.type"))
		})

		It("should reject hostDisk without given 'path'", func() {
			vmi := api.NewMinimalVMI("testvmi")
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					HostDisk: &v1.HostDisk{
						Type: v1.HostDiskExistsOrCreate,
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].hostDisk.path"))
		})

		It("should reject hostDisk with invalid type", func() {
			vmi := api.NewMinimalVMI("testvmi")
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					HostDisk: &v1.HostDisk{
						Path: "fakePath",
						Type: "fakeType",
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].hostDisk.type"))
		})

		It("should reject hostDisk when the capacity is specified with a `DiskExists` type", func() {
			vmi := api.NewMinimalVMI("testvmi")
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					HostDisk: &v1.HostDisk{
						Path:     "fakePath",
						Type:     v1.HostDiskExists,
						Capacity: resource.MustParse("1Gi"),
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].hostDisk.capacity"))
		})

		It("should reject a configMap without the configMapName field", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].configMap.name"))
		})

		It("should reject a secret without the secretName field", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].secret.secretName"))
		})

		It("should reject a serviceAccount without the serviceAccountName field", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				VolumeSource: v1.VolumeSource{
					ServiceAccount: &v1.ServiceAccountVolumeSource{},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake[0].serviceAccount.serviceAccountName"))
		})

		It("should reject multiple serviceAccounts", func() {
			vmi := api.NewMinimalVMI("testvmi")

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "sa1",
				VolumeSource: v1.VolumeSource{
					ServiceAccount: &v1.ServiceAccountVolumeSource{ServiceAccountName: "test1"},
				},
			})
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "sa2",
				VolumeSource: v1.VolumeSource{
					ServiceAccount: &v1.ServiceAccountVolumeSource{ServiceAccountName: "test2"},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake"))
		})
	})

	Context("with downwardmetrics virtio serial", func() {
		var vmi *v1.VirtualMachineInstance
		validate := func() []metav1.StatusCause {
			return baseValidator.ValidateDownwardMetrics(k8sfield.NewPath("fake"), &vmi.Spec, config)
		}

		BeforeEach(func() {
			vmi = api.NewMinimalVMI("testvmi")
			vmi.Spec.Domain.Devices.DownwardMetrics = &v1.DownwardMetrics{}
		})

		It("should accept a single virtio serial", func() {
			enableFeatureGate(featuregate.DownwardMetricsFeatureGate)
			causes := validate()
			Expect(causes).To(BeEmpty())
		})

		It("should reject if feature gate is not enabled", func() {
			causes := validate()
			Expect(causes).To(HaveLen(1))
			Expect(causes).To(ContainElement(metav1.StatusCause{Type: metav1.CauseTypeFieldValueInvalid,
				Field:   "fake.domain.devices.downwardMetrics",
				Message: "downwardMetrics virtio serial is not allowed: DownwardMetrics feature gate is not enabled"}))
		})
	})

	Context("with kernel boot defined", func() {
		createKernelBoot := func(kernelArgs, initrdPath, kernelPath, image string) *v1.KernelBoot {
			var kbContainer *v1.KernelBootContainer
			if image != "" || kernelPath != "" || initrdPath != "" {
				kbContainer = &v1.KernelBootContainer{
					Image:      image,
					KernelPath: kernelPath,
					InitrdPath: initrdPath,
				}
			}

			return &v1.KernelBoot{
				KernelArgs: kernelArgs,
				Container:  kbContainer,
			}
		}

		const (
			validKernelArgs   = "args"
			withoutKernelArgs = ""

			validImage   = "image"
			withoutImage = ""

			invalidInitrd = "initrd"
			validInitrd   = "/initrd"
			withoutInitrd = ""

			invalidKernel = "kernel"
			validKernel   = "/kernel"
			withoutKernel = ""
		)

		DescribeTable("", func(kernelBoot *v1.KernelBoot, shouldBeValid bool) {
			kernelBootField := k8sfield.NewPath("spec").Child("domain").Child("firmware").Child("kernelBoot")
			causes := baseValidator.ValidateKernelBoot(kernelBootField, kernelBoot)

			if shouldBeValid {
				Expect(causes).To(BeEmpty())
			} else {
				Expect(causes).ToNot(BeEmpty())
			}
		},
			Entry("without kernel args and null container - should approve",
				createKernelBoot(withoutKernelArgs, withoutInitrd, withoutKernel, withoutImage), true),
			Entry("with kernel args and null container - should reject",
				createKernelBoot(validKernelArgs, withoutInitrd, withoutKernel, withoutImage), false),
			Entry("without kernel args, with container that has image & kernel & initrd defined - should approve",
				createKernelBoot(withoutKernelArgs, validInitrd, validKernel, validImage), true),
			Entry("with kernel args, with container that has image & kernel & initrd defined - should approve",
				createKernelBoot(validKernelArgs, validInitrd, validKernel, validImage), true),
			Entry("with kernel args, with container that has image & kernel defined - should approve",
				createKernelBoot(validKernelArgs, withoutInitrd, validKernel, validImage), true),
			Entry("with kernel args, with container that has image & initrd defined - should approve",
				createKernelBoot(validKernelArgs, validInitrd, withoutKernel, validImage), true),
			Entry("with kernel args, with container that has only image defined - should reject",
				createKernelBoot(validKernelArgs, withoutInitrd, withoutKernel, validImage), false),
			Entry("with invalid kernel path - should reject",
				createKernelBoot(validKernelArgs, validInitrd, invalidKernel, validImage), false),
			Entry("with invalid initrd path - should reject",
				createKernelBoot(validKernelArgs, invalidInitrd, validKernel, validImage), false),
			Entry("with kernel args, with container that has initrd and kernel defined but without image - should reject",
				createKernelBoot(validKernelArgs, validInitrd, validKernel, withoutImage), false),
		)
	})

	Context("with volume", func() {
		var vmi *v1.VirtualMachineInstance

		BeforeEach(func() {
			vmi = api.NewMinimalVMI("testvmi")
		})

		It("should accept a single downwardmetrics volume", func() {
			enableFeatureGate(featuregate.DownwardMetricsFeatureGate)

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testDownwardMetrics",
				VolumeSource: v1.VolumeSource{
					DownwardMetrics: &v1.DownwardMetricsVolumeSource{},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(BeEmpty())
		})

		It("should reject downwardMetrics volumes if the feature gate is not enabled", func() {
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testDownwardMetrics",
				VolumeSource: v1.VolumeSource{
					DownwardMetrics: &v1.DownwardMetricsVolumeSource{},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Message).To(ContainSubstring("downwardMetrics disks are not allowed: DownwardMetrics feature gate is not enabled."))
		})

		It("should reject downwardMetrics volumes if more than one exist", func() {
			enableFeatureGate(featuregate.DownwardMetricsFeatureGate)

			vmi.Spec.Volumes = append(vmi.Spec.Volumes,
				v1.Volume{
					Name: "testDownwardMetrics",
					VolumeSource: v1.VolumeSource{
						DownwardMetrics: &v1.DownwardMetricsVolumeSource{},
					},
				},
				v1.Volume{
					Name: "testDownwardMetrics1",
					VolumeSource: v1.VolumeSource{
						DownwardMetrics: &v1.DownwardMetricsVolumeSource{},
					},
				},
			)
			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Message).To(ContainSubstring("fake must have max one downwardMetric volume set"))
		})

		It("should reject hostDisk volumes if the feature gate is not enabled", func() {
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testHostDisk",
				VolumeSource: v1.VolumeSource{
					HostDisk: &v1.HostDisk{
						Type: v1.HostDiskExistsOrCreate,
						Path: "/hostdisktest.img",
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
		})

		It("should accept hostDisk volumes if the feature gate is enabled", func() {
			enableFeatureGate(featuregate.HostDiskGate)

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testHostDisk",
				VolumeSource: v1.VolumeSource{
					HostDisk: &v1.HostDisk{
						Type: v1.HostDiskExistsOrCreate,
						Path: "/hostdisktest.img",
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(BeEmpty())
		})

		It("should accept sysprep volumes", func() {
			vmi := api.NewMinimalVMI("fake-vmi")
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "sysprep-configmap-volume",
				VolumeSource: v1.VolumeSource{
					Sysprep: &v1.SysprepSource{
						ConfigMap: &k8sv1.LocalObjectReference{
							Name: "test-config",
						},
					},
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(BeEmpty())
		})

		It("should reject CloudInitNoCloud volume if either userData or networkData is missing", func() {
			vmi.Spec.Domain.Devices.Disks = append(vmi.Spec.Domain.Devices.Disks, v1.Disk{
				Name: "testdisk",
			})

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testdisk",
				VolumeSource: v1.VolumeSource{
					CloudInitNoCloud: &v1.CloudInitNoCloudSource{},
				},
			})
			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
		})

		It("should accept CloudInitNoCloud volume if it has only a userData source", func() {
			vmi.Spec.Domain.Devices.Disks = append(vmi.Spec.Domain.Devices.Disks, v1.Disk{
				Name: "testdisk",
			})

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testdisk",
				VolumeSource: v1.VolumeSource{
					CloudInitNoCloud: &v1.CloudInitNoCloudSource{UserData: " "},
				},
			})
			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(BeEmpty())
		})

		It("should accept CloudInitNoCloud volume if it has only a networkData source", func() {
			vmi.Spec.Domain.Devices.Disks = append(vmi.Spec.Domain.Devices.Disks, v1.Disk{
				Name: "testdisk",
			})

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testdisk",
				VolumeSource: v1.VolumeSource{
					CloudInitNoCloud: &v1.CloudInitNoCloudSource{NetworkData: " "},
				},
			})
			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(BeEmpty())
		})

		It("should accept CloudInitNoCloud volume if it has both userData and networkData sources", func() {
			vmi.Spec.Domain.Devices.Disks = append(vmi.Spec.Domain.Devices.Disks, v1.Disk{
				Name: "testdisk",
			})

			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testdisk",
				VolumeSource: v1.VolumeSource{
					CloudInitNoCloud: &v1.CloudInitNoCloudSource{UserData: " ", NetworkData: " "},
				},
			})
			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(BeEmpty())
		})

		It("should accept a single memoryDump volume without a matching disk", func() {
			vmi.Spec.Volumes = append(vmi.Spec.Volumes, v1.Volume{
				Name: "testMemoryDump",
				VolumeSource: v1.VolumeSource{
					MemoryDump: testutils.NewFakeMemoryDumpSource("testMemoryDump"),
				},
			})

			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(BeEmpty())
		})

		It("should reject memoryDump volumes if more than one exist", func() {
			vmi.Spec.Volumes = append(vmi.Spec.Volumes,
				v1.Volume{
					Name: "testMemoryDump",
					VolumeSource: v1.VolumeSource{
						MemoryDump: testutils.NewFakeMemoryDumpSource("testMemoryDump"),
					},
				},
				v1.Volume{
					Name: "testMemoryDump2",
					VolumeSource: v1.VolumeSource{
						MemoryDump: testutils.NewFakeMemoryDumpSource("testMemoryDump2"),
					},
				},
			)
			causes := baseValidator.ValidateVolumes(k8sfield.NewPath("fake"), vmi.Spec.Volumes, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Message).To(ContainSubstring("fake must have max one memory dump volume set"))
		})

	})

	DescribeTable("should validate ACPI", func(acpi *v1.ACPI, volumes []v1.Volume, expectedLen int, expectedMessage string) {
		vmi := api.NewMinimalVMI("testvmi")
		vmi.Spec.Domain.Firmware = &v1.Firmware{ACPI: acpi}
		vmi.Spec.Volumes = volumes
		causes := baseValidator.ValidateFirmwareACPI(k8sfield.NewPath("fake"), &vmi.Spec)
		Expect(causes).To(HaveLen(expectedLen))
		if expectedLen != 0 {
			Expect(causes[0].Message).To(ContainSubstring(expectedMessage))
		}
	},
		Entry("Not set is ok", nil, []v1.Volume{}, 0, ""),
		Entry("ACPI SLIC with Volume match is ok",
			&v1.ACPI{SlicNameRef: "slic"},
			[]v1.Volume{
				{
					Name: "slic",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{SecretName: "secret-slic"},
					},
				},
			}, 0, ""),
		Entry("ACPI MSDM with Volume match is ok",
			&v1.ACPI{SlicNameRef: "msdm"},
			[]v1.Volume{
				{
					Name: "msdm",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{SecretName: "secret-msdm"},
					},
				},
			}, 0, ""),
		Entry("ACPI SLIC without Volume match should fail",
			&v1.ACPI{SlicNameRef: "slic"},
			[]v1.Volume{}, 1, "does not have a matching Volume"),
		Entry("ACPI MSDM without Volume match should fail",
			&v1.ACPI{MsdmNameRef: "msdm"},
			[]v1.Volume{}, 1, "does not have a matching Volume"),
		Entry("ACPI SLIC with wrong Volume type should fail",
			&v1.ACPI{SlicNameRef: "slic"},
			[]v1.Volume{
				{
					Name: "slic",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: k8sv1.LocalObjectReference{Name: "configmap-slic"},
						},
					},
				},
			}, 1, "Volume of unsupported type"),
		Entry("ACPI MSDM with wrong Volume type should fail",
			&v1.ACPI{MsdmNameRef: "msdm"},
			[]v1.Volume{
				{
					Name: "msdm",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: k8sv1.LocalObjectReference{Name: "configmap-msdm"},
						},
					},
				},
			}, 1, "Volume of unsupported type"),
	)
})
