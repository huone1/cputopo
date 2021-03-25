package args

import (
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

const (
	defaultCheckInterval = 3 * time.Second
)

// ClientOptions used to build kube rest config.
type ClientOptions struct {
	Master     string
	KubeConfig string
}

type Argument struct {
	CheckInterval     time.Duration
	KubeletConf       string
	DevicePath          string
	CpuMngstate       string
	KubeClientOptions ClientOptions
}

func NewArgument() *Argument {
	return &Argument{}
}

func (args *Argument) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&args.CheckInterval, "check-period", defaultCheckInterval, "Burst to use while talking with kubernetes apiserver")
	fs.StringVar(&args.KubeletConf, "kubelet-conf", args.KubeletConf, "Path to kubelet configure file")
	fs.StringVar(&args.NumaPath, "device-path", args.NumaPath, "Path to numa node information")
	fs.StringVar(&args.CpuMngstate, "cpu-manager-state", args.CpuMngstate, "Path to cpu_manager_state")

	fs.StringVar(&args.KubeClientOptions.Master, "master", args.KubeClientOptions.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&args.KubeClientOptions.KubeConfig, "kubeconfig", args.KubeClientOptions.KubeConfig, "Path to kubeconfig file with authorization and master location information.")
}

// BuildConfig builds kube rest config with the given options.
func BuildConfig(opt ClientOptions) (*rest.Config, error) {
	var cfg *rest.Config
	var err error

	master := opt.Master
	kubeconfig := opt.KubeConfig
	cfg, err = clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
