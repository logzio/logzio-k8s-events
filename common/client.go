package common

import (
	"fmt"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

var K8sClient *kubernetes.Clientset
var DynamicClient *dynamic.DynamicClient

func CreateClusterClient() {
	// Create a Kubernetes client.
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	K8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ConfigureClusterDynamicClient() (clusterClient *dynamic.DynamicClient) {
	//
	var err error
	var clusterConfig *rest.Config
	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig != "" {
		clusterConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		clusterConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatalln(err)
	}
	clusterClient, err = dynamic.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalln(err)
	}
	return clusterClient
}
