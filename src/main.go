package main

import (
	// std packages
	"flag"
	"fmt"
	"log/slog"
	"time"
	"os"

	// k8s packages
	kubeinformers "k8s.io/client-go/informers" // inform changes about resources
	"k8s.io/client-go/kubernetes"              // clientset for k8s APIs
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"                   // work with kubeconfigfile
	"sigs.k8s.io/controller-runtime/pkg/manager/signals" // handle signals for controller
	// custom made packages
)

var (
	masterURL  string
	kubeconfig string

	lvl		   string
	Logger 	   slog.Logger
)

func main() {

	// set up signals so we handle the shutdown signal gracefully
	ctx := signals.SetupSignalHandler()

	// Create out-of-cluster kubeconfig
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		Logger.Warn(err.Error(), "Error building out-of-cluster kubeconfig")
		Logger.Info("Falling back to building in-cluster config")

		// Creates in-cluster kubeconfig
		config, err = rest.InClusterConfig()
		if err != nil {
			Logger.Error(err.Error(), "Error building in-cluster kubeconfig")
			os.Exit(1)
		}
	}

	// Create clientset using kubeconfig file
	// Client set can request Kubernetes resources and perform operations
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		Logger.Error(err.Error(), "Error building kubernetes clientset")
		os.Exit(1)
	}

	// Loop

}

// Init flags
func init() {
	// Set flags
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&lvl, "log-level", "warn", "Log level set for program, available values are: debug, info, warn, error")

	flag.Parse()
	fmt.Println("Finished parsing flags")
}

// Init logger
func init() {
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