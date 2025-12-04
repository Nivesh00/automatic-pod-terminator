package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" // default k8s methods and attrs
)

// Group Version Kind
var (
	Group := 
)

type AutoPodTerminatorList struct {
	metav1.TypeMeta			  `json:",inline"`
	metav1.ListMeta 		  `json:"metadata,omitempty"`
	Items []AutoPodTerminator `json:"items"`
}

type AutoPodTerminator struct {
	metav1.TypeMeta   	   	   `json:",inline"`
	metav1.ObjectMeta 	   	   `json:"metadata,omitempty"`
	Spec AutoPodTerminatorSpec `json:"spec"`
}

type AutoPodTerminatorSpec struct {
	metav1.LabelSelector
	LivePeriod string `json:"liveperiod,omitempty"`
}