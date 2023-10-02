package resources

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
	"log"
	"main.go/common"
	"sigs.k8s.io/yaml"
	"testing"
)

// createFakeResourceInformer creates a fake informer for testing purposes
func createFakeResourceInformer(gvr schema.GroupVersionResource, fakeDynamicClient *fakeDynamic.FakeDynamicClient) (fakeResourceInformer cache.SharedIndexInformer) {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(fakeDynamicClient, 0, corev1.NamespaceAll, nil)
	fakeResourceInformer = factory.ForResource(gvr).Informer()
	if fakeResourceInformer == nil {
		log.Fatalf("[ERROR] Resource Informer was not created") // program will exit if this happens
	} else {
		log.Printf("Resource Informer created successfully")
	}
	return fakeResourceInformer
}

// TestCreateResourceInformer tests the creation of a resource informer
func TestCreateResourceInformer(t *testing.T) {
	fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
	resourceGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	informer := createFakeResourceInformer(resourceGVR, fakeDynamicClient)

	if informer == nil {
		t.Fatalf("Failed to create resource informer") // test will fail if this happens
	}
}

// TestEventObject tests the creation of an event object from a map
func TestEventObject(t *testing.T) {
	testDeployment := GetTestDeployment()
	// Marshal the struct to JSON
	jsonData, err := yaml.Marshal(testDeployment)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	// Unmarshal the JSON to a map
	var deploymentMap map[string]interface{}
	err = yaml.Unmarshal(jsonData, &deploymentMap)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}
	deploymentMap["eventType"] = common.EventTypeAdded
	deploymentMap["kind"] = "Deployment"
	deploymentMap["newObject"] = &deploymentMap
	eventObject := EventObject(deploymentMap, true)

	if eventObject.Kind != "Deployment" {
		t.Errorf("Failed to create event object, expected kind Deployment, got %s", eventObject.Kind)
	}

	if eventObject.KubernetesMetadata.Name != "test-deployment" {
		t.Errorf("Failed to create event object, expected name test-deployment, got %s", eventObject.KubernetesMetadata.Name)
	}

	if eventObject.KubernetesMetadata.Namespace != "default" {
		t.Errorf("Failed to create event object, expected namespace default, got %s", eventObject.KubernetesMetadata.Namespace)
	}
}

// TestStructResourceLog tests the creation of a structured resource log
func TestStructResourceLog(t *testing.T) {
	var deploymentMap map[string]interface{}
	testDeployment := GetTestDeployment()
	jsonDeployment, err := json.Marshal(testDeployment)
	if err != nil {
		t.Errorf("Failed to marshal test deployment.\nError:\n %v", err)
	}

	err = json.Unmarshal(jsonDeployment, &deploymentMap)
	if err != nil {
		t.Errorf("Failed to unmarshal test deployment.\nError:\n %v", err)
	}
	deploymentEventMap := map[string]interface{}{

		"eventType": common.EventTypeAdded,
		"kind":      "Deployment",
		"newObject": deploymentMap,
	}
	isStructured, _ := StructResourceLog(deploymentEventMap)

	if !isStructured {
		t.Errorf("Failed to structure resource log")
	}
}
