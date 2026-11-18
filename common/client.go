package common

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"sync"
)

var K8sClient kubernetes.Interface
var DynamicClient *dynamic.DynamicClient
var createClusterClientOnce sync.Once

// CreateClusterClient creates the Kubernetes client using in-cluster configuration, once
func CreateClusterClient() {
	createClusterClientOnce.Do(func() {
		// Getting the in-cluster configuration
		clusterConfig, err := rest.InClusterConfig()
		// If there is an error in getting the configuration, log the error and return
		if err != nil {
			log.Printf("Failed to get in-cluster configuration for Kubernetes client.\nError:\n%v\n", err)
			return
		}

		// Creating the Kubernetes client using the in-cluster configuration
		client, err := kubernetes.NewForConfig(clusterConfig)
		if err != nil {
			log.Printf("Failed to configure Kubernetes client.\nError:\n%v\n", err)
			return
		}
		K8sClient = client
	})
}

// ConfigureClusterDynamicClient configures an in-cluster dynamic client for the Kubernetes cluster
func ConfigureClusterDynamicClient() (DynamicClient *dynamic.DynamicClient) {

	// Getting the in-cluster configuration
	clusterConfig, err := rest.InClusterConfig()
	// If there is an error in getting the configuration, log the error and exit
	if err != nil {
		log.Fatalf("Failed to get in-cluster configuration for dynamic Kubernetes client.\nError:\n%v\n", err)
	}

	// Creating the dynamic client using the cluster configuration
	DynamicClient, err = dynamic.NewForConfig(clusterConfig)

	// If there is an error in creating the dynamic client, log the error and exit
	if err != nil {
		log.Fatalf("Failed to configure dynamic Kubernetes client.\nError:\n%v\n", err)
	}

	return DynamicClient
}
