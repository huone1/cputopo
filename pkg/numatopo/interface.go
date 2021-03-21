package numatopo

import (
	"github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"
	"github.com/huone1/cputopo/pkg/args"
)

type NumaInfo interface {
	Name() v1alpha1.ResourceName
	Update(opt *args.Argument) NumaInfo
	GetResourceInfoMap() v1alpha1.ResourceInfoMap
}
