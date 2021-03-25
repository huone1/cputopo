package numatopo

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"

	"k8s.io/klog"
	cpustate "k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/state"

	"github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"
	"github.com/huone1/cputopo/pkg/args"
	"github.com/huone1/cputopo/pkg/util"
)

// CPUInfo contains the NUMA, socket, and core IDs associated with a CPU.
type CPUInfo struct {
	NUMANodeID int
	SocketID   int
	CoreID     int
}

type CpuNumaInfo struct {
	NUMANodes   []int
	NUMA2CpuCap map[int]int
	cpu2NUMA    map[int]int
	cpuDetail   map[int]CPUInfo

	NUMA2FreeCpus    map[int][]int
	NUMA2FreeCpusNum map[int]int
}

func NewCpuNumaInfo() *CpuNumaInfo {
	numaInfo := &CpuNumaInfo{
		NUMA2CpuCap:      make(map[int]int),
		cpu2NUMA:         make(map[int]int),
		cpuDetail:        make(map[int]CPUInfo),
		NUMA2FreeCpus:    make(map[int][]int),
		NUMA2FreeCpusNum: make(map[int]int),
	}

	return numaInfo
}

func (info *CpuNumaInfo) Name() string {
	return "cpu"
}

func getNumaOnline(onlinePath string) []int {
	data, err := ioutil.ReadFile(onlinePath)
	if err != nil {
		klog.Errorf("getNumaOnline read file failed.")
		return []int{}
	}

	nodeList, apiErr := util.Parse(string(data))
	if apiErr != nil {
		klog.Errorf("getNumaOnline parse failed.")
		return []int{}
	}

	return nodeList
}

func (info *CpuNumaInfo) cpu2numa(cpuid int) int {
	return info.cpu2NUMA[cpuid]
}

func getNumaNodeCpucap(nodePath string, nodeId int) []int {
	cpuPath := filepath.Join(nodePath, fmt.Sprintf("node%d", nodeId), "cpulist")
	data, err := ioutil.ReadFile(cpuPath)
	if err != nil {
		klog.Errorf("numa node cpulist read file failed, err: %v", err)
		return nil
	}

	cpuList, apiErr := util.Parse(string(data))
	if apiErr != nil {
		klog.Errorf("numa node cpulist parse failed, err: %v", err)
		return nil
	}

	return cpuList
}

func getFreeCpulist(cpuMngstate string) []int {
	data, err := ioutil.ReadFile(cpuMngstate)
	if err != nil {
		klog.Errorf("cpu-mem-state read failed, err: %v", err)
		return nil
	}

	checkpoint := cpustate.NewCPUManagerCheckpoint()
	checkpoint.UnmarshalCheckpoint(data)

	cpuList, apiErr := util.Parse(checkpoint.DefaultCPUSet)
	if apiErr != nil {
		klog.Errorf("cpu-mem-state parse failed, err: %v", err)
		return nil
	}

	return cpuList
}

func (info *CpuNumaInfo) numaCapUpdate(numaPath string) {
	for _, node := range info.NUMANodes {
		cpuList := getNumaNodeCpucap(numaPath, node)
		info.NUMA2CpuCap[node] = len(cpuList)

		for _, cpu := range cpuList {
			info.cpu2NUMA[cpu] = node
		}
	}

	return
}

func (info *CpuNumaInfo) numaAllocUpdate(cpuMngstate string) {
	freeCpuList := getFreeCpulist(cpuMngstate)
	for _, cpuid := range freeCpuList {
		numaId := info.cpu2numa(cpuid)
		info.NUMA2FreeCpus[numaId] = append(info.NUMA2FreeCpus[numaId], cpuid)
	}

	for numaId, cpus := range info.NUMA2FreeCpus {
		info.NUMA2FreeCpusNum[numaId] = len(cpus)
	}
}

func (info *CpuNumaInfo) Update(opt *args.Argument) NumaInfo {
	cpuNumaBasePath := filepath.Join(opt.DevicePath, "node")
	newInfo := NewCpuNumaInfo()
	newInfo.NUMANodes = getNumaOnline(filepath.Join(cpuNumaBasePath, "online"))
	newInfo.numaCapUpdate(cpuNumaBasePath)
	newInfo.numaAllocUpdate(opt.CpuMngstate)
	newInfo.cpuDetail = newInfo.getAllCpuTopoInfo(opt.DevicePath)
	if !reflect.DeepEqual(newInfo, info) {
		return newInfo
	}

	return nil
}

func (info *CpuNumaInfo) getAllCpuTopoInfo(devicePath string) map[int]CPUInfo {
	cpuTopoInfo := make(map[int]CPUInfo)
	for cpuId, numaId := range info.cpu2NUMA {
		coreId, socketId, err := getCoreIdScoketIdForcpu(devicePath, cpuId)
		if err != nil {
			return nil
		}

		cpuTopoInfo[cpuId] = CPUInfo{
			NUMANodeID: numaId,
			CoreID:     coreId,
			SocketID:   socketId,
		}
	}

	info.cpuDetail = cpuTopoInfo
	return cpuTopoInfo
}

func getCoreIdScoketIdForcpu(devicePath string, cpuId int) (coreId, socketId int, err error) {
	topoPath := filepath.Join(devicePath, fmt.Sprintf("cpu/cpu%d", cpuId), "topology")
	corePath := filepath.Join(topoPath, "core_id")
	data, err := ioutil.ReadFile(corePath)
	if err != nil {
		return 0, 0, fmt.Errorf("cpu %d read core_id file failed", cpuId)
	}

	tmpData, apiErr := util.Parse(string(data))
	if apiErr != nil {
		return 0, 0, fmt.Errorf("cpu %d core_id parse failed", cpuId)
	}

	coreId = tmpData[0]

	socketPath := filepath.Join(topoPath, "physical_package_id")
	data, err = ioutil.ReadFile(socketPath)
	if err != nil {
		return 0, 0, fmt.Errorf("cpu %d read scoket_id file failed", cpuId)
	}

	tmpData, apiErr = util.Parse(string(data))
	if apiErr != nil {
		return 0, 0, fmt.Errorf("cpu %d scoket_id parse failed", cpuId)
	}

	socketId = tmpData[0]

	return coreId, socketId, nil
}

func getCoreIdForcpu(cpupath string, cpuId int) CPUInfo {

}

func (info *CpuNumaInfo) GetResourceInfoMap() v1alpha1.ResourceInfoMap {
	resMap := make(v1alpha1.ResourceInfoMap)
	for _, numaId := range info.NUMANodes {
		resMap[strconv.Itoa(numaId)] = v1alpha1.ResourceInfo{
			Allocatable: info.NUMA2FreeCpusNum[numaId],
			Capacity:    info.NUMA2CpuCap[numaId],
		}
	}

	return resMap
}
