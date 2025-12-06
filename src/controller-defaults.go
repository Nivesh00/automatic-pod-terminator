package main

import (
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

const (
	// Pod GVR
	PodGroupName	= ""
	PodVersion  	= "v1"
	PodResource		= "pods"

	// Auto pod terminator GVR
	PodTerminatorGroupName	 = "k8s.niv-ram.dev"
	PodTerminatorVersion 	 = "v1"
	PodTerminatorResource	 = "autopodterminators"
)

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

// ##############################################
// Register CR types with the scheme. Controller 
// runtime needs scheme to encode and decode 
// CR types properly
// ##############################################

// Create schema for Auto Pod Terminator
var APTGroupVersionResource = schema.GroupVersionResource{
	Group: 	  PodTerminatorGroupName,
	Version:  PodTerminatorVersion,
	Resource: PodTerminatorResource,
}

// Create schema for Pod
var PodGroupVersion = schema.GroupVersion{
	Group: 	  PodGroupName,
	Version:  PodVersion,
}