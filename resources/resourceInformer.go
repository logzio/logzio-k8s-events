package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"log"
	"main.go/common"
	"sync"
	"time"
)

// deleteInternalFields deletes internal fields from a Kubernetes object.
// managedFields are already stripped when the object is cached.
func deleteInternalFields(obj *unstructured.Unstructured) {
	if meta, ok := obj.Object[common.Metadata].(map[string]interface{}); ok {
		delete(meta, common.ResourceVersion)
		if annotations, ok := meta[common.Annotations].(map[string]interface{}); ok {
			delete(annotations, common.DeploymentRevision)
			// An empty annotations map compares equal to none
			if len(annotations) == 0 {
				delete(meta, common.Annotations)
			}
		}
	}
	delete(obj.Object, common.Status)
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
		newJSON, _ := json.Marshal(newCopy)
		oldJSON, _ := json.Marshal(oldCopy)

		return string(oldJSON) == string(newJSON)
	}

	return false
}

const eventTimestampFormat = "2006-01-02T15:04:05.000Z07:00"

// cacheSyncWarning is how long the caches may take to sync before a warning is logged
const cacheSyncWarning = 2 * time.Minute

// shippedResources are the resources whose events are shipped. Pods are only cached, to find related workloads.
var shippedResources = []schema.GroupVersionResource{
	{Version: "v1", Resource: "configmaps"},
	{Version: "v1", Resource: "secrets"},
	{Version: "v1", Resource: "serviceaccounts"},
	deploymentsGVR,
	daemonSetsGVR,
	statefulSetsGVR,
	{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
	clusterRoleBindingsGVR,
}

// inFlightEvents tracks events still being processed, so shutdown waits for them before the final flush
var inFlightEvents sync.WaitGroup

// handleEvent is replaced in tests to hold an event in progress
var handleEvent = func(event map[string]interface{}) { StructResourceLog(event) }

// eventGate admits events until shutdown, and holds them until every cache synced, so their related resources are complete
type eventGate struct {
	mux     sync.RWMutex
	open    bool
	synced  chan struct{}
	stopped <-chan struct{}
}

func newEventGate(stopped <-chan struct{}) *eventGate {
	return &eventGate{open: true, synced: make(chan struct{}), stopped: stopped}
}

// admit processes the event, unless the gate was closed for shutdown
func (g *eventGate) admit(event map[string]interface{}) {
	g.mux.RLock()
	defer g.mux.RUnlock()
	if g.open {
		processEvent(event, g.synced, g.stopped)
	}
}

// close stops admitting events. Lock waits for running handlers, so no event starts after this
func (g *eventGate) close() {
	g.mux.Lock()
	g.open = false
	g.mux.Unlock()
}

// processEvent stamps the event and handles it in the background once the caches synced.
// An event still waiting for the caches at shutdown is dropped.
func processEvent(event map[string]interface{}, synced, stopped <-chan struct{}) {
	// Logz.io would otherwise stamp the event when it arrives, after lookups and batching
	event["@timestamp"] = time.Now().UTC().Format(eventTimestampFormat)
	inFlightEvents.Add(1)
	go func() {
		defer inFlightEvents.Done()
		select {
		case <-synced:
		default:
			select {
			case <-synced:
			case <-stopped:
				return
			}
		}
		handleEvent(event)
	}()
}

// addEventHandler sends the informer's add, update and delete events through the gate
func addEventHandler(resourceInformer cache.SharedIndexInformer, gate *eventGate) error {
	_, err := resourceInformer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			// The initial list is the cluster's state at startup, not a change
			if isInInitialList {
				return
			}
			gate.admit(map[string]interface{}{
				"newObject": obj,
				"eventType": common.EventTypeAdded,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if IgnoreInternalChanges(oldObj, newObj) {
				return // ignore internal cluster updates
			}
			gate.admit(map[string]interface{}{
				"oldObject": oldObj,
				"newObject": newObj,
				"eventType": common.EventTypeModified,
			})
		},
		DeleteFunc: func(obj interface{}) {
			gate.admit(map[string]interface{}{
				"newObject": unwrapDeletedObject(obj),
				"eventType": common.EventTypeDeleted,
			})
		},
	})
	return err
}

// unwrapDeletedObject returns the deleted object itself when the informer missed the delete and passed a tombstone
func unwrapDeletedObject(obj interface{}) interface{} {
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		return tombstone.Obj
	}
	return obj
}

// Run watches the shipped resources and ships their events, from one shared informer factory that also caches
// the Pods related workloads are looked up in. It blocks until ctx is done and every event in progress was
// handed to the sender.
func Run(ctx context.Context, client dynamic.Interface) {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(client, 0)
	relations, err := newRelationCaches(factory)
	if err != nil {
		log.Fatalf("[FATAL] Failed to index the informer caches.\nERROR:\n%v", err)
	}
	caches = relations

	gate := newEventGate(ctx.Done())
	for _, gvr := range shippedResources {
		log.Printf("Adding event handler for resource API: '%s'", gvr)
		informer := factory.ForResource(gvr).Informer()
		if err := informer.SetTransform(stripManagedFields); err != nil {
			log.Fatalf("[FATAL] Failed to set the cache transform for resource API: '%s'.\nERROR:\n%v", gvr, err)
		}
		if err := addEventHandler(informer, gate); err != nil {
			common.SendLog(fmt.Sprintf("[ERROR] Failed to add event handler for resource API: '%s'.\nERROR:\n%v", gvr, err))
		}
	}
	factory.Start(ctx.Done())

	// A cache that never syncs, e.g. without permission to watch its resource, holds every event, so say which one
	warnCtx, cancelWarning := context.WithTimeout(ctx, cacheSyncWarning)
	for gvr, synced := range factory.WaitForCacheSync(warnCtx.Done()) {
		if !synced && ctx.Err() == nil {
			common.SendLog(fmt.Sprintf("[WARN] The cache of '%s' did not sync within %s. Events wait until it does.", gvr, cacheSyncWarning))
		}
	}
	cancelWarning()
	factory.WaitForCacheSync(ctx.Done())
	if ctx.Err() == nil {
		close(gate.synced)
		log.Printf("Informer caches synced.")
	}

	// Wait for the process to be stopped (SIGINT or SIGTERM)
	<-ctx.Done()
	gate.close()
	factory.Shutdown()
	inFlightEvents.Wait()
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
	newObject := &unstructured.Unstructured{Object: logEvent.NewObject}
	if eventType == common.EventTypeModified {
		oldObject := &unstructured.Unstructured{Object: logEvent.OldObject}
		msg = common.ParseEventMessage(eventType, oldObject.GetName(), newObject.GetKind(), oldObject.GetNamespace(), newObject.GetResourceVersion(), oldObject.GetResourceVersion())
	} else {
		msg = common.ParseEventMessage(eventType, newObject.GetName(), newObject.GetKind(), newObject.GetNamespace(), newObject.GetResourceVersion())
	}

	event["relatedClusterServices"] = GetClusterRelatedResources(logEvent.NewObject)

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
