package controller

import (
	"log/slog"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	toolbox "github.com/Nivesh00/automatic-pod-terminator/src/toolbox"
)


type PodTerminatorController struct {
	Workqueue workqueue.TypedRateLimitingInterface[string]
	Informer  cache.SharedIndexInformer
	Indexer   cache.Indexer
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
func (c *PodTerminatorController) AddToWorkqueue(obj interface{}, operation string) {
	// Unstructured lets us handle object as maps
	podTerminator := obj.(*unstructured.Unstructured).DeepCopy()

	toolbox.Logger.Debug(
		"processing pod terminator object for workqueue",
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
	switch operation {
	case "CREATE", "UPDATE":
		key, err = cache.MetaNamespaceKeyFunc(obj)
	case "DELETE":
		key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	}

	if err != nil {
		toolbox.Logger.Error(
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

	toolbox.Logger.Debug("adding key to workqueue", "key", key)

	c.Workqueue.AddRateLimited(key)
}