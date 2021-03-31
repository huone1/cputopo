package numatopo

import (
	"github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"
	"github.com/huone1/cputopo/pkg/args"
)

var numaMap = map[string]NumaInfo{}

func RegisterNumaType(info NumaInfo) {
	numaMap[info.Name()] = info
}

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

func GetAllResTopoInfo() map[string]v1alpha1.ResourceInfo {
	numaResMap := make(map[string]v1alpha1.ResourceInfo)

	for str, info := range numaMap {
		numaResMap[str] = info.GetResourceInfoMap()
	}

	return numaResMap
}

func GetCpusDetail() map[string]v1alpha1.CPUInfo{
	for _, info := range numaMap {
		cpuDetail := info.GetCpuDetail()
		if cpuDetail == nil {
			continue
		}

		return cpuDetail
	}

	return nil
}

func init() {
	RegisterNumaType(NewCpuNumaInfo())
}
