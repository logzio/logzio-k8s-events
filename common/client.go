package common

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"log"
)

// ConfigureClusterDynamicClient configures an in-cluster dynamic client for the Kubernetes cluster.
// It exits on failure, so it never returns nil.
func ConfigureClusterDynamicClient() (dynamicClient *dynamic.DynamicClient) {

	// Getting the in-cluster configuration
	clusterConfig, err := rest.InClusterConfig()
	// If there is an error in getting the configuration, log the error and exit
	if err != nil {
		log.Fatalf("Failed to get in-cluster configuration for dynamic Kubernetes client.\nError:\n%v\n", err)
	}

	// Creating the dynamic client using the cluster configuration
	dynamicClient, err = dynamic.NewForConfig(clusterConfig)

	// If there is an error in creating the dynamic client, log the error and exit
	if err != nil {
		log.Fatalf("Failed to configure dynamic Kubernetes client.\nError:\n%v\n", err)
	}

	return dynamicClient
}
