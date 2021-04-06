package numatopo

import (
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"reflect"

	"github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"

	"sigs.k8s.io/yaml"

	"k8s.io/klog"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

type kubeletConfig struct {
	topoPolicy  map[v1alpha1.PolicyName]string
	resReserved map[string]string
}

var config = &kubeletConfig{
	topoPolicy:  make(map[v1alpha1.PolicyName]string),
	resReserved: make(map[string]string),
}

func GetPolicy() map[v1alpha1.PolicyName]string {
	return config.topoPolicy
}

func GetResReserved() map[string]string {
	return config.resReserved
}

func GetKubeletConfigFromLocalFile(kubeletConfigPath string) (*kubeletconfigv1beta1.KubeletConfiguration, error) {
	kubeletBytes, err := ioutil.ReadFile(kubeletConfigPath)
	if err != nil {
		return nil, err
	}

	kubeletConfig := &kubeletconfigv1beta1.KubeletConfiguration{}
	if err := yaml.Unmarshal(kubeletBytes, kubeletConfig); err != nil {
		return nil, err
	}
	return kubeletConfig, nil
}

func GetkubeletConfig(confPath string, resReserved map[string]string) bool {
	klConfig, err := GetKubeletConfigFromLocalFile(confPath)
	if err != nil {
		klog.Errorf("get topology Manager Policy failed, err: %v", err)
		return false
	}

	var isChange bool = false
	policy := make(map[v1alpha1.PolicyName]string)
	policy[v1alpha1.CPUManagerPolicy] = klConfig.CPUManagerPolicy
	policy[v1alpha1.TopologyManagerPolicy] = klConfig.TopologyManagerPolicy

	if !reflect.DeepEqual(config.topoPolicy, policy) {
		for key := range config.topoPolicy {
			config.topoPolicy[key] = policy[key]
		}

		isChange = true
	}

	var cpuReserved string
 	if _, ok := resReserved[string(v1.ResourceCPU)]; ok {
		cpuReserved = resReserved[string(v1.ResourceCPU)]
	} else {
		cpuReserved = klConfig.KubeReserved[string(v1.ResourceCPU)]
	}

	if config.resReserved[string(v1.ResourceCPU)] != cpuReserved {
		config.resReserved[string(v1.ResourceCPU)] = cpuReserved
		isChange = true
	}

	return isChange
}

func init() {
	config.topoPolicy[v1alpha1.CPUManagerPolicy] = "none"
	config.topoPolicy[v1alpha1.TopologyManagerPolicy] = "none"
}
