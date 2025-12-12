package virtruntime

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"kubevirt.io/client-go/log"

	"github.com/mitchellh/go-ps"
	"golang.org/x/sys/unix"
	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/util/hardware"
	"kubevirt.io/kubevirt/pkg/virt-handler/cgroup"
	"kubevirt.io/kubevirt/pkg/virt-handler/isolation"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

type maskType bool

type cpuMask struct {
	mask map[string]maskType
}

const (
	enabled  maskType = true
	disabled maskType = false
)

var (
	// parse CPU Mask expressions
	cpuRangeRegex  = regexp.MustCompile(`^(\d+)-(\d+)$`)
	negateCPURegex = regexp.MustCompile(`^\^(\d+)$`)
	singleCPURegex = regexp.MustCompile(`^(\d+)$`)

	// parse thread comm value expression
	vcpuRegex = regexp.MustCompile(`^CPU (\d+)/KVM\n$`) // These threads follow this naming pattern as their command value (/proc/{pid}/task/{taskid}/comm)
	// QEMU uses threads to represent vCPUs.

)

type KvmVirtRuntime struct {
	podIsolationDetector isolation.PodIsolationDetector
	logger               *log.FilteredLogger
}

func (k *KvmVirtRuntime) HandleHousekeeping(vmi *v1.VirtualMachineInstance, cgroupManager cgroup.Manager, domain *api.Domain) error {
	if vmi.IsCPUDedicated() && vmi.Spec.Domain.CPU.IsolateEmulatorThread {
		err := k.configureHousekeepingCgroup(vmi, cgroupManager, domain)
		if err != nil {
			return err
		}
	}

	// Configure vcpu scheduler for realtime workloads and affine PIT thread for dedicated CPU
	if vmi.IsRealtimeEnabled() && !vmi.IsRunning() && !vmi.IsFinal() {
		k.logger.Object(vmi).Info("Configuring vcpus for real time workloads")
		if err := k.configureVCPUScheduler(vmi); err != nil {
			return err
		}
	}
	if vmi.IsCPUDedicated() && !vmi.IsRunning() && !vmi.IsFinal() {
		k.logger.V(3).Object(vmi).Info("Affining PIT thread")
		if err := k.affinePitThread(vmi); err != nil {
			return err
		}
	}
	// TODO L1VH: Move this recording to the function calling handleHousekeeping

	return nil
}

func (k *KvmVirtRuntime) AdjustQemuProcessMemoryLimits(podIsoDetector isolation.PodIsolationDetector, vmi *v1.VirtualMachineInstance, additionalOverheadRatio *string) error {
	// KVM does not require any special handling for QEMU memory limits
	return nil
}

func (k *KvmVirtRuntime) configureHousekeepingCgroup(vmi *v1.VirtualMachineInstance, cgroupManager cgroup.Manager, domain *api.Domain) error {
	if err := cgroupManager.CreateChildCgroup("housekeeping", "cpuset"); err != nil {
		k.logger.Reason(err).Error("CreateChildCgroup ")
		return err
	}

	// bail out if domain does not exist
	if domain == nil {
		return nil
	}

	if domain.Spec.CPUTune == nil || domain.Spec.CPUTune.EmulatorPin == nil {
		return nil
	}

	hkcpus, err := hardware.ParseCPUSetLine(domain.Spec.CPUTune.EmulatorPin.CPUSet, 100)
	if err != nil {
		return err
	}

	k.logger.V(3).Object(vmi).Infof("housekeeping cpu: %v", hkcpus)

	err = cgroupManager.SetCpuSet("housekeeping", hkcpus)
	if err != nil {
		return err
	}

	tids, err := cgroupManager.GetCgroupThreads()
	if err != nil {
		return err
	}
	hktids := make([]int, 0, 10)

	for _, tid := range tids {
		proc, err := ps.FindProcess(tid)
		if err != nil {
			k.logger.Object(vmi).Errorf("Failure to find process: %s", err.Error())
			return err
		}
		if proc == nil {
			return fmt.Errorf("failed to find process with tid: %d", tid)
		}
		comm := proc.Executable()
		if strings.Contains(comm, "CPU ") && strings.Contains(comm, "KVM") {
			continue
		}
		hktids = append(hktids, tid)
	}

	k.logger.V(3).Object(vmi).Infof("hk thread ids: %v", hktids)
	for _, tid := range hktids {
		err = cgroupManager.AttachTID("cpuset", "housekeeping", tid)
		if err != nil {
			k.logger.Object(vmi).Errorf("Error attaching tid %d: %v", tid, err.Error())
			return err
		}
	}

	return nil
}

// configureRealTimeVCPUs parses the realtime mask value and configured the selected vcpus
// for real time workloads by setting the scheduler to FIFO and process priority equal to 1.
func (k *KvmVirtRuntime) configureVCPUScheduler(vmi *v1.VirtualMachineInstance) error {
	res, err := k.podIsolationDetector.Detect(vmi)
	if err != nil {
		return err
	}
	qemuProcess, err := res.GetQEMUProcess()
	if err != nil {
		return err
	}
	vcpus, err := getVCPUThreadIDs(qemuProcess.Pid())
	if err != nil {
		return err
	}
	mask, err := parseCPUMask(vmi.Spec.Domain.CPU.Realtime.Mask)
	if err != nil {
		return err
	}
	for vcpuID, threadID := range vcpus {
		if mask.isEnabled(vcpuID) {
			param := schedParam{priority: 1}
			tid, err := strconv.Atoi(threadID)
			if err != nil {
				return err
			}
			err = schedSetScheduler(tid, schedFIFO, param)
			if err != nil {
				return fmt.Errorf("failed to set FIFO scheduling and priority 1 for thread %d: %w", tid, err)
			}
		}
	}
	return nil
}

func (k *KvmVirtRuntime) affinePitThread(vmi *v1.VirtualMachineInstance) error {
	res, err := k.podIsolationDetector.Detect(vmi)
	if err != nil {
		return err
	}
	var Mask unix.CPUSet
	Mask.Zero()
	qemuprocess, err := res.GetQEMUProcess()
	if err != nil {
		return err
	}
	qemupid := qemuprocess.Pid()
	if qemupid == -1 {
		return nil
	}

	pitpid, err := res.KvmPitPid()
	if err != nil {
		return err
	}
	if pitpid == -1 {
		return nil
	}
	if vmi.IsRealtimeEnabled() {
		param := schedParam{priority: 2}
		err = schedSetScheduler(pitpid, schedFIFO, param)
		if err != nil {
			return fmt.Errorf("failed to set FIFO scheduling and priority 2 for thread %d: %w", pitpid, err)
		}
	}
	vcpus, err := getVCPUThreadIDs(qemupid)
	if err != nil {
		return err
	}
	vpid, ok := vcpus["0"]
	if ok == false {
		return nil
	}
	vcpupid, err := strconv.Atoi(vpid)
	if err != nil {
		return err
	}
	err = unix.SchedGetaffinity(vcpupid, &Mask)
	if err != nil {
		return err
	}
	return unix.SchedSetaffinity(pitpid, &Mask)
}

func isVCPU(comm []byte) (string, bool) {
	if !vcpuRegex.MatchString(string(comm)) {
		return "", false
	}
	v := vcpuRegex.FindSubmatch(comm)
	return string(v[1]), true
}
func getVCPUThreadIDs(pid int) (map[string]string, error) {

	p := filepath.Join(string(os.PathSeparator), "proc", strconv.Itoa(pid), "task")
	d, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}
	ret := map[string]string{}
	for _, f := range d {
		if f.IsDir() {
			c, err := os.ReadFile(filepath.Join(p, f.Name(), "comm"))
			if err != nil {
				return nil, err
			}
			if v, ok := isVCPU(c); ok {
				ret[v] = f.Name()
			}
		}
	}
	return ret, nil
}

// parseCPUMask parses the mask and maps the results into a structure that contains which
// CPUs are enabled or disabled for the scheduling and priority changes.
// This implementation reimplements the libvirt parsing logic defined here:
// https://github.com/libvirt/libvirt/blob/56de80cb793aa7aedc45572f8b6ec3fc32c99309/src/util/virbitmap.c#L382
// except that in this case it uses a map[string]maskType instead of a bit array.
func parseCPUMask(mask string) (*cpuMask, error) {

	vcpus := cpuMask{}
	if len(mask) == 0 {
		return &vcpus, nil
	}
	vcpus.mask = make(map[string]maskType)

	masks := strings.Split(mask, ",")
	for _, i := range masks {
		m := strings.TrimSpace(i)
		switch {
		case cpuRangeRegex.MatchString(m):
			match := cpuRangeRegex.FindSubmatch([]byte(m))
			startID, err := strconv.Atoi(string(match[1]))
			if err != nil {
				return nil, err
			}
			endID, err := strconv.Atoi(string(match[2]))
			if err != nil {
				return nil, err
			}
			if startID < 0 {
				return nil, fmt.Errorf("invalid vcpu mask start index `%d`", startID)
			}
			if endID < 0 {
				return nil, fmt.Errorf("invalid vcpu mask end index `%d`", endID)
			}
			if startID > endID {
				return nil, fmt.Errorf("invalid mask range `%d-%d`", startID, endID)
			}
			for id := startID; id <= endID; id++ {
				vid := strconv.Itoa(id)
				if !vcpus.has(vid) {
					vcpus.set(vid, enabled)
				}
			}
		case singleCPURegex.MatchString(m):
			match := singleCPURegex.FindSubmatch([]byte(m))
			vid, err := strconv.Atoi(string(match[1]))
			if err != nil {
				return nil, err
			}
			if vid < 0 {
				return nil, fmt.Errorf("invalid vcpu index `%d`", vid)
			}
			if !vcpus.has(string(match[1])) {
				vcpus.set(string(match[1]), enabled)
			}
		case negateCPURegex.MatchString(m):
			match := negateCPURegex.FindSubmatch([]byte(m))
			vid, err := strconv.Atoi(string(match[1]))
			if err != nil {
				return nil, err
			}
			if vid < 0 {
				return nil, fmt.Errorf("invalid vcpu index `%d`", vid)
			}
			vcpus.set(string(match[1]), disabled)
		default:
			return nil, fmt.Errorf("invalid mask value '%s' in '%s'", i, mask)
		}
	}
	return &vcpus, nil
}

func (c cpuMask) isEnabled(vcpuID string) bool {
	if len(c.mask) == 0 {
		return true
	}
	if t, ok := c.mask[vcpuID]; ok {
		return t == enabled
	}
	return false
}

func (c *cpuMask) has(vcpuID string) bool {
	_, ok := c.mask[vcpuID]
	return ok
}

func (c *cpuMask) set(vcpuID string, mtype maskType) {
	c.mask[vcpuID] = mtype
}
