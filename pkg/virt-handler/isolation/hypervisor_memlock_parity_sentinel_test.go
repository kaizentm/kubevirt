package isolation

import (
	"embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Sentinel rationale:
// AdjustQemuProcessMemoryLimits currently has no hypervisor-specific branching (e.g. it does not
// consult ConfigurableHypervisorEnabled). If future changes introduce gating like:
//
//	if cfg.ConfigurableHypervisorEnabled() { ... }
//
// then OFF/ON branch coverage + memlock expectation tests must be added. This sentinel will fail
// the moment a call to ConfigurableHypervisorEnabled appears in detector.go, forcing that update.
//
//go:embed detector.go
var detectorSource string

// Prevent goimports from pruning the embed import before the compiler processes the directive.
var _ embed.FS

var _ = Describe("AdjustQemuProcessMemoryLimits hypervisor parity sentinel", func() {
	It("should fail fast once AdjustQemuProcessMemoryLimits introduces ConfigurableHypervisor gating", func() {
		Expect(detectorSource).NotTo(ContainSubstring("ConfigurableHypervisorEnabled("),
			"AdjustQemuProcessMemoryLimits started using ConfigurableHypervisor gating; add OFF/ON gate tests and update memlock invariants")
	})
})

// (detectorSource embedded above)
