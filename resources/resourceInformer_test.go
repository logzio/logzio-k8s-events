package resources

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
	"log"
	"main.go/common"
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
	jsonData, err := json.Marshal(testDeployment)
	if err != nil {
		t.Fatalf("Failed to marshal test deployment: %s", err)
	}

	var deploymentMap map[string]interface{}
	if err = json.Unmarshal(jsonData, &deploymentMap); err != nil {
		t.Fatalf("Failed to unmarshal deployment map: %s", err)
	}

	// Deep copy of the map
	mapBytes, _ := json.Marshal(deploymentMap)
	var newObject map[string]interface{}
	if err = json.Unmarshal(mapBytes, &newObject); err != nil {
		t.Fatalf("Failed to unmarshal new object: %s", err)
	}

	deploymentMap["eventType"] = common.EventTypeAdded
	deploymentMap["kind"] = "Deployment"
	deploymentMap["newObject"] = newObject
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
	testDeployment := GetTestDeployment()
	jsonData, err := json.Marshal(testDeployment)
	if err != nil {
		t.Fatalf("Failed to marshal test deployment: %s", err)
	}

	var deploymentMap map[string]interface{}
	if err = json.Unmarshal(jsonData, &deploymentMap); err != nil {
		t.Fatalf("Failed to unmarshal deployment map: %s", err)
	}

	// Deep copy of the map
	mapBytes, _ := json.Marshal(deploymentMap)
	var newObject map[string]interface{}
	if err = json.Unmarshal(mapBytes, &newObject); err != nil {
		t.Fatalf("Failed to unmarshal new object: %s", err)
	}

	deploymentMap["eventType"] = common.EventTypeAdded
	deploymentMap["kind"] = "Deployment"
	deploymentMap["newObject"] = newObject

	isStructured, _ := StructResourceLog(deploymentMap)

	if !isStructured {
		t.Errorf("Failed to structure resource log")
	}
}
