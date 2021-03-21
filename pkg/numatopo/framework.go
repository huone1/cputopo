package numatopo

import (
	"github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"
	"github.com/huone1/cputopo/pkg/args"
)

var numaMap = map[v1alpha1.ResourceName]NumaInfo{}

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

func GetAllResTopoInfo() map[v1alpha1.ResourceName]v1alpha1.ResourceInfoMap {
	numaResMap := make(map[v1alpha1.ResourceName]v1alpha1.ResourceInfoMap)

	for str, info := range numaMap {
		numaResMap[str] = info.GetResourceInfoMap()
	}

	return numaResMap
}

func init() {
	RegisterNumaType(NewCpuNumaInfo())
}
