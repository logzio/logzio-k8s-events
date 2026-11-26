package resources

import (
	"log"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"main.go/common"
)

// Index names. Workload caches are indexed by what they reference, ClusterRoleBindings by role and subject,
// so finding the resources related to an event is a map lookup instead of a scan of the cluster
const (
	configMapIndex             = "configmap"
	secretIndex                = "secret"
	serviceAccountIndex        = "serviceaccount"
	roleRefIndex               = "roleref"
	subjectServiceAccountIndex = "subject-serviceaccount"
)

var (
	podsGVR                = schema.GroupVersionResource{Version: "v1", Resource: "pods"}
	deploymentsGVR         = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	daemonSetsGVR          = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	statefulSetsGVR        = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	clusterRoleBindingsGVR = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}
)

// relationCaches holds the informer caches that related resources are looked up in
type relationCaches struct {
	pods, deployments, daemonSets, statefulSets, clusterRoleBindings cache.Indexer
}

// caches is set before the informers start, and replaced in tests
var caches relationCaches

// newRelationCaches adds the indexes to the factory's informers, and strips cached Pods to the fields the indexes read.
// It must run before the factory starts.
func newRelationCaches(factory dynamicinformer.DynamicSharedInformerFactory) (relationCaches, error) {
	podInformer := factory.ForResource(podsGVR).Informer()
	if err := podInformer.SetTransform(stripPod); err != nil {
		return relationCaches{}, err
	}
	indexed := map[schema.GroupVersionResource]cache.Indexers{
		podsGVR:                workloadIndexers("Pod"),
		deploymentsGVR:         workloadIndexers("Deployment"),
		daemonSetsGVR:          workloadIndexers("DaemonSet"),
		statefulSetsGVR:        workloadIndexers("StatefulSet"),
		clusterRoleBindingsGVR: clusterRoleBindingIndexers(),
	}
	indexers := map[schema.GroupVersionResource]cache.Indexer{}
	for gvr, gvrIndexers := range indexed {
		informer := factory.ForResource(gvr).Informer()
		if err := informer.AddIndexers(gvrIndexers); err != nil {
			return relationCaches{}, err
		}
		indexers[gvr] = informer.GetIndexer()
	}
	return relationCaches{
		pods:                indexers[podsGVR],
		deployments:         indexers[deploymentsGVR],
		daemonSets:          indexers[daemonSetsGVR],
		statefulSets:        indexers[statefulSetsGVR],
		clusterRoleBindings: indexers[clusterRoleBindingsGVR],
	}, nil
}

// fromUnstructured converts a cached object to its typed form
func fromUnstructured[T any](obj interface{}) (typed T, ok bool) {
	u, isUnstructured := obj.(*unstructured.Unstructured)
	if !isUnstructured {
		return typed, false
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &typed); err != nil {
		log.Printf("[ERROR] Failed to convert %s %s/%s: %v", u.GetKind(), u.GetNamespace(), u.GetName(), err)
		return typed, false
	}
	return typed, true
}

// workloadOf converts a cached or event object of the given kind to a Workload
func workloadOf(kind string, obj interface{}) (Workload, bool) {
	switch kind {
	case "Pod":
		pod, ok := fromUnstructured[corev1.Pod](obj)
		return Pod(pod), ok
	case "Deployment":
		deployment, ok := fromUnstructured[appsv1.Deployment](obj)
		return Deployment(deployment), ok
	case "DaemonSet":
		daemonSet, ok := fromUnstructured[appsv1.DaemonSet](obj)
		return DaemonSet(daemonSet), ok
	case "StatefulSet":
		statefulSet, ok := fromUnstructured[appsv1.StatefulSet](obj)
		return StatefulSet(statefulSet), ok
	}
	return nil, false
}

// namespacedKeys prefixes each name with the namespace, the form every namespaced index key takes
func namespacedKeys(namespace string, names []string) (keys []string) {
	for _, name := range names {
		keys = append(keys, namespace+"/"+name)
	}
	return keys
}

// workloadIndexers indexes workloads of a kind by the ConfigMaps, Secrets and ServiceAccount they reference.
// An index function must not fail: the cache panics on an index error, so an unconvertible object is left unindexed.
func workloadIndexers(kind string) cache.Indexers {
	referenced := func(references func(Workload) []string) cache.IndexFunc {
		return func(obj interface{}) ([]string, error) {
			workload, ok := workloadOf(kind, obj)
			if !ok {
				return nil, nil
			}
			return namespacedKeys(workload.GetNamespace(), references(workload)), nil
		}
	}
	return cache.Indexers{
		configMapIndex:      referenced(GetWorkloadRelatedConfigMaps),
		secretIndex:         referenced(GetWorkloadRelatedSecrets),
		serviceAccountIndex: referenced(GetWorkloadRelatedServiceAccounts),
	}
}

// clusterRoleBindingIndexers indexes ClusterRoleBindings by their ClusterRole and by their ServiceAccount subjects
func clusterRoleBindingIndexers() cache.Indexers {
	return cache.Indexers{
		roleRefIndex: func(obj interface{}) ([]string, error) {
			binding, ok := fromUnstructured[rbacv1.ClusterRoleBinding](obj)
			if !ok || binding.RoleRef.Name == "" {
				return nil, nil
			}
			return []string{binding.RoleRef.Name}, nil
		},
		subjectServiceAccountIndex: func(obj interface{}) (keys []string, err error) {
			binding, ok := fromUnstructured[rbacv1.ClusterRoleBinding](obj)
			if !ok {
				return nil, nil
			}
			for _, subject := range binding.Subjects {
				if subject.Kind == "ServiceAccount" {
					keys = append(keys, subject.Namespace+"/"+subject.Name)
				}
			}
			return keys, nil
		},
	}
}

// stripManagedFields drops metadata.managedFields before an object is cached. It's server-side apply bookkeeping
// that is never shipped, and a third of a cached Deployment. The handlers get the cached copy, so it must stay
// limited to fields the events don't carry. It's idempotent, as a transform must be.
func stripManagedFields(obj interface{}) (interface{}, error) {
	if object, ok := obj.(*unstructured.Unstructured); ok {
		object.SetManagedFields(nil)
	}
	return obj, nil
}

// stripPod keeps only what the indexes read of a Pod, so the cache holds a fraction of every Pod in the cluster.
// Literal env values are dropped too, so the cache never holds them. It's idempotent, as a transform must be.
func stripPod(obj interface{}) (interface{}, error) {
	pod, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return obj, nil
	}
	podMetadata, _ := pod.Object[common.Metadata].(map[string]interface{})
	metadata := map[string]interface{}{}
	for _, field := range []string{"name", "namespace", "uid", "resourceVersion", "ownerReferences"} {
		if value, found := podMetadata[field]; found {
			metadata[field] = value
		}
	}
	spec, _ := pod.Object["spec"].(map[string]interface{})
	strippedSpec := map[string]interface{}{}
	for _, field := range []string{"serviceAccountName", "imagePullSecrets"} {
		if value, found := spec[field]; found {
			strippedSpec[field] = value
		}
	}
	if containers, found := spec["containers"].([]interface{}); found {
		strippedSpec["containers"] = stripContainers(containers)
	}
	var volumes []interface{}
	for _, volume := range keepFields(asList(spec["volumes"]), "name", "configMap", "secret") {
		if len(volume.(map[string]interface{})) > 1 {
			volumes = append(volumes, volume) // a volume with a ConfigMap or Secret source
		}
	}
	if volumes != nil {
		strippedSpec["volumes"] = volumes
	}
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion":    pod.GetAPIVersion(),
		"kind":          pod.GetKind(),
		common.Metadata: metadata,
		"spec":          strippedSpec,
	}}, nil
}

// stripContainers keeps each container's name, envFrom, and the env vars that reference another resource
func stripContainers(containers []interface{}) (stripped []interface{}) {
	for _, item := range containers {
		container, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		strippedContainer := map[string]interface{}{"name": container["name"]}
		if envFrom, found := container["envFrom"]; found {
			strippedContainer["envFrom"] = envFrom
		}
		var env []interface{}
		for _, envItem := range keepFields(asList(container["env"]), "name", "valueFrom") {
			if _, references := envItem.(map[string]interface{})["valueFrom"]; references {
				env = append(env, envItem)
			}
		}
		if env != nil {
			strippedContainer["env"] = env
		}
		stripped = append(stripped, strippedContainer)
	}
	return stripped
}

// keepFields copies each map in the list with only the given fields
func keepFields(list []interface{}, fields ...string) (kept []interface{}) {
	for _, item := range list {
		object, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		keptObject := map[string]interface{}{}
		for _, field := range fields {
			if value, found := object[field]; found {
				keptObject[field] = value
			}
		}
		kept = append(kept, keptObject)
	}
	return kept
}

func asList(value interface{}) []interface{} {
	list, _ := value.([]interface{})
	return list
}

// indexedNames returns the sorted names of the cached objects with the key in the index
func indexedNames(indexer cache.Indexer, index, key string) (names []string) {
	if indexer == nil {
		return nil
	}
	objects, err := indexer.ByIndex(index, key)
	if err != nil {
		log.Printf("[ERROR] Failed to look up %s in index %s: %v", key, index, err)
		return nil
	}
	for _, obj := range objects {
		if object, ok := obj.(*unstructured.Unstructured); ok {
			names = append(names, object.GetName())
		}
	}
	sort.Strings(names)
	return names
}

// referencingWorkloads returns the workloads in the namespace that reference the named resource through the index
func referencingWorkloads(index, namespace, name string) common.RelatedClusterServices {
	key := namespace + "/" + name
	return common.RelatedClusterServices{
		Deployments:  indexedNames(caches.deployments, index, key),
		DaemonSets:   indexedNames(caches.daemonSets, index, key),
		StatefulSets: indexedNames(caches.statefulSets, index, key),
		Pods:         indexedNames(caches.pods, index, key),
	}
}

// serviceAccountBindings returns the ClusterRoleBindings with the ServiceAccount as a subject, and their ClusterRoles
func serviceAccountBindings(namespace, serviceAccount string) (bindings, roles []string) {
	if serviceAccount == "" || caches.clusterRoleBindings == nil {
		return nil, nil
	}
	objects, err := caches.clusterRoleBindings.ByIndex(subjectServiceAccountIndex, namespace+"/"+serviceAccount)
	if err != nil {
		log.Printf("[ERROR] Failed to look up the ClusterRoleBindings of ServiceAccount %s/%s: %v", namespace, serviceAccount, err)
		return nil, nil
	}
	for _, obj := range objects {
		if binding, ok := fromUnstructured[rbacv1.ClusterRoleBinding](obj); ok {
			bindings = appendMissing(bindings, []string{binding.Name})
			roles = appendMissing(roles, []string{binding.RoleRef.Name})
		}
	}
	sort.Strings(bindings)
	sort.Strings(roles)
	return bindings, roles
}

// workloadRelatedResources returns what the workload references, read from the workload itself
func workloadRelatedResources(workload Workload) common.RelatedClusterServices {
	bindings, roles := serviceAccountBindings(workload.GetNamespace(), workload.GetServiceAccountName())
	return common.RelatedClusterServices{
		ConfigMaps:          GetWorkloadRelatedConfigMaps(workload),
		Secrets:             GetWorkloadRelatedSecrets(workload),
		ServiceAccounts:     GetWorkloadRelatedServiceAccounts(workload),
		ClusterRoleBindings: bindings,
		ClusterRoles:        roles,
	}
}

// GetClusterRelatedResources returns the resources related to the event's object.
// Everything is read from the object itself or from the informer caches, never from the API server,
// so a deleted object still reports what it was related to.
func GetClusterRelatedResources(object map[string]interface{}) (relatedClusterServices common.RelatedClusterServices) {
	event := &unstructured.Unstructured{Object: object}
	kind, name, namespace := event.GetKind(), event.GetName(), event.GetNamespace()
	switch kind {
	case "ConfigMap":
		return referencingWorkloads(configMapIndex, namespace, name)
	case "Secret":
		return referencingWorkloads(secretIndex, namespace, name)
	case "ServiceAccount":
		return referencingWorkloads(serviceAccountIndex, namespace, name)
	case "ClusterRoleBinding":
		if binding, ok := fromUnstructured[rbacv1.ClusterRoleBinding](event); ok {
			return subjectsRelatedWorkloads(binding.Subjects)
		}
	case "ClusterRole":
		if caches.clusterRoleBindings == nil {
			return relatedClusterServices
		}
		bindings, err := caches.clusterRoleBindings.ByIndex(roleRefIndex, name)
		if err != nil {
			log.Printf("[ERROR] Failed to look up the ClusterRoleBindings of ClusterRole %s: %v", name, err)
			return relatedClusterServices
		}
		for _, obj := range bindings {
			if binding, ok := fromUnstructured[rbacv1.ClusterRoleBinding](obj); ok {
				relatedClusterServices = mergeRelatedWorkloads(relatedClusterServices, subjectsRelatedWorkloads(binding.Subjects))
			}
		}
	case "Deployment", "DaemonSet", "StatefulSet":
		if workload, ok := workloadOf(kind, event); ok {
			return workloadRelatedResources(workload)
		}
	default:
		log.Printf("[ERROR] Unknown resource kind %s", kind)
	}
	return relatedClusterServices
}
