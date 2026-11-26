package resources

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"main.go/common"
)

// useTestFixtures caches every test workload and cluster role binding
func useTestFixtures(t *testing.T) {
	objects := make([]runtime.Object, 0, 9)
	for _, pod := range GetTestPods() {
		objects = append(objects, &pod)
	}
	for _, deployment := range GetTestDeployments() {
		objects = append(objects, &deployment)
	}
	for _, daemonSet := range GetTestDaemonSets() {
		objects = append(objects, &daemonSet)
	}
	for _, statefulSet := range GetTestStatefulSets() {
		objects = append(objects, &statefulSet)
	}
	for _, binding := range GetTestClusterRoleBindings() {
		objects = append(objects, &binding)
	}
	useFakeCaches(t, objects...)
}

// allTestWorkloads are the test workloads that all reference the same Secret, ConfigMap and ServiceAccount
var allTestWorkloads = common.RelatedClusterServices{
	Deployments:  []string{"test-deployment", "test-deployment-2"},
	DaemonSets:   []string{"test-daemonset", "test-daemonset-2"},
	StatefulSets: []string{"test-statefulset", "test-statefulset-2"},
	Pods:         []string{"test-pod", "test-pod-2"},
}

func TestReferencedResourcesRelateToEveryWorkload(t *testing.T) {
	useTestFixtures(t)
	inDefault := func(name string) metav1.ObjectMeta { return metav1.ObjectMeta{Name: name, Namespace: "default"} }

	for kind, object := range map[string]runtime.Object{
		"Secret":             &corev1.Secret{ObjectMeta: inDefault("test-secret")},
		"ConfigMap":          &corev1.ConfigMap{ObjectMeta: inDefault("test-configmap")},
		"ServiceAccount":     &corev1.ServiceAccount{ObjectMeta: inDefault("test-serviceaccount")},
		"ClusterRole":        &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "test-clusterrole"}},
		"ClusterRoleBinding": func() runtime.Object { b := GetTestClusterRoleBinding("test-clusterrolebinding"); return &b }(),
	} {
		if related := GetClusterRelatedResources(eventObject(t, object)); !reflect.DeepEqual(related, allTestWorkloads) {
			t.Errorf("%s: expected %+v, got %+v", kind, allTestWorkloads, related)
		}
	}
}

func TestUnreferencedResourceHasNoRelatedWorkloads(t *testing.T) {
	useTestFixtures(t)

	related := GetClusterRelatedResources(eventObject(t, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "unused", Namespace: "default"}}))

	if !reflect.DeepEqual(related, common.RelatedClusterServices{}) {
		t.Errorf("expected no related workloads, got %+v", related)
	}
}

func TestWorkloadEventsReportTheirReferencesAndBindings(t *testing.T) {
	useTestFixtures(t)
	want := common.RelatedClusterServices{
		ConfigMaps:          []string{"test-configmap"},
		Secrets:             []string{"test-secret"},
		ServiceAccounts:     []string{"test-serviceaccount"},
		ClusterRoleBindings: []string{"test-clusterrolebinding", "test-clusterrolebinding-2"},
		ClusterRoles:        []string{"test-clusterrole"},
	}
	deployment, daemonSet, statefulSet := GetTestDeployment(), GetTestDaemonSet(), GetTestStatefulSet()

	for _, object := range []runtime.Object{&deployment, &daemonSet, &statefulSet} {
		if related := GetClusterRelatedResources(eventObject(t, object)); !reflect.DeepEqual(related, want) {
			t.Errorf("%T: expected %+v, got %+v", object, want, related)
		}
	}
}
