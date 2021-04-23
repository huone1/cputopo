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
	"github.com/huone1/cputopo/pkg/args"
	"volcano.sh/apis/pkg/apis/nodeinfo/v1alpha1"
)

var numaMap = map[string]NumaInfo{}

// RegisterNumaType is the funtion to register the info provider
func RegisterNumaType(info NumaInfo) {
	numaMap[info.Name()] = info
}

// TopoInfoUpdate get the latest node topology information
// if info is changed , return true
func TopoInfoUpdate(opt *args.Argument) bool {
	isChg := false

	for str, info := range numaMap {
		ret := info.Update(opt)
		if ret == nil {
			continue
		}

		numaMap[str] = ret
		isChg = true
	}

	return isChg
}

// GetAllResAllocatableInfo returns the latest info abaut the allocatable nums of all resource
func GetAllResAllocatableInfo() map[string]v1alpha1.ResourceInfo {
	numaResMap := make(map[string]v1alpha1.ResourceInfo)

	for str, info := range numaMap {
		numaResMap[str] = info.GetResourceInfoMap()
	}

	return numaResMap
}

// GetCpusDetail returns the cpu capability topology info
func GetCpusDetail() map[string]v1alpha1.CPUInfo {
	for _, info := range numaMap {
		cpuDetail := info.GetCPUDetail()
		if cpuDetail == nil {
			continue
		}

		return cpuDetail
	}

	return nil
}

func init() {
	RegisterNumaType(NewCPUNumaInfo())
}
