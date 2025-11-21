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

package arch_validators

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/api"

	"kubevirt.io/kubevirt/pkg/pointer"
	"kubevirt.io/kubevirt/pkg/testutils"
)

var _ = Describe("Arch-specific validations ", func() {

	config, _, _ := testutils.NewFakeClusterConfigUsingKVConfig(&v1.KubeVirtConfiguration{})

	Context("Watchdog device validation", func() {
		var vmi *v1.VirtualMachineInstance

		BeforeEach(func() {
			vmi = api.NewMinimalVMI("testvmi")
		})

		DescribeTable("validate for amd64",
			func(watchdog *v1.Watchdog, expectedMessage string, shouldReject bool) {
				vmi.Spec.Architecture = "amd64"
				vmi.Spec.Domain.Devices.Watchdog = watchdog
				archValidator := NewArchValidator(vmi.Spec.Architecture)
				causes := archValidator.ValidateVirtualMachineInstanceArchSetting(k8sfield.NewPath("fake"), &vmi.Spec, config)

				if shouldReject {
					Expect(causes).To(HaveLen(1))
					Expect(causes[0].Field).To(Equal("fake.domain.devices.watchdog"))
					Expect(causes[0].Message).To(Equal(expectedMessage))
				} else {
					Expect(causes).To(BeEmpty())
				}
			},
			Entry("I6300ESB is accepted", &v1.Watchdog{
				Name: "w1",
				WatchdogDevice: v1.WatchdogDevice{
					I6300ESB: &v1.I6300ESBWatchdog{Action: v1.WatchdogActionPoweroff},
				},
			}, "", false),

			Entry("Diag288 is rejected", &v1.Watchdog{
				Name: "w2",
				WatchdogDevice: v1.WatchdogDevice{
					Diag288: &v1.Diag288Watchdog{Action: v1.WatchdogActionPoweroff},
				},
			}, "amd64 only supports I6300ESB watchdog device", true),

			Entry("no watchdog configured", nil, "", false),
		)

		DescribeTable("validate for s390x",
			func(watchdog *v1.Watchdog, expectedMessage string, shouldReject bool) {
				vmi.Spec.Architecture = "s390x"
				vmi.Spec.Domain.Devices.Watchdog = watchdog
				archValidator := NewArchValidator(vmi.Spec.Architecture)
				causes := archValidator.ValidateVirtualMachineInstanceArchSetting(k8sfield.NewPath("fake"), &vmi.Spec, config)

				if shouldReject {
					Expect(causes).To(HaveLen(1))
					Expect(causes[0].Field).To(Equal("fake.domain.devices.watchdog"))
					Expect(causes[0].Message).To(Equal(expectedMessage))
				} else {
					Expect(causes).To(BeEmpty())
				}
			},
			Entry("Diag288 is accepted", &v1.Watchdog{
				Name: "w3",
				WatchdogDevice: v1.WatchdogDevice{
					Diag288: &v1.Diag288Watchdog{Action: v1.WatchdogActionPoweroff},
				},
			}, "", false),

			Entry("I6300ESB is rejected", &v1.Watchdog{
				Name: "w4",
				WatchdogDevice: v1.WatchdogDevice{
					I6300ESB: &v1.I6300ESBWatchdog{Action: v1.WatchdogActionPoweroff},
				},
			}, "s390x only supports Diag288 watchdog device", true),

			Entry("no watchdog configured", nil, "", false),
		)

		DescribeTable("validate for arm64",
			func(watchdog *v1.Watchdog, expectedMessage string, shouldReject bool) {
				vmi.Spec.Architecture = "arm64"
				vmi.Spec.Domain.Devices.Watchdog = watchdog
				archValidator := NewArchValidator(vmi.Spec.Architecture)
				causes := archValidator.ValidateVirtualMachineInstanceArchSetting(k8sfield.NewPath("fake"), &vmi.Spec, config)

				if shouldReject {
					Expect(causes).To(HaveLen(1))
					Expect(causes[0].Field).To(Equal("fake.domain.devices.watchdog"))
					Expect(causes[0].Message).To(Equal(expectedMessage))
				} else {
					Expect(causes).To(BeEmpty())
				}
			},
			Entry("I6300ESB is rejected", &v1.Watchdog{
				Name: "w5",
				WatchdogDevice: v1.WatchdogDevice{
					I6300ESB: &v1.I6300ESBWatchdog{Action: v1.WatchdogActionPoweroff},
				},
			}, "Arm64 not support Watchdog device", true),

			Entry("Diag288 is rejected", &v1.Watchdog{
				Name: "w6",
				WatchdogDevice: v1.WatchdogDevice{
					Diag288: &v1.Diag288Watchdog{Action: v1.WatchdogActionPoweroff},
				},
			}, "Arm64 not support Watchdog device", true),

			Entry("no watchdog configured", nil, "", false),
		)
	})

	Context("specific verification for Arm64", func() {
		var vmi *v1.VirtualMachineInstance
		archValidator := NewArchValidator("arm64")

		BeforeEach(func() {
			vmi = api.NewMinimalVMI("testvmi")
		})

		It("should reject BIOS bootloader", func() {
			vmi.Spec.Domain.Firmware = &v1.Firmware{
				Bootloader: &v1.Bootloader{
					BIOS: &v1.BIOS{},
				},
			}

			causes := archValidator.ValidateVirtualMachineInstanceArchSetting(k8sfield.NewPath("fake"), &vmi.Spec, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake.domain.firmware.bootloader.bios"))
			Expect(causes[0].Message).To(Equal("Arm64 does not support bios boot, please change to uefi boot"))
		})

		// When setting UEFI default bootloader, UEFI secure bootloader would be applied which is not supported on Arm64
		It("should reject UEFI default bootloader", func() {
			vmi.Spec.Domain.Firmware = &v1.Firmware{
				Bootloader: &v1.Bootloader{
					EFI: &v1.EFI{},
				},
			}

			causes := archValidator.ValidateVirtualMachineInstanceArchSetting(k8sfield.NewPath("fake"), &vmi.Spec, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake.domain.firmware.bootloader.efi.secureboot"))
			Expect(causes[0].Message).To(Equal("UEFI secure boot is currently not supported on aarch64 Arch"))
		})

		It("should reject UEFI secure bootloader", func() {
			vmi.Spec.Domain.Firmware = &v1.Firmware{
				Bootloader: &v1.Bootloader{
					EFI: &v1.EFI{
						SecureBoot: pointer.P(true),
					},
				},
			}

			causes := archValidator.ValidateVirtualMachineInstanceArchSetting(k8sfield.NewPath("fake"), &vmi.Spec, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake.domain.firmware.bootloader.efi.secureboot"))
			Expect(causes[0].Message).To(Equal("UEFI secure boot is currently not supported on aarch64 Arch"))
		})

		DescribeTable("validating cpu model with", func(model string, expectedLen int) {
			vmi.Spec.Domain.CPU = &v1.CPU{Model: model}

			causes := archValidator.ValidateVirtualMachineInstanceArchSetting(k8sfield.NewPath("fake"), &vmi.Spec, config)
			Expect(causes).To(HaveLen(expectedLen))
			if expectedLen != 0 {
				Expect(causes[0].Field).To(Equal("fake.domain.cpu.model"))
				Expect(causes[0].Message).To(Equal(fmt.Sprintf("currently, %v is the only model supported on Arm64", v1.CPUModeHostPassthrough)))
			}
		},
			Entry("host-model should get rejected with arm64", "host-model", 1),
			Entry("named model should get rejected with arm64", "Cooperlake", 1),
			Entry("host-passthrough should be accepted with arm64", "host-passthrough", 0),
			Entry("empty model should be accepted with arm64", "", 0),
		)

		It("should reject setting sound device", func() {
			vmi.Spec.Domain.Devices.Sound = &v1.SoundDevice{
				Name:  "test-audio-device",
				Model: "ich9",
			}
			causes := archValidator.ValidateVirtualMachineInstanceArchSetting(k8sfield.NewPath("fake"), &vmi.Spec, config)
			Expect(causes).To(HaveLen(1))
			Expect(causes[0].Field).To(Equal("fake.domain.devices.sound"))
			Expect(causes[0].Message).To(Equal("Arm64 not support sound device"))
		})
	})

})
