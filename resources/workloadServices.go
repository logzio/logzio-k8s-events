package resources

import (
	"k8s.io/utils/strings/slices"
	"main.go/common"
	"reflect"
)

// GetWorkloadRelatedConfigMaps returns a list of all config maps related to the workload
func GetWorkloadRelatedConfigMaps(workload Workload) (relatedConfigMaps []string) {
	for _, container := range workload.GetContainers() {
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name != "" && !slices.Contains(relatedConfigMaps, env.ValueFrom.ConfigMapKeyRef.Name) {
				relatedConfigMaps = append(relatedConfigMaps, env.ValueFrom.ConfigMapKeyRef.Name)
			}
		}
	}
	for _, volume := range workload.GetVolumes() {
		if volume.ConfigMap != nil && volume.ConfigMap.Name != "" && !slices.Contains(relatedConfigMaps, volume.ConfigMap.Name) {
			relatedConfigMaps = append(relatedConfigMaps, volume.ConfigMap.Name)
		}
	}
	return relatedConfigMaps
}

// GetWorkloadRelatedSecrets returns a list of all secrets related to the workload
func GetWorkloadRelatedSecrets(workload Workload) (relatedSecrets []string) {
	for _, container := range workload.GetContainers() {
		for _, env := range container.Env {
			if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name != "" && !slices.Contains(relatedSecrets, env.ValueFrom.SecretKeyRef.Name) {
				relatedSecrets = append(relatedSecrets, env.ValueFrom.SecretKeyRef.Name)
			}
		}
	}
	for _, volume := range workload.GetVolumes() {
		if reflect.ValueOf(volume).IsValid() && volume.Secret != nil && volume.Secret.SecretName != "" && !slices.Contains(relatedSecrets, volume.Secret.SecretName) {
			secretName := volume.Secret.SecretName
			relatedSecrets = append(relatedSecrets, secretName)
		}
	}
	return relatedSecrets
}

// GetWorkloadRelatedServiceAccounts returns a list of all service accounts related to the workload
func GetWorkloadRelatedServiceAccounts(workload Workload) (relatedServiceAccounts []string) {

	if workload.GetServiceAccountName() != "" && !slices.Contains(relatedServiceAccounts, workload.GetServiceAccountName()) {
		relatedServiceAccounts = append(relatedServiceAccounts, workload.GetServiceAccountName())
	}
	return relatedServiceAccounts
}

// GetWorkloadRelatedClusterRoleBindings returns a list of all cluster role bindings related to the workload
func GetWorkloadRelatedClusterRoleBindings(workload Workload) (relatedClusterRoleBindings []string) {

	if serviceAccountName := workload.GetServiceAccountName(); serviceAccountName != "" {
		clusterRoleBindings := GetClusterRoleBindings()
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

// GetWorkloadRelatedClusterRoles returns a list of all cluster roles related to the workload
func GetWorkloadRelatedClusterRoles(workload Workload) (relatedClusterRoles []string) {

	if serviceAccountName := workload.GetServiceAccountName(); serviceAccountName != "" {
		clusterRoleBindings := GetClusterRoleBindings()
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

// DeploymentRelatedResources returns a list of all resources related to the deployment
func DeploymentRelatedResources(deploymentName string, namespace string) (relatedResources common.RelatedClusterServices) {
	//
	deployment := GetDeployment(deploymentName, namespace)
	deploymentWorkload := Deployment(deployment)
	if reflect.ValueOf(deployment).IsValid() {
		relatedConfigMaps := GetWorkloadRelatedConfigMaps(deploymentWorkload)
		relatedSecrets := GetWorkloadRelatedSecrets(deploymentWorkload)
		relatedServiceAccounts := GetWorkloadRelatedServiceAccounts(deploymentWorkload)
		relatedClusterRoleBindings := GetWorkloadRelatedClusterRoleBindings(deploymentWorkload)
		relatedClusterRoles := GetWorkloadRelatedClusterRoles(deploymentWorkload)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}
	return relatedResources
}

// DaemonSetRelatedResources returns a list of all resources related to the daemonset
func DaemonSetRelatedResources(daemonSetName string, namespace string) (relatedResources common.RelatedClusterServices) {
	//

	daemonSet := GetDaemonSet(daemonSetName, namespace)
	daemonSetWorkload := DaemonSet(daemonSet)
	if reflect.ValueOf(daemonSet).IsValid() {
		relatedConfigMaps := GetWorkloadRelatedConfigMaps(daemonSetWorkload)
		relatedSecrets := GetWorkloadRelatedSecrets(daemonSetWorkload)
		relatedServiceAccounts := GetWorkloadRelatedServiceAccounts(daemonSetWorkload)
		relatedClusterRoleBindings := GetWorkloadRelatedClusterRoleBindings(daemonSetWorkload)
		relatedClusterRoles := GetWorkloadRelatedClusterRoles(daemonSetWorkload)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}

	return relatedResources
}

// StatefulSetRelatedResources returns a list of all resources related to the statefulset
func StatefulSetRelatedResources(statefulSetName string, namespace string) (relatedResources common.RelatedClusterServices) {
	//

	statefulSet := GetStatefulSet(statefulSetName, namespace)
	statefulSetWorkload := StatefulSet(statefulSet)
	if reflect.ValueOf(statefulSet).IsValid() {
		relatedConfigMaps := GetWorkloadRelatedConfigMaps(statefulSetWorkload)
		relatedSecrets := GetWorkloadRelatedSecrets(statefulSetWorkload)
		relatedServiceAccounts := GetWorkloadRelatedServiceAccounts(statefulSetWorkload)
		relatedClusterRoleBindings := GetWorkloadRelatedClusterRoleBindings(statefulSetWorkload)
		relatedClusterRoles := GetWorkloadRelatedClusterRoles(statefulSetWorkload)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}

	return relatedResources
}
