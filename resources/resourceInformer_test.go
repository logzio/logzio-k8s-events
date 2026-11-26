package resources

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
	"main.go/common"
	"reflect"
	"strings"
	"testing"
	"time"
)

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

func TestRunInformersWaitsForEventsInProgressOnShutdown(t *testing.T) {
	started, release := make(chan struct{}, 100), make(chan struct{})
	t.Cleanup(func() { handleEvent = func(event map[string]interface{}) { StructResourceLog(event) } })
	handleEvent = func(map[string]interface{}) {
		started <- struct{}{}
		<-release
	}

	gvr := deploymentsGVR
	listKinds := map[schema.GroupVersionResource]string{podsGVR: "PodList"}
	for _, shipped := range shippedResources {
		listKinds[shipped] = strings.TrimSuffix(strings.ToUpper(shipped.Resource[:1])+shipped.Resource[1:], "s") + "List"
	}
	client := fakeDynamic.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds)
	previous := caches
	t.Cleanup(func() { caches = previous })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		Run(ctx, client)
		close(stopped)
	}()

	// Events are held until the caches synced, so create Deployments until one is handled
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
		t.Fatal("shutdown must wait for the event in progress")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("informers did not stop after the event in progress finished")
	}
}

func TestProcessEventStampsTheEventTime(t *testing.T) {
	handled := make(chan map[string]interface{}, 1)
	t.Cleanup(func() { handleEvent = func(event map[string]interface{}) { StructResourceLog(event) } })
	handleEvent = func(event map[string]interface{}) { handled <- event }

	before := time.Now().UTC().Truncate(time.Millisecond)
	synced := make(chan struct{})
	close(synced)
	processEvent(map[string]interface{}{"eventType": common.EventTypeAdded}, synced, nil)
	event := <-handled
	inFlightEvents.Wait()

	stamp, err := time.Parse(eventTimestampFormat, event["@timestamp"].(string))
	if err != nil || stamp.Before(before) || stamp.After(time.Now().UTC()) {
		t.Errorf("expected @timestamp to be the time the event was received, got %v (%v)", event["@timestamp"], err)
	}
}

func TestStructResourceLogShipsTheEventTimestamp(t *testing.T) {
	_, parsedEvent := StructResourceLog(map[string]interface{}{
		"eventType":  common.EventTypeAdded,
		"@timestamp": "2026-11-18T10:00:00.000Z",
		"newObject": &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]interface{}{"name": "c", "namespace": "default"},
		}},
	})

	var shipped map[string]interface{}
	if err := json.Unmarshal([]byte(common.ParseEventLog("msg", parsedEvent)), &shipped); err != nil {
		t.Fatal(err)
	}
	if shipped["@timestamp"] != "2026-11-18T10:00:00.000Z" {
		t.Errorf("expected @timestamp in the shipped log, got %v", shipped["@timestamp"])
	}
}

// A Pod created after startup is stripped, indexed and found by an event about its Secret, through the real informers
func TestSecretEventFindsAPodThroughTheInformerCaches(t *testing.T) {
	related := make(chan common.RelatedClusterServices, 10)
	t.Cleanup(func() { handleEvent = func(event map[string]interface{}) { StructResourceLog(event) } })
	handleEvent = func(event map[string]interface{}) {
		if object, ok := event["newObject"].(*unstructured.Unstructured); ok && object.GetKind() == "Secret" {
			related <- GetClusterRelatedResources(object.Object)
		}
	}
	listKinds := map[schema.GroupVersionResource]string{podsGVR: "PodList"}
	for _, shipped := range shippedResources {
		listKinds[shipped] = strings.TrimSuffix(strings.ToUpper(shipped.Resource[:1])+shipped.Resource[1:], "s") + "List"
	}
	client := fakeDynamic.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds)
	previous := caches
	t.Cleanup(func() { caches = previous })
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go func() {
		Run(ctx, client)
		close(stopped)
	}()
	t.Cleanup(func() { cancel(); <-stopped })

	pod := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1", "kind": "Pod",
		"metadata": map[string]interface{}{"name": "app-pod", "namespace": "team-a"},
		"spec": map[string]interface{}{"containers": []interface{}{map[string]interface{}{
			"name": "app", "image": "app:1",
			"envFrom": []interface{}{map[string]interface{}{"secretRef": map[string]interface{}{"name": "db"}}},
		}}},
	}}
	if _, err := client.Resource(podsGVR).Namespace("team-a").Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	// The Pod reaches the cache asynchronously, so update the Secret until its event finds the Pod
	secretsGVR := schema.GroupVersionResource{Version: "v1", Resource: "secrets"}
	deadline := time.After(5 * time.Second)
	for i := 0; ; i++ {
		secret := &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "v1", "kind": "Secret",
			"metadata": map[string]interface{}{"name": "db", "namespace": "team-a"},
			"data":     map[string]interface{}{"n": fmt.Sprint(i)},
		}}
		var err error
		if i == 0 {
			_, err = client.Resource(secretsGVR).Namespace("team-a").Create(ctx, secret, metav1.CreateOptions{})
		} else {
			_, err = client.Resource(secretsGVR).Namespace("team-a").Update(ctx, secret, metav1.UpdateOptions{})
		}
		if err != nil {
			t.Fatal(err)
		}
		select {
		case got := <-related:
			if len(got.Pods) == 1 && got.Pods[0] == "app-pod" {
				return
			}
		case <-time.After(50 * time.Millisecond):
		case <-deadline:
			t.Fatal("no Secret event found the Pod")
		}
	}
}

// Shipped kinds are cached without managedFields, so the handlers never see them
func TestShippedObjectsAreCachedWithoutManagedFields(t *testing.T) {
	received := make(chan *unstructured.Unstructured, 10)
	t.Cleanup(func() { handleEvent = func(event map[string]interface{}) { StructResourceLog(event) } })
	handleEvent = func(event map[string]interface{}) {
		if object, ok := event["newObject"].(*unstructured.Unstructured); ok {
			received <- object
		}
	}
	listKinds := map[schema.GroupVersionResource]string{podsGVR: "PodList"}
	for _, shipped := range shippedResources {
		listKinds[shipped] = strings.TrimSuffix(strings.ToUpper(shipped.Resource[:1])+shipped.Resource[1:], "s") + "List"
	}
	client := fakeDynamic.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds)
	previous := caches
	t.Cleanup(func() { caches = previous })
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go func() {
		Run(ctx, client)
		close(stopped)
	}()
	t.Cleanup(func() { cancel(); <-stopped })

	configMapsGVR := schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
	deadline := time.After(5 * time.Second)
	for i := 0; ; i++ {
		configMap := &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "v1", "kind": "ConfigMap",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("cm-%d", i), "namespace": "team-a",
				"managedFields": []interface{}{map[string]interface{}{"manager": "kubectl", "operation": "Apply"}},
			},
		}}
		if _, err := client.Resource(configMapsGVR).Namespace("team-a").Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
		select {
		case object := <-received:
			if _, found, _ := unstructured.NestedFieldNoCopy(object.Object, "metadata", "managedFields"); found {
				t.Fatal("the handler got an object with managedFields")
			}
			if object.GetName() == "" {
				t.Fatal("the transform must keep the rest of the object")
			}
			return
		case <-time.After(50 * time.Millisecond):
		case <-deadline:
			t.Fatal("no event was handled")
		}
	}
}

func TestStripManagedFieldsIsIdempotent(t *testing.T) {
	object := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"name": "a", "managedFields": []interface{}{map[string]interface{}{"manager": "m"}}},
		"data":     map[string]interface{}{"k": "v"},
	}}
	once, _ := stripManagedFields(object.DeepCopy())
	twice, _ := stripManagedFields(once.(*unstructured.Unstructured).DeepCopy())
	want := map[string]interface{}{"metadata": map[string]interface{}{"name": "a"}, "data": map[string]interface{}{"k": "v"}}
	if !reflect.DeepEqual(once.(*unstructured.Unstructured).Object, want) || !reflect.DeepEqual(twice, once) {
		t.Errorf("unexpected result: once %v, twice %v", once, twice)
	}
	if tombstone := (cache.DeletedFinalStateUnknown{Key: "k"}); !reflect.DeepEqual(func() interface{} { o, _ := stripManagedFields(tombstone); return o }(), tombstone) {
		t.Error("a non-unstructured object must pass through unchanged")
	}
}
