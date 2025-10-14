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
 */

package main

import (
	"fmt"

	"github.com/rhobs/operator-observability-toolkit/pkg/operatorrules"
	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/kubevirt/pkg/monitoring/rules"
	"kubevirt.io/kubevirt/pkg/util/featuregate"
	operatorutil "kubevirt.io/kubevirt/pkg/virt-operator/util"
)

func main() {
	fmt.Println("KubeVirt Configurable Hypervisor Prometheus Rules Demo")
	fmt.Println("=====================================================")

	// Simulate different KubeVirt configurations
	configs := []struct {
		name        string
		kv          *v1.KubeVirt
		description string
	}{
		{
			name:        "Default (KVM)",
			description: "Default configuration without ConfigurableHypervisor feature gate",
			kv: &v1.KubeVirt{
				Spec: v1.KubeVirtSpec{
					Configuration: v1.KubeVirtConfiguration{
						DeveloperConfiguration: &v1.DeveloperConfiguration{
							FeatureGates: []string{},
						},
					},
				},
			},
		},
		{
			name:        "KVM with Feature Gate",
			description: "ConfigurableHypervisor enabled with explicit KVM configuration",
			kv: &v1.KubeVirt{
				Spec: v1.KubeVirtSpec{
					Configuration: v1.KubeVirtConfiguration{
						DeveloperConfiguration: &v1.DeveloperConfiguration{
							FeatureGates: []string{featuregate.ConfigurableHypervisor},
						},
						HypervisorConfiguration: &v1.HypervisorConfiguration{
							Name: v1.KvmHypervisorName,
						},
					},
				},
			},
		},
		{
			name:        "Hyper-V Layered",
			description: "ConfigurableHypervisor enabled with Hyper-V configuration",
			kv: &v1.KubeVirt{
				Spec: v1.KubeVirtSpec{
					Configuration: v1.KubeVirtConfiguration{
						DeveloperConfiguration: &v1.DeveloperConfiguration{
							FeatureGates: []string{featuregate.ConfigurableHypervisor},
						},
						HypervisorConfiguration: &v1.HypervisorConfiguration{
							Name: v1.HyperVLayeredHypervisorName,
						},
					},
				},
			},
		},
	}

	envManager := &operatorutil.EnvVarManagerImpl{}

	for _, config := range configs {
		fmt.Printf("\n--- %s ---\n", config.name)
		fmt.Printf("Description: %s\n", config.description)

		// Get the deployment config
		deploymentConfig := operatorutil.GetTargetConfigFromKVWithEnvVarManager(config.kv, envManager)
		hypervisorName := deploymentConfig.GetHypervisorName()
		fmt.Printf("Detected hypervisor: %s\n", hypervisorName)

		// Setup rules with the detected hypervisor
		err := rules.SetupRulesWithHypervisor("demo-namespace", hypervisorName)
		if err != nil {
			fmt.Printf("Error setting up rules: %v\n", err)
			continue
		}

		// Get the prometheus rules
		prometheusRule, err := rules.BuildPrometheusRule("demo-namespace")
		if err != nil {
			fmt.Printf("Error building prometheus rule: %v\n", err)
			continue
		}

		// Find and display the relevant metric
		fmt.Println("Generated hypervisor-specific metrics:")
		for _, group := range prometheusRule.Spec.Groups {
			for _, rule := range group.Rules {
				if rule.Record != "" && 
				   (rule.Record == "kubevirt_nodes_with_kvm" || 
				    rule.Record == "kubevirt_nodes_with_hyperv") {
					fmt.Printf("  - %s: %s\n", rule.Record, rule.Expr.StrVal)
				}
			}
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("This demo shows how the Prometheus rules are now dynamically generated")
	fmt.Println("based on the hypervisor configuration in the KubeVirt CR.")
	fmt.Println("- When ConfigurableHypervisor feature gate is disabled: uses KVM metrics")
	fmt.Println("- When enabled with KVM hypervisor: uses KVM metrics")
	fmt.Println("- When enabled with Hyper-V hypervisor: uses Hyper-V metrics")
}