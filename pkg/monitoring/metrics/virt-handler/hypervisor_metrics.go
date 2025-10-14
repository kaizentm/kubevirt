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
	"encoding/xml"
	"strings"

	"github.com/rhobs/operator-observability-toolkit/pkg/operatormetrics"
	v1 "kubevirt.io/api/core/v1"
)

// HypervisorType represents the detected virtualization backend
type HypervisorType string

const (
	HypervisorTypeKVM     HypervisorType = "kvm"      // Hardware acceleration
	HypervisorTypeQEMUTCG HypervisorType = "qemu-tcg" // Software emulation
	HypervisorTypeUnknown HypervisorType = "unknown"  // Cannot determine
)

var (
	hypervisorMetrics = []operatormetrics.Metric{
		hypervisorInfoMetric,
	}

	hypervisorInfoMetric = operatormetrics.NewGaugeVec(
		operatormetrics.MetricOpts{
			Name: "kubevirt_vmi_hypervisor_info",
			Help: "Information about the hypervisor type used by a VirtualMachineInstance",
		},
		[]string{"namespace", "name", "node", "hypervisor_type"},
	)
)

// domainXMLStruct represents minimal structure needed to parse domain type
type domainXMLStruct struct {
	XMLName xml.Name `xml:"domain"`
	Type    string   `xml:"type,attr"`
}

// detectHypervisorType parses libvirt domain XML to determine hypervisor type
func detectHypervisorType(domainXML string) HypervisorType {
	if domainXML == "" {
		return HypervisorTypeUnknown
	}

	var domain domainXMLStruct
	if err := xml.Unmarshal([]byte(domainXML), &domain); err != nil {
		return HypervisorTypeUnknown
	}

	switch strings.ToLower(domain.Type) {
	case "kvm":
		return HypervisorTypeKVM
	case "qemu":
		return HypervisorTypeQEMUTCG
	default:
		return HypervisorTypeUnknown
	}
}

// updateHypervisorMetric creates or updates the hypervisor metric for a VMI
func updateHypervisorMetric(vmi *v1.VirtualMachineInstance, hypervisorType HypervisorType) {
	if vmi == nil {
		return
	}

	node := vmi.Status.NodeName
	if node == "" {
		// VMI not yet scheduled to a node
		return
	}

	hypervisorInfoMetric.WithLabelValues(
		vmi.Namespace,
		vmi.Name,
		node,
		string(hypervisorType),
	).Set(1)
}

// removeHypervisorMetric removes the hypervisor metric for a VMI
func removeHypervisorMetric(vmi *v1.VirtualMachineInstance) {
	if vmi == nil {
		return
	}

	node := vmi.Status.NodeName
	if node == "" {
		return
	}

	// Use DeletePartialMatch to remove all metrics for this specific VMI
	// This will match and delete metrics with namespace, name, and node labels
	// regardless of the hypervisor_type value
	hypervisorInfoMetric.DeletePartialMatch(map[string]string{
		"namespace": vmi.Namespace,
		"name":      vmi.Name,
		"node":      node,
	})
}
