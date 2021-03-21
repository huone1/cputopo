/*


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
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	nodeinfov1alpha1 "github.com/huone1/cputopo/pkg/apis/nodeinfo/v1alpha1"
	versioned "github.com/huone1/cputopo/pkg/client/clientset/versioned"
	internalinterfaces "github.com/huone1/cputopo/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/huone1/cputopo/pkg/client/listers/nodeinfo/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NumatopoInformer provides access to a shared informer and lister for
// Numatopos.
type NumatopoInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.NumatopoLister
}

type numatopoInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNumatopoInformer constructs a new informer for Numatopo type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNumatopoInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNumatopoInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNumatopoInformer constructs a new informer for Numatopo type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNumatopoInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NodeinfoV1alpha1().Numatopos(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NodeinfoV1alpha1().Numatopos(namespace).Watch(context.TODO(), options)
			},
		},
		&nodeinfov1alpha1.Numatopo{},
		resyncPeriod,
		indexers,
	)
}

func (f *numatopoInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNumatopoInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *numatopoInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&nodeinfov1alpha1.Numatopo{}, f.defaultInformer)
}

func (f *numatopoInformer) Lister() v1alpha1.NumatopoLister {
	return v1alpha1.NewNumatopoLister(f.Informer().GetIndexer())
}
