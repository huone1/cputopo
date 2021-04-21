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
	"context"
	"os"

	"github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"
	"github.com/huone1/cputopo/pkg/args"
	"github.com/huone1/cputopo/pkg/client/clientset/versioned"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// NodeInfoRefresh check the data changes
func NodeInfoRefresh(opt *args.Argument) bool {
	isChange := false

	if GetkubeletConfig(opt.KubeletConf, opt.ResReserved) {
		isChange = true
	}

	if TopoInfoUpdate(opt) {
		isChange = true
	}

	return isChange
}

// CreateOrUpdateNumatopo create or update the numatopo to etcd
func CreateOrUpdateNumatopo(client *versioned.Clientset) {
	hostname := os.Getenv("MY_NODE_NAME")
	if hostname == "" {
		klog.Errorf("Get env MY_NODE_NAME failed.")
		return
	}

	numaInfo, err := client.NodeinfoV1alpha1().Numatopos("default").Get(context.TODO(), hostname, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Errorf("Get Numatopo for node %s failed, err=%v", hostname, err)
			return
		}

		numaInfo = &v1alpha1.Numatopo{
			ObjectMeta: metav1.ObjectMeta{
				Name: hostname,
			},
			Spec: v1alpha1.NumatopoSpec{
				Policies:    GetPolicy(),
				ResReserved: GetResReserved(),
				NumaResMap:  GetAllResAllocatableInfo(),
				CpuDetail:   GetCpusDetail(),
			},
		}

		_, err = client.NodeinfoV1alpha1().Numatopos("default").Create(context.TODO(), numaInfo, metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("Create Numatopo for node %s failed, err=%v", hostname, err)
		}
	} else {
		numaInfo.Spec = v1alpha1.NumatopoSpec{
			Policies:    GetPolicy(),
			ResReserved: GetResReserved(),
			NumaResMap:  GetAllResAllocatableInfo(),
			CpuDetail:   GetCpusDetail(),
		}
		_, err = client.NodeinfoV1alpha1().Numatopos("default").Update(context.TODO(), numaInfo, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Update Numatopo for node %s failed, err=%v", hostname, err)
		}
	}
}
