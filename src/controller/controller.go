package controller

import (
	"context"
	"encoding/json"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	toolbox "github.com/Nivesh00/automatic-pod-terminator/src/toolbox"
	customresources "github.com/Nivesh00/automatic-pod-terminator/src/customresources"
)

// ##############################################
// Error handling for key in workqueue
// ##############################################

func (c *PodTerminatorController) handleError(key string) {
	retries := 10
	// If key has been requed more than 10 times, forget it
	if 	c.Workqueue.NumRequeues(key) >= retries {
		toolbox.Logger.Error(
			"cannot process item: exceeded number of retries for item",
			"key",
			key,
			"number_of_retries",
			retries,
		)
		c.Workqueue.Forget(key)
		return
	}
	// Only readd key if error is recoverable
	toolbox.Logger.Error(
		"cannot process item: requeueing item",
		"key",
		key,
	)
	c.Workqueue.AddRateLimited(key)
	c.Workqueue.Done(key)
}

// ##############################################
// Controller logic
// ##############################################

// Controller logic which is triggered on create, update and delete events
func (c *PodTerminatorController) Worker(clientset *dynamic.DynamicClient) {
  for {
    key, _ := c.Workqueue.Get()

	// Tell queue we are done working with this key
	// Only one instance of the key can exist in a work queue, so we need
	// to call Done so that key can be readded

    obj, exists, err := c.Indexer.GetByKey(key)
    if err != nil {
		c.handleError(key)
      	continue
    }
    if !exists {
	    c.Workqueue.Done(key)
      	c.Workqueue.Forget(key)
     	continue
    }

	// Convert unstructured object to object of type PodTerminator
    podTerminatorUnstructured := obj.(*unstructured.Unstructured)
	var podTerminator customresources.PodTerminator
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		podTerminatorUnstructured.UnstructuredContent(),
		&podTerminator,
	)
	if err != nil {
		c.handleError(key)
		toolbox.Logger.Error(
			"an error occured while converting unstructured object to typed object",
			"resource",
			"pod terminator",
			slog.Any("error", err),
		)
      	continue
	}

	toolbox.Logger.Info(
		"controller processing pod terminator resource",
		"namespace",
		podTerminator.GetNamespace(),
		"name",
		podTerminator.GetName(),
	)

	client := clientset.Resource(schema.GroupVersionResource{
		Group: 	  "",
		Version:  "v1",
		Resource: "pods",
	})

	podList, err := getAllPods(client)

	if err != nil {
		c.handleError(key)
		toolbox.Logger.Error(
			"an error occured while getting pod list",
			slog.Any("error", err),
		)
      	continue
	}

	err = deletePodsWithoutOwners(client, podList)
	if err != nil {
		c.handleError(key)
		toolbox.Logger.Error(
			"an error occured while deleting orphan pods",
			slog.Any("error", err),
		)
		continue
	}

	// Finished processing key, all ok
	c.Workqueue.Done(key)
    c.Workqueue.Forget(key)
  }
}

// Function retrieves all pods in all namespaces and convertes them to their correct type
func getAllPods(client dynamic.NamespaceableResourceInterface) (*corev1.PodList, error) {

	// List all pods
	podListUnstructured, err := client.Namespace(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		toolbox.Logger.Info("could not list pod resources")
		return nil, err
	}

	// Convert pod list to typed object
	var podList corev1.PodList
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		podListUnstructured.UnstructuredContent(),
		&podList,
	)
	if err != nil {
		toolbox.Logger.Info("could not convert unstructured object to object of type podlist")
		return nil, err
	}

	// Print
	podListBytes, err := json.Marshal(podList)
	if err != nil {
		toolbox.Logger.Info("could not marshal pod list")
		return nil, err
	}
	toolbox.Logger.Debug(
		"finished processing pod list",
		"pod_list",
		string(podListBytes),
	)

	return &podList, nil
}

// Function loops through a list of pods and checks the .metadata.ownerReference field. If the field is empty, 
// then the pod is deleted, otherwise nothing happens. Therefore pods with owners are deleted
func deletePodsWithoutOwners(client dynamic.NamespaceableResourceInterface, podList *corev1.PodList) error {
	for _, pod := range podList.Items {
		if len(pod.ObjectMeta.OwnerReferences) > 0 {
			continue
		}

		err := client.Namespace(pod.GetNamespace()).Delete(context.TODO(), pod.GetName(), metav1.DeleteOptions{})
		if err != nil {
			toolbox.Logger.Info(
				"could not delete pod",
				"namespace",
				pod.GetNamespace(),
				"name",
				pod.GetName(),
			)
			return err
		}

		toolbox.Logger.Debug(
			"successfully deleted pod",
				"namespace",
				pod.GetNamespace(),
				"name",
				pod.GetName(),
		)
	}

	return nil
}
