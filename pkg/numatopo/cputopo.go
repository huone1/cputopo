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

type CpuNumaInfo struct {
	NUMANodes   []int
	NUMA2CpuCap map[int]int
	cpu2NUMA    map[int]int

	NUMA2FreeCpus    map[int][]int
	NUMA2FreeCpusNum map[int]int
}

func NewCpuNumaInfo() *CpuNumaInfo {
	numaInfo := &CpuNumaInfo{
		NUMA2CpuCap:      make(map[int]int),
		cpu2NUMA:         make(map[int]int),
		NUMA2FreeCpus:    make(map[int][]int),
		NUMA2FreeCpusNum: make(map[int]int),
	}

	return numaInfo
}

func (info *CpuNumaInfo) Name() v1alpha1.ResourceName {
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
	newInfo := NewCpuNumaInfo()
	newInfo.NUMANodes = getNumaOnline(filepath.Join(opt.NumaPath, "online"))
	newInfo.numaCapUpdate(opt.NumaPath)
	newInfo.numaAllocUpdate(opt.CpuMngstate)

	if !reflect.DeepEqual(newInfo, info) {
		return newInfo
	}

	return nil
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
