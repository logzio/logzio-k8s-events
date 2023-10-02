package resources

import (
	"main.go/common"
	"reflect"
	"testing"
)

// TestSecretRelatedWorkloads tests that the related workloads for a secret are correctly identified.
func TestSecretRelatedWorkloads(t *testing.T) {
	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetTestPods() {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetTestDaemonSets() {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetTestDeployments() {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetTestStatefulSets() {
		statefulsets = append(statefulsets, StatefulSet(statefulset))
	}

	secretName := "test-secret"
	relatedPods := GetSecretRelatedWorkloads(secretName, pods)
	relatedDaemonsets := GetSecretRelatedWorkloads(secretName, daemonsets)
	relatedDeployments := GetSecretRelatedWorkloads(secretName, deployments)
	relatedStatefulSets := GetSecretRelatedWorkloads(secretName, statefulsets)
	// Similarly, call getRelatedWorkloads for other workload types...

	relatedWorkloads := common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	if reflect.ValueOf(relatedWorkloads).IsZero() {
		t.Errorf("Expected related workloads for secret: %s, got zero", secretName)
	} else {
		t.Logf("Secret: %s related workloads:\n%v", secretName, relatedWorkloads)
	}
}

// TestConfigMapRelatedWorkloads tests that the related workloads for a configmap are correctly identified.
func TestConfigMapRelatedWorkloads(t *testing.T) {

	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetTestPods() {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetTestDaemonSets() {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetTestDeployments() {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetTestStatefulSets() {
		statefulsets = append(statefulsets, StatefulSet(statefulset))
	}

	configMapName := "test-configmap"
	relatedPods := GetConfigMapRelatedWorkloads(configMapName, pods)
	relatedDaemonsets := GetConfigMapRelatedWorkloads(configMapName, daemonsets)
	relatedDeployments := GetConfigMapRelatedWorkloads(configMapName, deployments)
	relatedStatefulSets := GetConfigMapRelatedWorkloads(configMapName, statefulsets)

	relatedWorkloads := common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	if reflect.ValueOf(relatedWorkloads).IsZero() {
		t.Errorf("Expected related workloads for configmap: %s, got zero", configMapName)
	} else {
		t.Logf("ConfigMap: %s related workloads:\n%v", configMapName, relatedWorkloads)
	}
}

// TestServiceAccountRelatedWorkloads tests that the related workloads for a service account are correctly identified.
func TestServiceAccountRelatedWorkloads(t *testing.T) {
	serviceAccountName := "test-serviceaccount"

	relatedWorkloads := ServiceAccountTestRelatedWorkloads(serviceAccountName)
	if reflect.ValueOf(relatedWorkloads).IsZero() {
		t.Errorf("Expected related workloads for serviceaccount: %s, got zero", serviceAccountName)
	} else {
		t.Logf("Service account: %s related workloads:\n%v", serviceAccountName, relatedWorkloads)
	}
}

// ServiceAccountTestRelatedWorkloads returns a list of related workloads for a test service account are correctly identified.
func ServiceAccountTestRelatedWorkloads(serviceAccountName string) (relatedWorkloads common.RelatedClusterServices) {
	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetTestPods() {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetTestDaemonSets() {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetTestDeployments() {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetTestStatefulSets() {
		statefulsets = append(statefulsets, StatefulSet(statefulset))
	}

	relatedPods := GetServiceAccountRelatedWorkloads(serviceAccountName, pods)
	relatedDaemonsets := GetServiceAccountRelatedWorkloads(serviceAccountName, daemonsets)
	relatedDeployments := GetServiceAccountRelatedWorkloads(serviceAccountName, deployments)
	relatedStatefulSets := GetServiceAccountRelatedWorkloads(serviceAccountName, statefulsets)

	relatedWorkloads = common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	return relatedWorkloads
}

// TestClusterRoleBindingRelatedWorkloads tests that the related workloads for a cluster role binding are correctly identified.
func TestClusterRoleBindingRelatedWorkloads(t *testing.T) {
	//
	var relatedWorkloads common.RelatedClusterServices
	clusterRoleBindingName := "test-clusterrolebinding"
	clusterRoleBinding := GetTestClusterRoleBinding(clusterRoleBindingName)

	// Iterate through the StatefulSets and check for the config map.
	if reflect.ValueOf(clusterRoleBinding).IsValid() {
		for _, clusterRoleBindingSubject := range clusterRoleBinding.Subjects {
			if clusterRoleBindingSubject.Kind == "ServiceAccount" {
				serviceAccountName := clusterRoleBindingSubject.Name
				relatedWorkloads = ServiceAccountTestRelatedWorkloads(serviceAccountName)
			}
		}

	}

	if reflect.ValueOf(relatedWorkloads).IsZero() {
		t.Errorf("Expected related workloads for cluster role binding: %s, got zero", clusterRoleBindingName)
	} else {
		t.Logf("Cluster role binding: %s related workloads:\n%v", clusterRoleBindingName, relatedWorkloads)
	}
}

// TestClusterRoleRelatedWorkloads tests that the related workloads for a cluster role are correctly identified.
func TestClusterRoleRelatedWorkloads(t *testing.T) {
	var relatedWorkloads common.RelatedClusterServices
	clusterRoleName := "test-clusterrole"
	clusterRoleBindings := GetTestClusterRoleBindings()

	for _, clusterRoleBinding := range clusterRoleBindings {
		clusterRoleRef := clusterRoleBinding.RoleRef.Name
		if clusterRoleRef == clusterRoleName && reflect.ValueOf(clusterRoleBinding).IsValid() {
			for _, clusterRoleBindingSubject := range clusterRoleBinding.Subjects {
				if clusterRoleBindingSubject.Kind == "ServiceAccount" {
					serviceAccountName := clusterRoleBindingSubject.Name
					relatedWorkloads = ServiceAccountTestRelatedWorkloads(serviceAccountName)
				}
			}

		}

	}

	if reflect.ValueOf(relatedWorkloads).IsZero() {
		t.Errorf("Expected related workloads for clusterrole: %s, got zero", clusterRoleName)
	} else {
		t.Logf("Cluster role: %s related workloads:\n%v", clusterRoleName, relatedWorkloads)
	}
}
