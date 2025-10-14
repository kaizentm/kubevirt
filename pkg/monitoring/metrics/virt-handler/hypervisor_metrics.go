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
	"fmt"

	"github.com/rhobs/operator-observability-toolkit/pkg/operatormetrics"
	"k8s.io/client-go/tools/cache"
	v1 "kubevirt.io/api/core/v1"

	cmdclient "kubevirt.io/kubevirt/pkg/virt-handler/cmd-client"
)

// HypervisorType represents the detected virtualization backend
type HypervisorType string

const (
	HypervisorTypeKVM     HypervisorType = "kvm"     // KVM Hardware acceleration
	HypervisorTypeHyperv  HypervisorType = "hyperv"  // MSHV Hardware acceleration
	HypervisorTypeQEMU    HypervisorType = "qemu"    // Software emulation
	HypervisorTypeUnknown HypervisorType = "unknown" // Cannot determine
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

// updateHypervisorMetric creates or updates the hypervisor metric for a VMI
func updateHypervisorMetric(vmi *v1.VirtualMachineInstance, hypervisorType string) {
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

// handleVMIAdd handles VMI creation and running events
func handleVMIAdd(obj interface{}) {
	vmi, ok := obj.(*v1.VirtualMachineInstance)
	if !ok {
		return
	}

	// Only track VMIs that are running on this node
	if vmi.Status.Phase != v1.Running {
		return
	}

	// TODO: In a complete implementation, this would query the domain manager
	// to get the actual libvirt domain XML for the VMI
	// For now, we'll use a placeholder detection
	hypervisorType, err := getHypervisorTypeForVMI(vmi)
	if err != nil {
		// Log the error but do not set any metric
		// This may happen if the VMI is not fully started yet
		// or if there is a transient issue communicating with libvirt
		fmt.Printf("Error getting domain XML for VMI %s/%s: %v\n", vmi.Namespace, vmi.Name, err)
		return
	}

	updateHypervisorMetric(vmi, hypervisorType)
}

// handleVMIUpdate handles VMI state changes
func handleVMIUpdate(oldObj, newObj interface{}) {
	oldVMI, oldOk := oldObj.(*v1.VirtualMachineInstance)
	newVMI, newOk := newObj.(*v1.VirtualMachineInstance)
	if !oldOk || !newOk {
		return
	}

	// Handle phase transitions
	oldPhase := oldVMI.Status.Phase
	newPhase := newVMI.Status.Phase

	// If VMI just started running, add metric
	if oldPhase != v1.Running && newPhase == v1.Running {
		handleVMIAdd(newVMI)
		return
	}

	// If VMI stopped running, remove metric
	if oldPhase == v1.Running && newPhase != v1.Running {
		removeHypervisorMetric(oldVMI)
		return
	}

	// For running VMIs, we could check if domain configuration changed
	// but for this static metrics implementation, we don't need to
	// re-detect hypervisor type once set unless the VMI restarts
}

// getHypervisorTypeForVMI gets the libvirt domain XML for a VMI
// TODO: In a complete implementation, this would integrate with the domain manager
// to retrieve the actual domain XML from libvirt
func getHypervisorTypeForVMI(vmi *v1.VirtualMachineInstance) (string, error) {
	if vmi == nil {
		return "", fmt.Errorf("VMI cannot be nil")
	}

	socketPath, err := cmdclient.FindSocket(vmi)
	if err != nil {
		// nothing to scrape...
		// this means there's no socket or the socket
		// is currently unreachable for this vmi.
		return "", fmt.Errorf("unable to find socket: %w", err)
	}

	cli, err := cmdclient.NewClient(socketPath)
	if err != nil {
		// Ignore failure to connect to client.
		// These are all local connections via unix socket.
		// A failure to connect means there's nothing on the other
		// end listening.
		return "", fmt.Errorf("failed to connect to cmd client socket: %w", err)
	}
	defer cli.Close()

	domain, exists, err := cli.GetDomain()
	if err != nil {
		return "", fmt.Errorf("failed to get domain XML: %w", err)
	} else if !exists {
		return "", fmt.Errorf("domain does not exist for VMI %s/%s", vmi.Namespace, vmi.Name)
	} else {
		return domain.Spec.Type, nil
	}
}

// handleVMIDelete handles VMI cleanup
func handleVMIDelete(obj interface{}) {
	vmi, ok := obj.(*v1.VirtualMachineInstance)
	if !ok {
		// Handle DeletedFinalStateUnknown
		if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
			vmi, ok = tombstone.Obj.(*v1.VirtualMachineInstance)
			if !ok {
				return
			}
		} else {
			return
		}
	}

	removeHypervisorMetric(vmi)
}

// SetupHypervisorMetrics registers VMI informer event handlers for hypervisor metrics
func SetupHypervisorMetrics(vmiInformer cache.SharedIndexInformer) error {
	_, err := vmiInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    handleVMIAdd,
		UpdateFunc: handleVMIUpdate,
		DeleteFunc: handleVMIDelete,
	})
	return err
}
