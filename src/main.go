package main

import (
	// std packages
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	// k8s packages
	"k8s.io/client-go/informers"
	kubeinformers "k8s.io/client-go/informers" // inform changes about resources
	"k8s.io/client-go/kubernetes"              // clientset for k8s APIs
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd" // work with kubeconfigfile
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/manager"         // controller manager
	"sigs.k8s.io/controller-runtime/pkg/manager/signals" // handle signals for controller
	corev1 "k8s.io/api/core/v1"
	// custom made packages
)

var (
	masterURL  string
	kubeconfig string

	lvl		   string
	Logger 	   *slog.Logger
)

func init(){
	initFlags()
	initLogger()
}

func main() {

	// Create out-of-cluster kubeconfig
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		Logger.Warn("error building out-of-cluster kubeconfig", slog.Any("warn", err.Error()))
		Logger.Info("falling back to building in-cluster config")

		// Creates in-cluster kubeconfig
		config, err = rest.InClusterConfig()
		if err != nil {
			Logger.Error("error building in-cluster kubeconfig", slog.Any("error", err))
			os.Exit(1)
		}
	}
	Logger.Info("successfully built kubeconfig file")

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Logger.Error("error creating kubernetes client", slog.Any("error", err))
	}
	Logger.Info("successfully created kubernetes client")

	// Create SharedinformerFactory (resync every 30 seconds)
  	informerFactory := informers.NewSharedInformerFactory(clientset,  30*time.Second)
	// Create informer for for auto pod terminator
	podTerminatorGenericInformer, err := informerFactory.ForResource(APTGroupVersionResource)
	if err != nil {
		Logger.Error("error creating shared informer for auto pod terminator", slog.Any("error", err))
		os.Exit(1)
	}
	podTerminatorInformer := podTerminatorGenericInformer.Informer()
	// Create informer for pod
	podInformer:= informerFactory.Core().V1().Pods().Informer()

	// Set up an indexer for Pods, which allows us to retrieve objects by key
	// The indexer is used to efficiently look up objects in the workqueue
	// without needing to fetch them from the API server every time
	podTerminatorIndexer := podTerminatorInformer.GetIndexer()
	podIndexer := podInformer.GetIndexer()

	// Create queue for workloads
	queue := workqueue.NewTypedRateLimitingQueue[string](
		workqueue.DefaultTypedControllerRateLimiter[string](),
	)

	// Register event handlers
	podTerminatorInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			APTItem := obj.(*PodTerminator)
			Logger.Info(
				"deleting pod terminator object",
				"namespace",
				APTItem.GetNamespace(),
				"name",
				APTItem.GetName(),
			)
		},
		UpdateFunc: func(_, newObj interface{}) {
			APTItem := newObj.(*PodTerminator)
			Logger.Info(
				"updating pod terminator object",
				"namespace",
				APTItem.GetNamespace(),
				"name",
				APTItem.GetName(),
			)
		},
		DeleteFunc: func(obj interface{}) {
			APTItem := obj.(*PodTerminator)
			Logger.Info(
				"deleting pod terminator object",
				"namespace",
				APTItem.GetNamespace(),
				"name",
				APTItem.GetName(),
			)
		},
	})

	// Create indexer
	customIndexer := cache.Indexers{
		""
	}
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

// Init logger
func initLogger() {
    logLevel := new(slog.LevelVar)
    // Set log level
    switch lvl {
    case "debug":
        logLevel.Set(slog.LevelDebug)
    case "info":
        logLevel.Set(slog.LevelInfo)
    case "error":
        logLevel.Set(slog.LevelError)
    // Default is warn
    default:
        logLevel.Set(slog.LevelWarn)
    }

    // Create logger
    Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: logLevel,
    }))

    Logger.Debug("successfully created logger")
}