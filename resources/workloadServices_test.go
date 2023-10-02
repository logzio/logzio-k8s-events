package resources

import (
	"main.go/common"
	"reflect"
	"testing"
)

// getTestWorkloadRelatedClusterRoleBindings is used to get related cluster role bindings for a test workload
func getTestWorkloadRelatedClusterRoleBindings(workload Workload) (relatedClusterRoleBindings []string) {

	if serviceAccountName := workload.GetServiceAccountName(); serviceAccountName != "" {
		clusterRoleBindings := GetTestClusterRoleBindings()
		for _, clusterRoleBinding := range clusterRoleBindings {
			if reflect.ValueOf(clusterRoleBinding).IsValid() {
				for _, clusterRoleBindingSubject := range clusterRoleBinding.Subjects {
					if clusterRoleBindingSubject.Kind == "ServiceAccount" && clusterRoleBindingSubject.Name == serviceAccountName {
						clusterRoleBindingName := clusterRoleBinding.Name
						relatedClusterRoleBindings = append(relatedClusterRoleBindings, clusterRoleBindingName)
					}
				}
			}
		}
	}
	return relatedClusterRoleBindings
}

// getTestWorkloadRelatedClusterRoles is used to get related cluster roles for a test workload
func getTestWorkloadRelatedClusterRoles(workload Workload) (relatedClusterRoles []string) {

	if serviceAccountName := workload.GetServiceAccountName(); serviceAccountName != "" {
		clusterRoleBindings := GetTestClusterRoleBindings()
		for _, clusterRoleBinding := range clusterRoleBindings {
			if reflect.ValueOf(clusterRoleBinding).IsValid() {
				for _, clusterRoleBindingSubject := range clusterRoleBinding.Subjects {
					if clusterRoleBindingSubject.Kind == "ServiceAccount" && clusterRoleBindingSubject.Name == serviceAccountName {
						clusterRoleName := clusterRoleBinding.RoleRef.Name
						relatedClusterRoles = append(relatedClusterRoles, clusterRoleName)
					}
				}
			}
		}
	}

	return relatedClusterRoles
}

// TestDeploymentRelatedResources is used to test getting related resources for a test deployment
func TestDeploymentRelatedResources(t *testing.T) {
	//
	var relatedResources common.RelatedClusterServices
	deployment := GetTestDeployment()
	deploymentWorkload := Deployment(deployment)
	if reflect.ValueOf(deployment).IsValid() {
		relatedConfigMaps := GetWorkloadRelatedConfigMaps(deploymentWorkload)
		relatedSecrets := GetWorkloadRelatedSecrets(deploymentWorkload)
		relatedServiceAccounts := GetWorkloadRelatedServiceAccounts(deploymentWorkload)
		relatedClusterRoleBindings := getTestWorkloadRelatedClusterRoleBindings(deploymentWorkload)
		relatedClusterRoles := getTestWorkloadRelatedClusterRoles(deploymentWorkload)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}

	if reflect.ValueOf(relatedResources).IsZero() {
		t.Errorf("Expected related resources for deployment: %s, got zero", deployment.Name)
	} else {
		t.Logf("Deployment: %s related resources:\n %v", deployment.Name, relatedResources)
	}

}

// TestDaemonSetRelatedResources is used to test getting related resources for a test daemonset
func TestDaemonSetRelatedResources(t *testing.T) {
	var relatedResources common.RelatedClusterServices
	daemonSet := GetTestDaemonSet()
	daemonSetWorkload := DaemonSet(daemonSet)
	if reflect.ValueOf(daemonSet).IsValid() {
		relatedConfigMaps := GetWorkloadRelatedConfigMaps(daemonSetWorkload)
		relatedSecrets := GetWorkloadRelatedSecrets(daemonSetWorkload)
		relatedServiceAccounts := GetWorkloadRelatedServiceAccounts(daemonSetWorkload)
		relatedClusterRoleBindings := getTestWorkloadRelatedClusterRoleBindings(daemonSetWorkload)
		relatedClusterRoles := getTestWorkloadRelatedClusterRoles(daemonSetWorkload)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}
	if reflect.ValueOf(relatedResources).IsZero() {
		t.Errorf("Expected related resources for daemonset: %s, got zero", daemonSet.Name)
	} else {
		t.Logf("Daemonset: %s related resources:\n %v", daemonSet.Name, relatedResources)
	}

}

// TestStatefulSetRelatedResources is used to test getting related resources for a test statefulset
func TestStatefulSetRelatedResources(t *testing.T) {
	//
	var relatedResources common.RelatedClusterServices
	statefulSet := GetTestStatefulSet()
	statefulSetWorkload := StatefulSet(statefulSet)
	if reflect.ValueOf(statefulSet).IsValid() {
		relatedConfigMaps := GetWorkloadRelatedConfigMaps(statefulSetWorkload)
		relatedSecrets := GetWorkloadRelatedSecrets(statefulSetWorkload)
		relatedServiceAccounts := GetWorkloadRelatedServiceAccounts(statefulSetWorkload)
		relatedClusterRoleBindings := getTestWorkloadRelatedClusterRoleBindings(statefulSetWorkload)
		relatedClusterRoles := getTestWorkloadRelatedClusterRoles(statefulSetWorkload)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}
	if reflect.ValueOf(relatedResources).IsZero() {
		t.Errorf("Expected related resources for statefulset: %s, got zero", statefulSet.Name)
	} else {
		t.Logf("Statefulset: %s related resources:\n %v", statefulSet.Name, relatedResources)
	}
}
