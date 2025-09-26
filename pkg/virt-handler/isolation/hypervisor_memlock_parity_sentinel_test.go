package isolation

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Sentinel rationale:
// AdjustQemuProcessMemoryLimits currently has no hypervisor-specific branching (e.g. it does not
// consult HyperVLayeredEnabled). If future changes introduce gating like:
//
//	if cfg.HyperVLayeredEnabled() { ... }
//
// then OFF/ON branch coverage + memlock expectation tests must be added. This sentinel will fail
// the moment a call to HyperVLayeredEnabled appears in detector.go, forcing that update.
var _ = Describe("AdjustQemuProcessMemoryLimits hypervisor parity sentinel", func() {
	It("should fail fast once AdjustQemuProcessMemoryLimits introduces HyperVLayered gating", func() {
		data, err := os.ReadFile("detector.go")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("HyperVLayeredEnabled("),
			"AdjustQemuProcessMemoryLimits started using HyperVLayered gating; add OFF/ON gate tests and update memlock invariants")
	})
})
