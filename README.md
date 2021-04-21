# cputopo

The DaemonSet **cputopo** is used to obtain the topology information about the CPU resources of the worker node, upload it to ETCD by the [CRD](https://github.com/volcano-sh/apis/tree/numatopo-api/pkg/apis/nodeinfo/v1alpha1) method for volcano scheduler [numa-aware plugin](https://github.com/volcano-sh/volcano/tree/master/pkg/scheduler/plugins/numaaware)
## Quick Start Guide

### Compile
```
   make image TAG=XXX
```

### Install
````
1. Edit the file ./installer/numa-topo.yaml
   - modify the image version 
   - set the kubelet config path
   - set the system device path 

2. Deploy cputopo
   kubectl apply -f ./installer/numa-topo.yaml
````