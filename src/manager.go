package main

import (
	"context"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)


type PodTerminatorController struct {
	// *dynamic.DynamicClient
	workqueue workqueue.TypedRateLimitingInterface[string]
	informer  cache.SharedIndexInformer
	indexer   cache.Indexer
}

// ##############################################
// Event handlers
// ##############################################

// Edge-driven logic: do changes after observing a change in state of the object.
// Level-driven logic: do changes based on current state of object, not whether the object was changed, i.e.
// whether its state changed.
// Use level-driven logic instead of edge-driven logic for Update functions.
// `unstructured.Unstructured` a flexible, map-like object that stores raw JSON from Kubernetes.
// Internally, it wraps a `map[string]interface{}` and behaves like raw JSON/YAML.

// Queue an event by adding it to the ratelimited workqueue
func (c *PodTerminatorController) addToWorkqueue(obj interface{}, operation string) {
	// Unstructured lets us handle object as maps
	podTerminator := obj.(*unstructured.Unstructured).DeepCopy()

	Logger.Debug(
		"adding pod terminator resource to workqueue",
		"operation",
		operation,
		"namespace",
		podTerminator.GetNamespace(),
		"name",
		podTerminator.GetName(),
	)

	// logic

	var key string
	var err error
	if operation != "delete" {
		key, err = cache.MetaNamespaceKeyFunc(obj)
	} else {
		key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	}

	if err != nil {
		Logger.Error(
			"failed adding pod terminator resource to work queue",
			"operation",
			operation,
			"namespace",
			podTerminator.GetNamespace(),
			"name",
			podTerminator.GetName(),
			slog.Any("error", err),
		)
		return
	}

	c.workqueue.AddRateLimited(key)
}

// ##############################################
// Controller logic
// ##############################################

func (c *PodTerminatorController) worker(clientset *dynamic.DynamicClient) {
  for {
    key, shutdown := c.workqueue.Get()
    if shutdown {
      	return
    }
	// Tell queue we are done working with this key
	// Only one instance of the key can exist in a work queue, so we need
	// to call Done so that key can be readded
    defer c.workqueue.Done(key)

    obj, exists, err := c.indexer.GetByKey(key)
    if err != nil {
		// If key has been requed more than 10 times, forget it
		if 	c.workqueue.NumRequeues(key) >= 10 {
			c.workqueue.Forget(key)
			continue
		}
		// Only readd key if error is recoverable
      	c.workqueue.AddRateLimited(key)
      	continue
    }
    if !exists {
      c.workqueue.Forget(key)
      continue
    }

	// Convert unstructured object to object of type PodTerminator
    podTerminatorUnstructured := obj.(*unstructured.Unstructured)
	var podTerminator PodTerminator
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		podTerminatorUnstructured.Object,
		&podTerminator,
	)
	if err != nil {
		Logger.Error(
			"controller failed to convert unstructured object to object of type podterminator, " +
			"readding resource to workqueue",
			"namespace",
			podTerminator.GetNamespace(),
			"name",
			podTerminator.GetName(),
			slog.Any("error", err),
		)
		// If key has been requed more than 10 times, forget it
		if 	c.workqueue.NumRequeues(key) >= 10 {
			c.workqueue.Forget(key)
			continue
		}
		// Only readd key if error is recoverable
      	c.workqueue.AddRateLimited(key)
      	continue
	}

	Logger.Info(
		"controller processing pod terminator resource",
		"namespace",
		podTerminator.GetNamespace(),
		"name",
		podTerminator.GetName(),
	)

	podTerminator.Annotations["processed"] = "true"
	client := clientset.Resource(schema.GroupVersionResource{
		Group: 	  "",
		Version:  "v1",
		Resource: "pods",
	})

	podListUnstructured, err := client.List(context.Background(), metav1.ListOptions{
		// LabelSelector: "app=pod-terminator",
	})
	if err != nil {
		Logger.Error(
			"controller cannot list pod resources",
			slog.Any("error", err),
		)
	}
	var podList corev1.PodList
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		podListUnstructured.Object,
		&podList,
	)
	if err != nil {
		Logger.Error(
			"controller failed to convert unstructured object to object of type podlist",
			slog.Any("error", err),
		)
	}

	Logger.Info(
		"printing podlist",
		"podlist",
		podList,
	)

	// Finished processing key, all ok
    c.workqueue.Forget(key)
  }
}
