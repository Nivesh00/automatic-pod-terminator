package main

import (
	"log/slog"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)


type PodTerminatorController struct {
	workqueue workqueue.TypedRateLimitingInterface[string]
	informer  cache.SharedIndexInformer
	indexer   cache.Indexer
}

// ##############################################
// Event handlers
// ##############################################

// Edge-driven logic: do changes after observing a change in state of the object
// Level-driven logic: do changes based on current state of object, not whether the object was changed, i.e.
// whether its state changed
// Use level-driven logic instead of edge-driven logic for Update functions

func (c *PodTerminatorController) onAdd(obj interface{}) {
	podTerminator := obj.(*PodTerminator)
	Logger.Info(
		"handler triggered for pod terminator object",
		"handler_function_name",
		"OnPodTerminatorAdd",
		"namespace",
		podTerminator.GetNamespace(),
		"name",
		podTerminator.GetName(),
	)

	// logic

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		Logger.Error(
			"cannot add pod teminator object to workqueue",
			"namespace",
			podTerminator.GetNamespace(),
			"name",
			podTerminator.GetName(),
			slog.Any("error", err),
		)
	}

	c.workqueue.AddRateLimited(key)
}

func (c *PodTerminatorController) onUpdate(obj interface{}) {
	podTerminator := obj.(*PodTerminator)
	Logger.Info(
		"handler triggered for pod terminator object",
		"handler_function_name",
		"OnPodTerminatorUpdate",
		"namespace",
		podTerminator.GetNamespace(),
		"name",
		podTerminator.GetName(),
	)

	// logic

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		Logger.Error(
			"cannot add pod teminator object to workqueue",
			"namespace",
			podTerminator.GetNamespace(),
			"name",
			podTerminator.GetName(),
			slog.Any("error", err),
		)
	}

	c.workqueue.AddRateLimited(key)
}

func (c *PodTerminatorController) onDelete(obj interface{}) {
	podTerminator := obj.(*PodTerminator)
	Logger.Info(
		"handler triggered for pod terminator object",
		"handler_function_name",
		"OnPodTerminatorDelete",
		"namespace",
		podTerminator.GetNamespace(),
		"name",
		podTerminator.GetName(),
	)

	// logic

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		Logger.Error(
			"cannot add pod teminator object to workqueue",
			"namespace",
			podTerminator.GetNamespace(),
			"name",
			podTerminator.GetName(),
			slog.Any("error", err),
		)
	}

	c.workqueue.AddRateLimited(key)
}