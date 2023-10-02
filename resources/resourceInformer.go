package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"log"
	"main.go/common"
	"os"
	"os/signal"
	"reflect"
	"sync"
)

var wg sync.WaitGroup

// createResourceInformer creates a dynamic resource informer for a given resource GVR.
func createResourceInformer(resourceGVR schema.GroupVersionResource, clusterClient *dynamic.DynamicClient) (resourceInformer cache.SharedIndexInformer) {
	// Creates a Kubernetes dynamic informer for the cluster API resources
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, 0, corev1.NamespaceAll, nil)
	resourceInformer = factory.ForResource(resourceGVR).Informer()

	if resourceInformer == nil {
		common.SendLog(fmt.Sprintf("Failed to create informer for resource GVR: '%v'", resourceGVR))
		return nil
	}
	// Creates a Kubernetes dynamic informer for the cluster API resources
	// Get a lister for the resource, which is part of the informer
	lister := factory.ForResource(resourceGVR).Lister()

	// If the lister is nil, informer creation likely failed
	if lister == nil {
		common.SendLog(fmt.Sprintf("Failed to create informer for resource GVR: '%v'", resourceGVR))
		return nil
	}

	return resourceInformer
}

// AddInformerEventHandler adds a new event handler to a given resource informer.
// It logs events when a resource is added, updated, or deleted.
func AddInformerEventHandler(resourceInformer cache.SharedIndexInformer) (synced bool) {
	var parsedEventLog []byte
	// Check if the resource informer is nil
	if resourceInformer == nil {
		log.Println("[ERROR] Resource informer is nil")
		return false
	}

	// Create a new mutex for handling read and write locks
	mux := &sync.RWMutex{}

	// Add event handler to the resource informer
	_, err := resourceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// This function gets called when a resource gets added
		AddFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if synced {
				_, parsedEventLog = StructResourceLog(map[string]interface{}{
					"eventType": common.EventTypeAdded,
					"newObject": obj,
				})
			}
		},
		// This function gets called when a resource gets updated
		UpdateFunc: func(oldObj, newObj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if synced {
				_, parsedEventLog = StructResourceLog(map[string]interface{}{
					"eventType": common.EventTypeModified,
					"newObject": newObj,
					"oldObject": oldObj,
				})
			}
		},
		// This function gets called when a resource gets deleted
		DeleteFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if synced {
				_, parsedEventLog = StructResourceLog(map[string]interface{}{
					"eventType": common.EventTypeDeleted,
					"newObject": obj,
				})
			}
		},
	})

	// Log any errors in adding the event handler
	if err != nil {
		common.SendLog(fmt.Sprintf("[ERROR] Failed to add event handler for informer.\nERROR:\n%v", err))
		return
	}
	common.SendLog(string(parsedEventLog))
	// Create a new context that will get cancelled when an interrupt signal is received
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Create a channel to indicate when the informer has started
	started := make(chan bool)

	// Run the informer in a separate goroutine
	go func() {
		resourceInformer.Run(ctx.Done())
		close(started)
	}()

	// Wait for the informer to start
	<-started

	// Wait for the informer's cache to sync
	mux.Lock()
	synced = cache.WaitForCacheSync(ctx.Done(), resourceInformer.HasSynced)
	mux.Unlock()

	// Log if the informer failed to sync
	if !synced {
		log.Printf("Informer event handler failed to sync.")
	}

	// Wait for the context to be done
	<-ctx.Done()

	// Return whether the informer's cache was synced successfully
	return synced

}

// AddEventHandlers adds event handlers and creates resource informer per cluster API resource.
func AddEventHandlers() {
	resourceAPIList := map[string]string{
		"configmaps":          "",
		"deployments":         "apps",
		"daemonsets":          "apps",
		"secrets":             "",
		"serviceaccounts":     "",
		"statefulsets":        "apps",
		"clusterroles":        "rbac.authorization.k8s.io",
		"clusterrolebindings": "rbac.authorization.k8s.io",
	}
	var eventHandlerSync sync.WaitGroup
	resourceIndex := 0
	routinesLimit := make(chan bool, len(resourceAPIList)) // limit to number of concurrent goroutines to list resources

	// Creates informer for each cluster API and events handler for each informer
	for resourceType, resourceGroup := range resourceAPIList {
		routinesLimit <- true //  will block if there is already goroutines running for the limit number of resources
		resourceIndex = resourceIndex + 1
		resourceGVR := schema.GroupVersionResource{Group: resourceGroup, Version: "v1", Resource: resourceType}
		resourceAPI := fmt.Sprintf("%s/v1/%s", resourceGroup, resourceType)
		resourceInformer := createResourceInformer(resourceGVR, common.DynamicClient)
		if resourceInformer == nil {
			common.SendLog(fmt.Sprintf("Failed to create informer for resource API: '%s'", resourceAPI))
			<-routinesLimit // release a slot when the informer is nil
			continue        // Skip to next iteration if informer is nil
		}
		common.SendLog(fmt.Sprintf("Attempting to add event handler to informer for resource API: '%s'", resourceAPI))
		eventHandlerSync.Add(1)
		go func(resourceInformer cache.SharedIndexInformer, resourceAPI string) {
			// Use defer to ensure Done() is called even if the goroutine exits prematurely
			defer eventHandlerSync.Done()

			// Log when the goroutine starts
			common.SendLog(fmt.Sprintf("Adding event handler for resource API: '%s'", resourceAPI))

			AddInformerEventHandler(resourceInformer)

			// Log when the goroutine finishes
			common.SendLog(fmt.Sprintf("Finished adding event handler to informer for resource API: '%s'", resourceAPI))
			<-routinesLimit // release a slot when the goroutine finishes

		}(resourceInformer, resourceAPI) // Pass the loop variables here
	}
	eventHandlerSync.Wait()
}

// EventObject transforms a raw object into a KubernetesEvent object.
// It takes a map representing the raw object and a boolean indicating whether the object is new or n
// ot.
func EventObject(rawObj map[string]interface{}, isNew bool) (resourceObject common.KubernetesEvent) {

	// Check if the raw object or its "newObject" and "oldObject" fields are nil
	// Check if the raw object is nil
	if rawObj == nil {
		log.Println("[ERROR] rawObj is nil.")
		// Return an empty KubernetesEvent object if the raw object is invalid
		return resourceObject
	}

	// Initialize an empty unstructured object
	rawUnstructuredObj := unstructured.Unstructured{}
	// Initialize a buffer to store the JSON-encoded raw object
	var buffer bytes.Buffer
	// Encode the raw object into JSON and write it to the buffer
	err := json.NewEncoder(&buffer).Encode(rawObj)
	if err != nil {
		log.Printf("Failed to encode unstructed resource object bytes:\n%v\nError:\n%v", rawObj, err)

	}

	// Unmarshal the JSON-encoded raw object into a KubernetesEvent object
	err = json.Unmarshal(buffer.Bytes(), &resourceObject)
	if err != nil {
		log.Printf("Failed to unmarshal resource object:\n%v\nError:\n%v", rawObj, err)
	} else {
		// If unmarshalling is successful, determine whether to set the unstructured object's content based on the "isNew" flag
		if isNew {
			// If the object is new, set the unstructured object's content to the new object's content
			rawUnstructuredObj.Object = resourceObject.NewObject
		} else {
			// If the object is not new, set the unstructured object's content to the old object's content
			rawUnstructuredObj.Object = resourceObject.OldObject
		}
	}

	return resourceObject
}

// StructResourceLog receives an event and logs it in a structured format.
func StructResourceLog(event map[string]interface{}) (isStructured bool, marshaledEvent []byte) {
	// Check if event is nil
	if event == nil {
		log.Println("[ERROR] Event is nil")
		return false, nil
	}
	// Check if the newObject field of rawObj is nil, if it is required
	if event["newObject"] == nil {
		log.Println("[ERROR] rawObj does not have required field: newObject.")
		// Return an empty KubernetesEvent object if the raw object is invalid
		return
	}

	// Assert that event["eventType"] is a string
	eventType, ok := event["eventType"].(string)
	if !ok {
		log.Println("[ERROR] eventType is not a string")
		return false, nil
	}

	// Initialize an empty LogEvent
	logEvent := &common.LogEvent{}

	// Marshal the event to a string
	eventStr, _ := json.Marshal(event)

	// Unmarshal the string back to a logEvent
	err := json.Unmarshal(eventStr, logEvent)
	if err != nil {

		return false, nil
	}

	// Get the new event object
	newEventObj := EventObject(logEvent.NewObject, true)

	// Get the resource details from the new event object
	resourceKind := newEventObj.Kind
	resourceName := newEventObj.KubernetesMetadata.Name
	resourceNamespace := newEventObj.KubernetesMetadata.Namespace
	newResourceVersion := newEventObj.ResourceVersion

	var msg string
	// If event is a modification event, get the old event object and parse the event message accordingly
	if eventType == common.EventTypeModified {
		oldEventObj := EventObject(logEvent.OldObject, false)
		oldResourceName := oldEventObj.KubernetesMetadata.Name
		oldResourceNamespace := oldEventObj.KubernetesMetadata.Namespace
		oldResourceVersion := oldEventObj.KubernetesMetadata.ResourceVersion
		msg = common.ParseEventMessage(eventType, oldResourceName, resourceKind, oldResourceNamespace, newResourceVersion, oldResourceVersion)
	} else {
		// If event is not a modification event, parse the event message with only the new event object
		msg = common.ParseEventMessage(eventType, resourceName, resourceKind, resourceNamespace, newResourceVersion)
	}

	// Get the related cluster services for the resource
	relatedClusterServices := GetClusterRelatedResources(resourceKind, resourceName, resourceNamespace)
	if !reflect.ValueOf(relatedClusterServices).IsZero() {
		event["relatedClusterServices"] = relatedClusterServices
	}
	event["message"] = msg

	// Marshal the event to a string
	marshaledEvent, err = json.Marshal(event)
	if err != nil {
		log.Printf("[ERROR] Failed to marshel resource event logs.\nERROR:\n%v", err)
	}

	// Mark the goroutine as done
	wg.Add(1)
	defer wg.Done()

	// Return true indicating the log is structured
	isStructured = true

	return isStructured, marshaledEvent
}
