package resources

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
	"log"
	"main.go/common"
	"testing"
	"time"
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
	// Initialize an empty LogEvent
	logEvent := &common.LogEvent{}

	// Marshal the event to a string
	eventStr, _ := json.Marshal(deploymentMap)
	// Unmarshal the string back to a logEvent
	err = json.Unmarshal(eventStr, logEvent)
	if err != nil {
		log.Printf(" Error unmarshalling event")
		return
	}

	log.Printf("NEW RAW OBJECT: %v\n", logEvent.NewObject)
	// Get the new event object
	eventObject := logEvent.KubernetesEvent
	log.Printf("NewObject before EventObject: %v\n", newObject)
	eventObject = EventObject(newObject)
	log.Printf("KubernetesEvent after EventObject: %v\n", eventObject)

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

// TestIgnoreInternalChanges tests the function IgnoreInternalChanges
func TestIgnoreInternalChanges(t *testing.T) {
	// Create two identical Kubernetes objects
	oldObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			common.Metadata: map[string]interface{}{
				"name": "test",
				common.Annotations: map[string]interface{}{
					common.DeploymentRevision: "1",
				},
				common.ResourceVersion: "1000",
				common.ManagedFields:   []interface{}{},
			},
			common.Status: map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	newObj := oldObj.DeepCopy()

	// Change the status field and internal fields of newObj
	newObj.Object[common.Status] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "Ready",
				"status": "False",
			},
		},
	}
	newObj.Object[common.Metadata].(map[string]interface{})[common.Annotations].(map[string]interface{})[common.DeploymentRevision] = "2"
	newObj.Object[common.Metadata].(map[string]interface{})[common.ResourceVersion] = "1001"
	newObj.Object[common.Metadata].(map[string]interface{})[common.ManagedFields] = []interface{}{
		map[string]interface{}{
			"manager":    "kubectl",
			"operation":  "Update",
			"apiVersion": "v1",
			"fieldsType": "FieldsV1",
			"fieldsV1": map[string]interface{}{
				"f:metadata": map[string]interface{}{
					"f:annotations": map[string]interface{}{
						".":                                   map[string]interface{}{},
						"f:deployment.kubernetes.io/revision": map[string]interface{}{},
					},
				},
			},
		},
	}

	// Pass the objects to IgnoreInternalChanges
	shouldIgnore := IgnoreInternalChanges(oldObj, newObj)

	// Check that the function correctly ignored the status change and internal field changes
	if !shouldIgnore {
		t.Errorf("IgnoreInternalChanges did not ignore internal changes")
	}
}

func TestUnwrapDeletedObject(t *testing.T) {
	secret := &unstructured.Unstructured{Object: map[string]interface{}{"kind": "Secret"}}

	if got := unwrapDeletedObject(cache.DeletedFinalStateUnknown{Key: "ns/name", Obj: secret}); got != secret {
		t.Errorf("expected the tombstone to be unwrapped, got %v", got)
	}
	if got := unwrapDeletedObject(secret); got != secret {
		t.Errorf("expected a regular object to be returned as is, got %v", got)
	}
}

func TestAddInformerEventHandlerWaitsForEventsInProgressOnShutdown(t *testing.T) {
	started, release := make(chan struct{}, 100), make(chan struct{})
	t.Cleanup(func() { handleEvent = func(event map[string]interface{}) { StructResourceLog(event) } })
	handleEvent = func(map[string]interface{}) {
		started <- struct{}{}
		<-release
	}

	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	client := fakeDynamic.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{gvr: "DeploymentList"})
	informer := createFakeResourceInformer(gvr, client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		addInformerEventHandler(ctx, informer)
		close(stopped)
	}()

	// Events before the handler finished syncing are ignored, so create Deployments until one is handled
	deadline := time.After(5 * time.Second)
	for i := 0; ; i++ {
		deployment := &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "apps/v1", "kind": "Deployment",
			"metadata": map[string]interface{}{"name": fmt.Sprintf("test-%d", i), "namespace": "default"},
		}}
		if _, err := client.Resource(gvr).Namespace("default").Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
		select {
		case <-started:
		case <-time.After(50 * time.Millisecond):
			continue
		case <-deadline:
			t.Fatal("no event was handled")
		}
		break
	}

	cancel()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("event handler did not stop after the context was done")
	}

	eventsDone := make(chan struct{})
	go func() {
		inFlightEvents.Wait()
		close(eventsDone)
	}()
	select {
	case <-eventsDone:
		t.Fatal("shutdown must wait for the event in progress")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	select {
	case <-eventsDone:
	case <-time.After(5 * time.Second):
		t.Fatal("events in progress did not finish")
	}
}

func TestStructResourceLogRemovesManagedFields(t *testing.T) {
	object := func(generation int64) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "apps/v1", "kind": "Deployment",
			"metadata": map[string]interface{}{
				"name": "app", "namespace": "default", "generation": generation,
				"managedFields": []interface{}{map[string]interface{}{"manager": "kubectl"}},
			},
		}}
	}
	oldObject, newObject := object(1), object(2)

	_, parsedEvent := StructResourceLog(map[string]interface{}{
		"eventType": common.EventTypeModified, "oldObject": oldObject, "newObject": newObject,
	})

	for _, objectField := range []string{"newObject", "oldObject"} {
		metadata := parsedEvent[objectField].(map[string]interface{})["metadata"].(map[string]interface{})
		if _, ok := metadata["managedFields"]; ok {
			t.Errorf("%s.metadata.managedFields must not be shipped", objectField)
		}
		if metadata["name"] != "app" || metadata["generation"] == nil {
			t.Errorf("%s.metadata must keep the other fields, got %v", objectField, metadata)
		}
	}
	if _, ok := newObject.Object["metadata"].(map[string]interface{})["managedFields"]; !ok {
		t.Error("the informer's cached object must not be modified")
	}
}
