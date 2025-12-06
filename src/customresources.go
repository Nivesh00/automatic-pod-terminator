package main

import (

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" // default k8s methods and attrs
	"k8s.io/apimachinery/pkg/runtime"
)

// ##############################################
// Kubernetes custom resource definitions objects
// CRDs are defined in this project
// ##############################################

// Custom resource list of PodTerminator
type PodTerminatorList struct {
	metav1.TypeMeta		  `json:",inline"`
	metav1.ListMeta 	  `json:"metadata,omitempty"`
	Items []PodTerminator `json:"items"`
}

// Custom resource PodTerminator
type PodTerminator struct {
	metav1.TypeMeta   	   `json:",inline"`
	metav1.ObjectMeta 	   `json:"metadata,omitempty"`
	Spec PodTerminatorSpec `json:"spec"`
}

// Custom resource spec PodTerminator
type PodTerminatorSpec struct {
	metav1.LabelSelector
	LivePeriod string `json:"liveperiod,omitempty"`
}

// ##############################################
// DeepCopy methods must be present as Kubernetes
// uses them extensively (e.g. caching)
// Methods are required to implement runtime.Object
// ##############################################

// Deep copy
func (in *PodTerminator) DeepCopyInto(out *PodTerminator) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = PodTerminatorSpec{
		LabelSelector: in.Spec.LabelSelector,
		LivePeriod: in.Spec.LivePeriod,
 	}
}

// Deep copy object
func (in *PodTerminator) DeepCopyObject() runtime.Object {
	out := PodTerminator{}
	in.DeepCopyInto(&out)

	return &out
}

// Deep copy list
func (in *PodTerminatorList) DeepCopyObject() runtime.Object {
	out := PodTerminatorList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]PodTerminator, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}

// // ##############################################
// // Controller logic
// // ##############################################

// func worker(q workqueue.TypedRateLimitingInterface[string], indexer cache.Indexer) {
//   for {
//     key, shutdown := q.Get()
//     if shutdown {
//       return
//     }
//     defer q.Done(key)

//     obj, exists, err := indexer.GetByKey(key)
//     if err != nil {
//       q.AddRateLimited(key)
//       continue
//     }
//     if !exists {
//       q.Forget(key)
//       continue
//     }

//     podTerminator := obj.(*PodTerminator)
//     q.Forget(key)
//   }
// }