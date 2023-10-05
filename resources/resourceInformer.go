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
	"os"
	"os/signal"
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

// addInformerEventHandler adds event handlers to the informer.
// It handles add, update, and delete events.
func addInformerEventHandler(resourceInformer cache.SharedIndexInformer) {
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
			go StructResourceLog(event)

		},
		// Handle update event
		UpdateFunc: func(oldObj, newObj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			event = map[string]interface{}{
				"oldObject": oldObj,
				"newObject": newObj,
				"eventType": common.EventTypeModified,
			}
			go StructResourceLog(event)

		},
		// Handle delete event
		DeleteFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			event = map[string]interface{}{
				"newObject": obj,
				"eventType": common.EventTypeDeleted,
			}
			go StructResourceLog(event)

		},
	})

	if err != nil {
		msg := fmt.Sprintf("[ERROR] Failed to add event handler for informer.\nERROR:\n%v", err)
		common.SendLog(msg)

		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go resourceInformer.Run(ctx.Done())

	isSynced := cache.WaitForCacheSync(ctx.Done(), resourceInformer.HasSynced)
	mux.Lock()
	synced = isSynced
	mux.Unlock()

	// If the informer failed to sync, log the error and terminate the program
	if !isSynced {
		log.Fatal("Informer event handler failed to sync.")
	}

	// Wait for the process to be interrupted (e.g. by a SIGINT signal)
	<-ctx.Done()

}

// AddEventHandlers creates informers and adds event handlers for the specified Kubernetes resources.
func AddEventHandlers() {

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
	resourceIndex := 0

	// Loop over the defined resources
	for resourceType, resourceGroup := range resourceAPIList {
		resourceIndex = resourceIndex + 1

		resourceAPI := fmt.Sprintf("%s/v1/%s", resourceGroup, resourceType)
		resourceGVR := schema.GroupVersionResource{Group: resourceGroup, Version: "v1", Resource: resourceType}

		// Attempt to create an informer for the resource
		common.SendLog(fmt.Sprintf("Attempting to create informer for resource API: '%s'", resourceAPI))
		resourceInformer := createResourceInformer(resourceGVR, common.DynamicClient)
		if resourceInformer != nil {
			// If the informer was successfully created, attempt to add an event handler to it
			common.SendLog(fmt.Sprintf("Attempting to add event handler to informer for resource API: '%s'", resourceAPI))
			eventHandlerSync.Add(resourceIndex)
			go addInformerEventHandler(resourceInformer)
			{
				defer eventHandlerSync.Done()
				common.SendLog(fmt.Sprintf("Finished adding event handler to informer for resource API: '%s'", resourceAPI))
			}
		} else {
			// If the informer could not be created, log the failure
			common.SendLog(fmt.Sprintf("Failed to create informer for resource API: '%s'", resourceAPI))
		}
	}

	// Wait for all event handlers to be added
	eventHandlerSync.Wait()
}

// EventObject converts the raw event object into a common.KubernetesEvent object.
func EventObject(rawObject map[string]interface{}) (resourceEventObject common.KubernetesEvent) {
	rawObjUnstructured := &unstructured.Unstructured{}
	rawObjUnstructured.Object = rawObject
	unstructuredObjectJSON, err := rawObjUnstructured.MarshalJSON()
	if err != nil {
		fmt.Printf("[ERROR] Failed to marshal unstructured event object.\nERROR:\n%v", err)
	}
	err = json.Unmarshal(unstructuredObjectJSON, &resourceEventObject)
	if err != nil {
		fmt.Printf("[ERROR] Failed to unmarshal unstructured event object.\nERROR:\n%v", err)
	}

	return resourceEventObject

}

// StructResourceLog structures the event log and sends it.
func StructResourceLog(event map[string]interface{}) (isStructured bool, parsedEvent map[string]interface{}) {
	var msg string
	logEvent := &common.LogEvent{}
	jsonString, err := json.Marshal(event)
	if err != nil {

		fmt.Printf("Failed to marshal structure event log.\nERROR:\n%v", err)
		return
	}
	err = json.Unmarshal(jsonString, logEvent)
	if err != nil {

		// event log.
		fmt.Printf("Failed to unmarshal structure event log.\nERROR:\n%v", err)
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
	go common.SendLog(msg, parsedEvent)
	return isStructured, parsedEvent
}
