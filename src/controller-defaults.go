package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

const (
	GroupName	 = "k8s.niv-ram.dev"
	GroupVersion = "v1"
)

// DeepCopy methods must be present as Kubernetes uses them extensively (e.g. caching)
// Methods are required to implement runtime.Object

// Deep copy
func (in *AutoPodTerminator) DeepCopyInto(out *AutoPodTerminator) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = AutoPodTerminatorSpec{
		LabelSelector: in.Spec.LabelSelector,
		LivePeriod: in.Spec.LivePeriod,
 	}
}

// Deep copy object
func (in *AutoPodTerminator) DeepCopyObject() runtime.Object {
	out := AutoPodTerminator{}
	in.DeepCopyInto(&out)

	return &out
}

// Deep copy list
func (in *AutoPodTerminatorList) DeepCopyObject() runtime.Object {
	out := AutoPodTerminatorList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]AutoPodTerminator, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}

// Register CR types with the scheme. Controller runtime needs
// scheme to encode and decode CR types properly

// Register all types
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemaGroupVersion,
		&AutoPodTerminator{},
		&AutoPodTerminatorList{},
	)

	metav1.AddToGroupVersion(scheme, SchemaGroupVersion)
	return nil
}

// Create schema for Auto Pod Terminator
var SchemaGroupVersion = schema.GroupVersion{
	Group: 	 GroupName,
	Version: GroupVersion,
}