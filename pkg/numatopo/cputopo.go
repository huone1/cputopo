/*
Copyright 2021 The Volcano Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package numatopo

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"
	"github.com/huone1/cputopo/pkg/args"
	"github.com/huone1/cputopo/pkg/util"

	"k8s.io/klog"
	cpustate "k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/state"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

type CpuNumaInfo struct {
	NUMANodes   []int
	NUMA2CpuCap map[int]int
	cpu2NUMA    map[int]int
	cpuDetail   map[int]v1alpha1.CPUInfo

	NUMA2FreeCpus map[int][]int
}

func NewCpuNumaInfo() *CpuNumaInfo {
	numaInfo := &CpuNumaInfo{
		NUMA2CpuCap:   make(map[int]int),
		cpu2NUMA:      make(map[int]int),
		cpuDetail:     make(map[int]v1alpha1.CPUInfo),
		NUMA2FreeCpus: make(map[int][]int),
	}

	return numaInfo
}

func (info *CpuNumaInfo) Name() string {
	return "cpu"
}

func getNumaOnline(onlinePath string) []int {
	data, err := ioutil.ReadFile(onlinePath)
	if err != nil {
		klog.Errorf("Read numa online file failed, err=%v.", err)
		return []int{}
	}

	nodeList, apiErr := util.Parse(string(data))
	if apiErr != nil {
		klog.Errorf("Parse numa online file failed, err=%v.", apiErr)
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
		klog.Errorf("Read node%d cpulist file failed, err: %v", nodeId, err)
		return nil
	}

	cpuList, apiErr := util.Parse(string(data))
	if apiErr != nil {
		klog.Errorf("Parse node%d cpulist file failed, err: %v", nodeId, apiErr)
		return nil
	}

	return cpuList
}

func getFreeCpulist(cpuMngstate string) []int {
	data, err := ioutil.ReadFile(cpuMngstate)
	if err != nil {
		klog.Errorf("Read cpu_manager_state failed, err: %v", err)
		return nil
	}

	checkpoint := cpustate.NewCPUManagerCheckpoint()
	checkpoint.UnmarshalCheckpoint(data)

	cpuList, apiErr := util.Parse(checkpoint.DefaultCPUSet)
	if apiErr != nil {
		klog.Errorf("Parse cpu_manager_state failed, err: %v", err)
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

func (info *CpuNumaInfo) getAllCpuTopoInfo(devicePath string) map[int]v1alpha1.CPUInfo {
	cpuTopoInfo := make(map[int]v1alpha1.CPUInfo)
	for cpuId, numaId := range info.cpu2NUMA {
		coreId, socketId, err := getCoreIdScoketIdForcpu(devicePath, cpuId)
		if err != nil {
			klog.Errorf("Get cpu detail failed, err=<%v>", err)
			return nil
		}

		cpuTopoInfo[cpuId] = v1alpha1.CPUInfo{
			NUMANodeID: numaId,
			CoreID:     coreId,
			SocketID:   socketId,
		}
	}

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

func (info *CpuNumaInfo) GetResourceInfoMap() v1alpha1.ResourceInfo {
	sets := cpuset.NewCPUSet()
	var cap = 0

	for _, freeCpus := range info.NUMA2FreeCpus {
		tmp := cpuset.NewCPUSet(freeCpus...)
		sets = sets.Union(tmp)
	}

	for numaId := range info.NUMA2CpuCap {
		cap += info.NUMA2CpuCap[numaId]
	}

	return v1alpha1.ResourceInfo{
		Allocatable: sets.String(),
		Capacity:    cap,
	}
}

func (info *CpuNumaInfo) GetCpuDetail() map[string]v1alpha1.CPUInfo {
	allCpuTopoInfo := make(map[string]v1alpha1.CPUInfo)

	for cpuId, cpuInfo := range info.cpuDetail {
		allCpuTopoInfo[strconv.Itoa(cpuId)] = cpuInfo
	}

	return allCpuTopoInfo
}
