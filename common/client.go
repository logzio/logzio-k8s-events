package common

import (
	"fmt"
	"k8s.io/client-go/dynamic"         // Importing the dynamic client package
	"k8s.io/client-go/kubernetes"      // Importing the kubernetes client package
	"k8s.io/client-go/rest"            // Importing the rest client package
	"k8s.io/client-go/tools/clientcmd" // Importing the clientcmd package for building config from kubeconfig
	"log"
	"os" // Importing the os package for reading environment variables
)

var K8sClient *kubernetes.Clientset      // Global variable for the Kubernetes client
var DynamicClient *dynamic.DynamicClient // Global variable for the dynamic client

func CreateClusterClient() {
	// This function creates a Kubernetes client using in-cluster configuration

	// Getting the in-cluster configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Creating the Kubernetes client using the in-cluster configuration
	K8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ConfigureClusterDynamicClient() (clusterClient *dynamic.DynamicClient) {
	// This function configures a dynamic client for the Kubernetes cluster
	// by either using the KUBECONFIG environment variable or falling back to in-cluster configuration

	var err error
	var clusterConfig *rest.Config

	// Reading the KUBECONFIG environment variable
	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig != "" {
		// If KUBECONFIG is set, build the configuration from KUBECONFIG
		clusterConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		// If KUBECONFIG is not set, get the in-cluster configuration
		clusterConfig, err = rest.InClusterConfig()
	}

	// If there is an error in getting the configuration, log the error and exit
	if err != nil {
		log.Fatalln(err)
	}

	// Creating the dynamic client using the cluster configuration
	clusterClient, err = dynamic.NewForConfig(clusterConfig)

	// If there is an error in creating the dynamic client, log the error and exit
	if err != nil {
		log.Fatalln(err)
	}

	return clusterClient
}
