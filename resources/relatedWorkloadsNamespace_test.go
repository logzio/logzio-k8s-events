package resources

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"main.go/common"
)

// deploymentWith mounts a Secret and a ConfigMap that are both named secret
func deploymentWith(name, namespace, serviceAccount, secret string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccount,
			Volumes: []corev1.Volume{
				{Name: "s", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: secret}}},
				{Name: "c", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: secret}}}},
			},
		}}},
	}
}

func clusterRoleBinding(name, role string, subjects ...rbacv1.Subject) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: name}, RoleRef: rbacv1.RoleRef{Name: role}, Subjects: subjects}
}

func serviceAccountSubject(name, namespace string) rbacv1.Subject {
	return rbacv1.Subject{Kind: "ServiceAccount", Name: name, Namespace: namespace}
}

// toUnstructured converts a typed object to the unstructured form the informers cache and deliver
func toUnstructured(t *testing.T, obj runtime.Object) *unstructured.Unstructured {
	t.Helper()
	kinds := map[reflect.Type][2]string{
		reflect.TypeOf(&corev1.Pod{}):                {"v1", "Pod"},
		reflect.TypeOf(&corev1.Secret{}):             {"v1", "Secret"},
		reflect.TypeOf(&corev1.ConfigMap{}):          {"v1", "ConfigMap"},
		reflect.TypeOf(&corev1.ServiceAccount{}):     {"v1", "ServiceAccount"},
		reflect.TypeOf(&appsv1.Deployment{}):         {"apps/v1", "Deployment"},
		reflect.TypeOf(&appsv1.DaemonSet{}):          {"apps/v1", "DaemonSet"},
		reflect.TypeOf(&appsv1.StatefulSet{}):        {"apps/v1", "StatefulSet"},
		reflect.TypeOf(&rbacv1.ClusterRole{}):        {"rbac.authorization.k8s.io/v1", "ClusterRole"},
		reflect.TypeOf(&rbacv1.ClusterRoleBinding{}): {"rbac.authorization.k8s.io/v1", "ClusterRoleBinding"},
	}
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		t.Fatal(err)
	}
	converted := &unstructured.Unstructured{Object: object}
	kind := kinds[reflect.TypeOf(obj)]
	converted.SetAPIVersion(kind[0])
	converted.SetKind(kind[1])
	return converted
}

// eventObject is the object of an event about obj
func eventObject(t *testing.T, obj runtime.Object) map[string]interface{} {
	return toUnstructured(t, obj).Object
}

// useFakeCaches fills indexed caches with the objects, as the informers would
func useFakeCaches(t *testing.T, objects ...runtime.Object) {
	t.Helper()
	previous := caches
	t.Cleanup(func() { caches = previous })
	indexer := func(indexers cache.Indexers) cache.Indexer {
		return cache.NewIndexer(cache.MetaNamespaceKeyFunc, indexers)
	}
	caches = relationCaches{
		pods:                indexer(workloadIndexers("Pod")),
		deployments:         indexer(workloadIndexers("Deployment")),
		daemonSets:          indexer(workloadIndexers("DaemonSet")),
		statefulSets:        indexer(workloadIndexers("StatefulSet")),
		clusterRoleBindings: indexer(clusterRoleBindingIndexers()),
	}
	for _, obj := range objects {
		object := toUnstructured(t, obj)
		var err error
		switch object.GetKind() {
		case "Pod":
			stripped, _ := stripPod(object)
			err = caches.pods.Add(stripped)
		case "Deployment":
			err = caches.deployments.Add(object)
		case "DaemonSet":
			err = caches.daemonSets.Add(object)
		case "StatefulSet":
			err = caches.statefulSets.Add(object)
		case "ClusterRoleBinding":
			err = caches.clusterRoleBindings.Add(object)
		default:
			t.Fatalf("no cache for kind %s", object.GetKind())
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRelatedWorkloadsAreLimitedToTheResourceNamespace(t *testing.T) {
	useFakeCaches(t,
		deploymentWith("app-a", "team-a", "default", "db"),
		deploymentWith("app-b", "team-b", "default", "db"),
	)
	meta := metav1.ObjectMeta{Name: "db", Namespace: "team-a"}

	for kind, related := range map[string]common.RelatedClusterServices{
		"Secret":         GetClusterRelatedResources(eventObject(t, &corev1.Secret{ObjectMeta: meta})),
		"ConfigMap":      GetClusterRelatedResources(eventObject(t, &corev1.ConfigMap{ObjectMeta: meta})),
		"ServiceAccount": GetClusterRelatedResources(eventObject(t, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "team-a"}})),
	} {
		if !reflect.DeepEqual(related.Deployments, []string{"app-a"}) {
			t.Errorf("%s in team-a: expected only app-a, got %v", kind, related.Deployments)
		}
	}
}

func TestClusterRoleRelatedWorkloadsMergesAllSubjects(t *testing.T) {
	readers := clusterRoleBinding("readers", "view", serviceAccountSubject("reader", "team-a"), serviceAccountSubject("reader", "team-b"))
	useFakeCaches(t,
		deploymentWith("app-a", "team-a", "reader", ""),
		deploymentWith("app-b", "team-b", "reader", ""),
		deploymentWith("app-c", "team-c", "reader", ""),
		readers,
		clusterRoleBinding("other", "edit", serviceAccountSubject("reader", "team-c")),
	)

	for kind, related := range map[string]common.RelatedClusterServices{
		"ClusterRole":        GetClusterRelatedResources(eventObject(t, &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "view"}})),
		"ClusterRoleBinding": GetClusterRelatedResources(eventObject(t, readers)),
	} {
		if !reflect.DeepEqual(related.Deployments, []string{"app-a", "app-b"}) {
			t.Errorf("%s: expected the workloads of both subjects and not team-c's, got %v", kind, related.Deployments)
		}
	}
}

func TestWorkloadRelatedClusterRolesMatchTheServiceAccountNamespace(t *testing.T) {
	appA := deploymentWith("app-a", "team-a", "default", "")
	useFakeCaches(t,
		appA,
		clusterRoleBinding("team-a-view", "view", serviceAccountSubject("default", "team-a")),
		clusterRoleBinding("team-b-admin", "admin", serviceAccountSubject("default", "team-b")),
	)

	related := GetClusterRelatedResources(eventObject(t, appA))

	if !reflect.DeepEqual(related.ClusterRoleBindings, []string{"team-a-view"}) || !reflect.DeepEqual(related.ClusterRoles, []string{"view"}) {
		t.Errorf("expected only team-a's binding and role, got bindings %v and roles %v", related.ClusterRoleBindings, related.ClusterRoles)
	}
}

// A deleted object is no longer in the cluster, but its event still carries what it referenced
func TestDeletedObjectsReportWhatTheyWereRelatedTo(t *testing.T) {
	deleted := deploymentWith("deleted-app", "team-a", "reader", "db")
	deletedBinding := clusterRoleBinding("deleted", "view", serviceAccountSubject("reader", "team-a"))
	useFakeCaches(t, deploymentWith("app-a", "team-a", "reader", ""))

	workload := GetClusterRelatedResources(eventObject(t, deleted))
	if !reflect.DeepEqual(workload.Secrets, []string{"db"}) || !reflect.DeepEqual(workload.ConfigMaps, []string{"db"}) {
		t.Errorf("expected the deleted Deployment's Secret and ConfigMap, got %+v", workload)
	}
	binding := GetClusterRelatedResources(eventObject(t, deletedBinding))
	if !reflect.DeepEqual(binding.Deployments, []string{"app-a"}) {
		t.Errorf("expected the workloads of the deleted binding's subject, got %+v", binding)
	}
}

func TestEnvFromAndImagePullSecretsAreRelated(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "app-pod", Namespace: "team-a"},
		Spec: corev1.PodSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: "registry"}},
			Containers: []corev1.Container{{
				Name: "app",
				EnvFrom: []corev1.EnvFromSource{
					{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "settings"}}},
					{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "credentials"}}},
				},
			}},
		},
	}
	useFakeCaches(t, pod)

	for _, check := range []struct {
		object runtime.Object
	}{
		{&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "settings", Namespace: "team-a"}}},
		{&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "credentials", Namespace: "team-a"}}},
		{&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "registry", Namespace: "team-a"}}},
	} {
		related := GetClusterRelatedResources(eventObject(t, check.object))
		if !reflect.DeepEqual(related.Pods, []string{"app-pod"}) {
			t.Errorf("%T %v: expected app-pod, got %+v", check.object, check.object.(metav1.Object).GetName(), related)
		}
	}
}

func TestStrippedPodKeepsOnlyReferences(t *testing.T) {
	pod := toUnstructured(t, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "app-pod", Namespace: "team-a", Labels: map[string]string{"app": "a"}},
		Spec: corev1.PodSpec{
			ServiceAccountName: "reader",
			NodeName:           "node-1",
			Containers: []corev1.Container{{
				Name:  "app",
				Image: "app:1",
				Env: []corev1.EnvVar{
					{Name: "PASSWORD", Value: "literal"},
					{Name: "TOKEN", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "db"}}}},
				},
			}},
			Volumes: []corev1.Volume{
				{Name: "data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "conf", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "conf"}}}},
			},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	})

	stripped, _ := stripPod(pod)
	again, _ := stripPod(stripped)
	if !reflect.DeepEqual(stripped, again) {
		t.Error("stripping a stripped Pod must not change it")
	}

	want := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata":   map[string]interface{}{"name": "app-pod", "namespace": "team-a"},
		"spec": map[string]interface{}{
			"serviceAccountName": "reader",
			"containers": []interface{}{map[string]interface{}{
				"name": "app",
				"env": []interface{}{map[string]interface{}{
					"name":      "TOKEN",
					"valueFrom": map[string]interface{}{"secretKeyRef": map[string]interface{}{"name": "db", "key": ""}},
				}},
			}},
			"volumes": []interface{}{
				map[string]interface{}{"name": "conf", "configMap": map[string]interface{}{"name": "conf"}},
			},
		},
	}
	if got := stripped.(*unstructured.Unstructured).Object; !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected stripped Pod:\n got %v\nwant %v", got, want)
	}
}
