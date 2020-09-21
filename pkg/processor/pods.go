package processor

import (
	"context"
	dnsw "github.com/jojimt/dnswatch/pkg/crd/apis/dnswatch/v1alpha"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/controller"
	"reflect"
)

func (p *Processor) initPodWatch() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.kubeClient.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.kubeClient.CoreV1().Pods(metav1.NamespaceAll).Watch(context.TODO(), options)
		},
	}

	p.podInformer = cache.NewSharedIndexInformer(
		lw,
		&v1.Pod{},
		controller.NoResyncPeriodFunc(),
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	p.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			p.podUpdated(obj.(*v1.Pod))
		},
		UpdateFunc: func(_ interface{}, obj interface{}) {
			p.podUpdated(obj.(*v1.Pod))
		},
		DeleteFunc: func(obj interface{}) {
			p.podDeleted(obj.(*v1.Pod))
		},
	})
}

func (p *Processor) podUpdated(pod *v1.Pod) {
	p.Lock()
	defer p.Unlock()

	if pod.Status.PodIP == "" {
		return
	}

	cvMeta := getCVMeta(&pod.ObjectMeta)
	p.ipToApp[pod.Status.PodIP] = cvMeta
}

func getCVMeta(m *metav1.ObjectMeta) *dnsw.ClientViewMeta {
	res := &dnsw.ClientViewMeta{
		Namespace: m.Namespace,
	}
	for _, owner := range m.OwnerReferences {
		if *owner.Controller {
			res.Name = owner.Name
			res.Kind = owner.Kind
			return res
		}
	}

	res.Name = m.Name
	res.Kind = "Pod"
	return res
}

func (p *Processor) podDeleted(pod *v1.Pod) {
	p.Lock()
	defer p.Unlock()

	if pod.Status.PodIP == "" {
		return
	}

	oldMeta := p.ipToApp[pod.Status.PodIP]
	newMeta := getCVMeta(&pod.ObjectMeta)
	if reflect.DeepEqual(oldMeta, newMeta) {
		delete(p.ipToApp, pod.Status.PodIP)
	}
}
