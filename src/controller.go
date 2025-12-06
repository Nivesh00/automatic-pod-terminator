package main

import (
	"context"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" // default k8s methods and attrs
	corev1 "k8s.io/api/core/v1"
    ctrl "sigs.k8s.io/controller-runtime"  
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/client-go/util/workqueue"

)

type PodTerminatorController struct {
	client.Client
}

// ##############################################
// Reconciler
// ##############################################

func (c *PodTerminatorController) Reconcile (ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// List pods
	podList := corev1.PodList{}
	if err := c.List(ctx, &podList); err != nil {
		Logger.Info("could not list pods")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// ##############################################
// Event handlers
// ##############################################

func AddFunction(obj interface{}, queue *workqueue.TypedRateLimitingInterface[string]) {
	autoPodTerminator := obj.(*PodTerminator)
	(*queue).AddRateLimited(autoPodTerminator.GetNamespace() + "/" + autoPodTerminator.GetName())
}