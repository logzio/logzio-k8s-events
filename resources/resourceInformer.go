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

func createResourceInformer(resourceGroup string, resourceType string, clusterClient *dynamic.DynamicClient) (informer cache.SharedIndexInformer) {
	//
	resource := schema.GroupVersionResource{Group: resourceGroup, Version: "v1", Resource: resourceType}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, 0, corev1.NamespaceAll, nil)
	informer = factory.ForResource(resource).Informer()

	return informer
}

func addInformerEventHandler(resourceInformer cache.SharedIndexInformer) {
	var event map[string]interface{}
	synced := false
	mux := &sync.RWMutex{}
	_, err := resourceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}
			// Handler logic

			event = map[string]interface{}{
				"newObject": obj,
				"eventType": "ADDED",
			}
			go resourceInformerLog(event)

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}
			// Handler logic

			event = map[string]interface{}{
				"oldObject": oldObj,
				"newObject": newObj,
				"eventType": "MODIFIED",
			}
			go resourceInformerLog(event)

		},
		DeleteFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			// Handler logic

			event = map[string]interface{}{
				"newObject": obj,
				"eventType": "DELETED",
			}
			go resourceInformerLog(event)

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

	if !isSynced {
		log.Fatal("Informer event handler failed to sync.")
	}

	<-ctx.Done()

}
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
	for resourceType, resourceGroup := range resourceAPIList {
		resourceIndex = resourceIndex + 1

		resourceAPI := fmt.Sprintf("%s/v1/%s", resourceGroup, resourceType)

		common.SendLog(fmt.Sprintf("Attempting to create informer for resource API: '%s'", resourceAPI))
		resourceInformer := createResourceInformer(resourceGroup, resourceType, common.DynamicClient)
		if resourceInformer != nil {
			common.SendLog(fmt.Sprintf("Attempting to add event handler to informer for resource API: '%s'", resourceAPI))
			eventHandlerSync.Add(resourceIndex)
			go addInformerEventHandler(resourceInformer)
			{
				defer eventHandlerSync.Done()
				common.SendLog(fmt.Sprintf("Finished adding event handler to informer for resource API: '%s'", resourceAPI))
			}
		} else {
			common.SendLog(fmt.Sprintf("Failed to create informer for resource API: '%s'", resourceAPI))
		}
	}

	eventHandlerSync.Wait()
}

func resourceInformerLog(event map[string]interface{}) {
	var msg string
	if reflect.ValueOf(event).IsValid() {
		logEvent := &common.LogEvent{}
		jsonString, _ := json.Marshal(event)
		json.Unmarshal(jsonString, logEvent)
		eventType := event["eventType"].(string)
		newRawObjUnstructured := &unstructured.Unstructured{}
		newRawObjUnstructured.Object = logEvent.NewObject
		newResourceObj := common.KubernetesEvent{}
		unstructuredObjectJSON, err := newRawObjUnstructured.MarshalJSON()
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(unstructuredObjectJSON, &newResourceObj)
		if err != nil {
			log.Printf("[ERROR] Failed to parse resource event logs.\nERROR:\n%v", err)
		} else {
			resourceKind := newResourceObj.Kind
			resourceName := newResourceObj.KubernetesMetadata.Name
			resourceNamespace := newResourceObj.KubernetesMetadata.Namespace
			newResourceVersion := newResourceObj.KubernetesMetadata.ResourceVersion
			msg = common.ParseEventMessage(eventType, resourceName, resourceKind, resourceNamespace, newResourceVersion)
			if eventType == "MODIFIED" {
				oldRawObjUnstructured := &unstructured.Unstructured{}
				oldRawObjUnstructured.Object = logEvent.OldObject
				oldResourceObj := common.KubernetesEvent{}
				unstructuredObjectJSON, err = oldRawObjUnstructured.MarshalJSON()
				if err != nil {
					fmt.Println(err)
				}
				json.Unmarshal(unstructuredObjectJSON, &oldResourceObj)
				if err == nil {
					oldResourceName := oldResourceObj.KubernetesMetadata.Name
					oldResourceNamespace := oldResourceObj.KubernetesMetadata.Namespace
					oldResourceVersion := oldResourceObj.KubernetesMetadata.ResourceVersion
					msg = common.ParseEventMessage(eventType, oldResourceName, resourceKind, oldResourceNamespace, newResourceVersion, oldResourceVersion)
				}
			}
			clusterRelatedResources := GetClusterRelatedResources(resourceKind, resourceName, resourceNamespace)

			if reflect.ValueOf(clusterRelatedResources).IsValid() {
				event["relatedClusterServices"] = clusterRelatedResources
			}
		}
		marshaledEvent, err := json.Marshal(event)
		if err != nil {
			log.Printf("[ERROR] Failed to marshel resource event logs.\nERROR:\n%v", err)
		}
		common.SendLog(msg, marshaledEvent)

	}

}
