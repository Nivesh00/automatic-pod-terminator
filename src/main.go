package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	toolbox "github.com/Nivesh00/automatic-pod-terminator/src/toolbox"
	controller "github.com/Nivesh00/automatic-pod-terminator/src/controller"
)

var (
	masterURL  string
	kubeconfig string

	lvl		   string
)

func init(){
	initFlags()
	toolbox.InitLogger(lvl)
}

func main() {

	// Create out-of-cluster kubeconfig
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		toolbox.Logger.Warn("failed building out-of-cluster kubeconfig", slog.Any("warn", err.Error()))
		toolbox.Logger.Info("falling back to building in-cluster config")

		// Creates in-cluster kubeconfig if previous task failed
		config, err = rest.InClusterConfig()
		if err != nil {
			toolbox.Logger.Error("error building in-cluster kubeconfig", slog.Any("error", err))
			os.Exit(1)
		}
	}
	toolbox.Logger.Info("successfully built kubeconfig file")

	// Create Kubernetes client
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		toolbox.Logger.Error("error creating kubernetes client", slog.Any("error", err))
	}
	toolbox.Logger.Info("successfully created kubernetes client")

	// Create resource GroupVersionResource
    podTerminatorGVR := schema.GroupVersionResource{Group: "k8s.niv-ram.dev", Version: "v1", Resource: "podterminators"}

	// Create contoller instance
	podTerminatorController := controller.PodTerminatorController{}
	// Create workqueue for controller instance
	podTerminatorController.Workqueue = workqueue.NewTypedRateLimitingQueue(
		workqueue.DefaultTypedControllerRateLimiter[string](),
	)

	// Dynamic informers help reduce API calls to Kubernetes API server and boost performance. They watches
	// for changes in the cluster. Data is stored in a thread-safe local in-memory cache.
	// Shared informer manages informers' lifecycle centrally and ensures efficient resource utilization.

	// Create dynamic informer for all namespaces with no resync period
    factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clientset, 0, corev1.NamespaceAll, nil)

	// Get informer
	podTerminatorController.Informer = factory.ForResource(podTerminatorGVR).Informer()
	// Get indexer
	podTerminatorController.Indexer  = podTerminatorController.Informer.GetIndexer()

	// Add event handler to informer. Parameters passed to function as `interface{}` and it is assumed
	// that instances are of `*unstructured.Unstructured` (map of k8s object) and can be cast safely
	podTerminatorController.Informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			podTerminatorController.AddToWorkqueue(obj, "CREATE")
		},
        UpdateFunc: func(_, obj interface{}) {
			podTerminatorController.AddToWorkqueue(obj, "UPDATE")
		},
        DeleteFunc: func(obj interface{}) {
			podTerminatorController.AddToWorkqueue(obj, "DELETE")
		},
	})

	// `context.Context` is only stopped when an interrupt signal is received. Application will therefore
	// run indefinitely unless `os.interrupt` is called
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

	// Keep informer running. Run must be called first before cache syncing can start, therefore a
	// is used to not halt the program. Informer therfore starts watching
	factory.Start(ctx.Done())

	// Since informers store data in-memory, all data is lost on reboot. Everytime application starts back up,
	// the informer needs to sync with the current status of the cluster
	// Starts a goroutine internally
	factory.WaitForCacheSync(ctx.Done())

	toolbox.Logger.Info("cache successfully synced")
	toolbox.Logger.Info("starting controller worker")

	go podTerminatorController.Worker(clientset)

	<-ctx.Done()

	podTerminatorController.Workqueue.ShutDown()

	toolbox.Logger.Info("controller shutting down...")

}

// Init flags
func initFlags() {
	// Set flags
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&lvl, "log-level", "warn", "Log level set for program, available values are: debug, info, warn, error")

	flag.Parse()
	fmt.Println("Finished parsing flags")
}