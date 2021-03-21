package numatopo

import (
	"io/ioutil"
	"reflect"
	"github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"

	"k8s.io/klog"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/yaml"
)

var topoPolicy = make(map[v1alpha1.PolicyName]string)

func GetPolicy() map[v1alpha1.PolicyName]string {
	return topoPolicy
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

func GetkubeletConfig(confPath string) map[v1alpha1.PolicyName]string {
	klConfig, err := GetKubeletConfigFromLocalFile(confPath)
	if err != nil {
		klog.Errorf("get topology Manager Policy failed, err: %v", err)
		return nil
	}

	policy := make(map[v1alpha1.PolicyName]string)

	policy[v1alpha1.CPUManagerPolicy] = klConfig.CPUManagerPolicy
	policy[v1alpha1.TopologyManagerPolicy] = klConfig.TopologyManagerPolicy

	if !reflect.DeepEqual(topoPolicy, policy) {
		for key := range topoPolicy {
			topoPolicy[key] = policy[key]
		}

		return topoPolicy
	}

	return nil
}

func init() {
	topoPolicy[v1alpha1.CPUManagerPolicy] = "none"
	topoPolicy[v1alpha1.TopologyManagerPolicy] = "none"
}
