package resources

import (
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
	"reflect"
	"sync"
)

// createResourceInformer creates a dynamic resource informer for a given resource GVR.
// It will return nil if the informer fails to create.
func createResourceInformer(resourceGVR schema.GroupVersionResource, clusterClient *dynamic.DynamicClient) (resourceInformer cache.SharedIndexInformer) {
	// Creates a Kubernetes dynamic informer for the cluster API resources
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, 0, corev1.NamespaceAll, nil)
	resourceInformer = factory.ForResource(resourceGVR).Informer()

	// If the informer is nil, log the failure and return nil
	if resourceInformer == nil {
		common.SendLog(fmt.Sprintf("Failed to create informer for resource GVR: '%v'", resourceGVR))
		return nil
	}

	// Get a lister for the resource, which is part of the informer
	lister := factory.ForResource(resourceGVR).Lister()

	// If the lister is nil, informer creation likely failed, log the failure and return nil
	if lister == nil {
		common.SendLog(fmt.Sprintf("Failed to create informer for resource GVR: '%v'", resourceGVR))
		return nil
	}

	return resourceInformer
}

// deleteInternalFields deletes internal fields from a Kubernetes object
func deleteInternalFields(obj *unstructured.Unstructured) {
	if meta, ok := obj.Object[common.Metadata].(map[string]interface{}); ok {
		if _, ok := meta[common.ManagedFields]; ok {
			delete(meta, common.ManagedFields)
		}
		if _, ok := meta[common.ResourceVersion]; ok {
			delete(meta, common.ResourceVersion)
		}
		if annotations, ok := meta[common.Annotations].(map[string]interface{}); ok {
			if _, ok := annotations[common.DeploymentRevision]; ok {
				delete(annotations, common.DeploymentRevision)
				// If annotations is empty, delete it
				if len(annotations) == 0 {
					delete(meta, common.Annotations)
				}
			}
		}
	}
	if _, ok := obj.Object[common.Status]; ok {
		delete(obj.Object, common.Status)
	}
}

// IgnoreInternalChanges determines whether the only changes between two Kubernetes objects are internal.
func IgnoreInternalChanges(oldObj, newObj interface{}) bool {
	oldUnst, ok1 := oldObj.(*unstructured.Unstructured)
	newUnst, ok2 := newObj.(*unstructured.Unstructured)

	if ok1 && ok2 {
		oldCopy := oldUnst.DeepCopy()
		newCopy := newUnst.DeepCopy()

		deleteInternalFields(oldCopy)
		deleteInternalFields(newCopy)
		newJson, _ := json.Marshal(newCopy)
		oldJson, _ := json.Marshal(oldCopy)

		return string(oldJson) == string(newJson)
	}

	return false
}

// inFlightEvents tracks events still being processed, so shutdown waits for them before the final flush
var inFlightEvents sync.WaitGroup

// handleEvent is replaced in tests to hold an event in progress
var handleEvent = func(event map[string]interface{}) { StructResourceLog(event) }

// processEvent structures and sends the event in the background
func processEvent(event map[string]interface{}) {
	inFlightEvents.Add(1)
	go func() {
		defer inFlightEvents.Done()
		handleEvent(event)
	}()
}

// addInformerEventHandler adds event handlers to the informer and runs it until ctx is done.
// It handles add, update, and delete events.
func addInformerEventHandler(ctx context.Context, resourceInformer cache.SharedIndexInformer) {
	var event map[string]interface{}
	synced := false

	mux := &sync.RWMutex{}

	_, err := resourceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Handle add event
		AddFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			event = map[string]interface{}{
				"newObject": obj,
				"eventType": common.EventTypeAdded,
			}
			processEvent(event)

		},
		// Handle update event
		UpdateFunc: func(oldObj, newObj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			if IgnoreInternalChanges(oldObj, newObj) {
				return // ignore internal cluster updates
			} else {
				event = map[string]interface{}{
					"oldObject": oldObj,
					"newObject": newObj,
					"eventType": common.EventTypeModified,
				}
				processEvent(event)
			}

		},
		// Handle delete event
		DeleteFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			event = map[string]interface{}{
				"newObject": unwrapDeletedObject(obj),
				"eventType": common.EventTypeDeleted,
			}
			processEvent(event)

		},
	})

	if err != nil {
		msg := fmt.Sprintf("[ERROR] Failed to add event handler for informer.\nERROR:\n%v", err)
		common.SendLog(msg)
		return
	}

	// Start the informer
	go resourceInformer.Run(ctx.Done())

	// Wait for all caches to sync
	cache.WaitForCacheSync(ctx.Done(), resourceInformer.HasSynced)

	// Set synced to true after all informers have synced
	mux.Lock()
	synced = true
	mux.Unlock()

	// If the informer failed to sync, log the error and terminate the program, unless it's shutting down
	if !resourceInformer.HasSynced() && ctx.Err() == nil {
		log.Fatal("Informer event handler failed to sync.")
	}

	// Wait for the process to be stopped (SIGINT or SIGTERM)
	<-ctx.Done()

	// Stop accepting events. Lock waits for running handlers, so no event starts after this
	mux.Lock()
	synced = false
	mux.Unlock()
}

// unwrapDeletedObject returns the deleted object itself when the informer missed the delete and passed a tombstone
func unwrapDeletedObject(obj interface{}) interface{} {
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		return tombstone.Obj
	}
	return obj
}

// AddEventHandlers creates informers and adds event handlers for the specified Kubernetes resources.
// It blocks until ctx is done and every event in progress was handed to the sender.
func AddEventHandlers(ctx context.Context) {

	// Define the Kubernetes resources for which to create informers and add event handlers
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

	// Loop over the defined resources
	for resourceType, resourceGroup := range resourceAPIList {
		resourceAPI := fmt.Sprintf("%s/v1/%s", resourceGroup, resourceType)
		resourceGVR := schema.GroupVersionResource{Group: resourceGroup, Version: "v1", Resource: resourceType}

		// Attempt to create an informer for the resource
		log.Printf("Attempting to create informer for resource API: '%s'", resourceAPI)
		resourceInformer := createResourceInformer(resourceGVR, common.DynamicClient)
		if resourceInformer != nil {
			// If the informer was successfully created, attempt to add an event handler to it
			log.Printf("Attempting to add event handler to informer for resource API: '%s'", resourceAPI)
			eventHandlerSync.Add(1)
			go func() {
				defer eventHandlerSync.Done()
				addInformerEventHandler(ctx, resourceInformer)
			}()
		} else {
			// If the informer could not be created, log the failure
			common.SendLog(fmt.Sprintf("Failed to create informer for resource API: '%s'", resourceAPI))
		}
	}

	// Wait for all informers to stop, then for the events they already started
	eventHandlerSync.Wait()
	inFlightEvents.Wait()
}

// EventObject converts the raw event object into a common.KubernetesEvent object.
func EventObject(rawObject map[string]interface{}) (resourceEventObject common.KubernetesEvent) {
	rawObjUnstructured := &unstructured.Unstructured{}
	rawObjUnstructured.Object = rawObject
	unstructuredObjectJSON, err := rawObjUnstructured.MarshalJSON()
	if err != nil {
		log.Printf("[ERROR] Failed to marshal unstructured event object.\nERROR:\n%v", err)
	}
	err = json.Unmarshal(unstructuredObjectJSON, &resourceEventObject)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal unstructured event object.\nERROR:\n%v", err)
	}

	return resourceEventObject

}

// StructResourceLog structures the event log and sends it.
func StructResourceLog(event map[string]interface{}) (isStructured bool, parsedEvent map[string]interface{}) {
	var msg string
	logEvent := &common.LogEvent{}
	jsonString, err := json.Marshal(event)
	if err != nil {

		log.Printf("Failed to marshal structure event log.\nERROR:\n%v", err)
		return
	}
	err = json.Unmarshal(jsonString, logEvent)
	if err != nil {

		// event log.
		log.Printf("Failed to unmarshal structure event log.\nERROR:\n%v", err)
		return
	}
	eventType := event["eventType"].(string)
	newResourceObj := EventObject(logEvent.NewObject)
	resourceKind := newResourceObj.Kind
	resourceName := newResourceObj.KubernetesMetadata.Name
	resourceNamespace := newResourceObj.KubernetesMetadata.Namespace
	newResourceVersion := newResourceObj.KubernetesMetadata.ResourceVersion
	msg = common.ParseEventMessage(eventType, resourceName, resourceKind, resourceNamespace, newResourceVersion)
	if eventType == common.EventTypeModified {
		oldResourceObj := EventObject(logEvent.OldObject)
		oldResourceName := oldResourceObj.KubernetesMetadata.Name
		oldResourceNamespace := oldResourceObj.KubernetesMetadata.Namespace
		oldResourceVersion := oldResourceObj.KubernetesMetadata.ResourceVersion
		msg = common.ParseEventMessage(eventType, oldResourceName, resourceKind, oldResourceNamespace, newResourceVersion, oldResourceVersion)

	}

	// Get cluster related resources
	clusterRelatedResources := GetClusterRelatedResources(resourceKind, resourceName, resourceNamespace)
	// If the cluster related resources are valid, add them to the event
	if reflect.ValueOf(clusterRelatedResources).IsValid() {
		event["relatedClusterServices"] = clusterRelatedResources
	}

	jsonString, _ = json.Marshal(event)
	err = json.Unmarshal(jsonString, &parsedEvent)

	// If there is an error in parsing the resource event logs, log the error
	if err != nil {
		log.Printf("[ERROR] Failed to parse resource event logs.\nERROR:\n%v", err)
	} else {
		isStructured = true
	}
	// Send the parsed event log
	common.SendLog(msg, parsedEvent)
	return isStructured, parsedEvent
}
