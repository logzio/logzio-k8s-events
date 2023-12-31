package resources

import (
	"k8s.io/utils/strings/slices"
	"main.go/common"
	"reflect"
)

// GetSecretRelatedWorkloads returns a list of workloads that reference a given secret and are not part of a given list of workloads.
func GetSecretRelatedWorkloads(secretName string, workloads []Workload) (relatedWorkloads []string) {
	// Create a map of workload names to workloads.
	workloadsMap := map[string]Workload{}
	for _, workload := range workloads {
		if reflect.ValueOf(workload).IsValid() {
			workloadsMap[workload.GetName()] = workload
		}
	}

	// Iterate through the workloads and check for the secret name.
	for workloadName, workload := range workloadsMap {
		for _, container := range workload.GetContainers() {
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secretName && !slices.Contains(relatedWorkloads, workloadName) {
					relatedWorkloads = append(relatedWorkloads, workloadName)
				}
			}
		}
		for _, volume := range workload.GetVolumes() {
			if volume.Secret != nil && volume.Secret.SecretName == secretName && !slices.Contains(relatedWorkloads, workloadName) {
				relatedWorkloads = append(relatedWorkloads, workloadName)
			}
		}
	}

	return relatedWorkloads
}

// SecretRelatedWorkloads returns a list of workloads that reference a given secret.
func SecretRelatedWorkloads(secretName string) (relatedWorkloads common.RelatedClusterServices) {
	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetPods() {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetDaemonSets() {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetDeployments() {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetStatefulSets() {
		statefulsets = append(statefulsets, StatefulSet(statefulset))
	}
	relatedPods := GetSecretRelatedWorkloads(secretName, pods)
	relatedDaemonsets := GetSecretRelatedWorkloads(secretName, daemonsets)
	relatedDeployments := GetSecretRelatedWorkloads(secretName, deployments)
	relatedStatefulSets := GetSecretRelatedWorkloads(secretName, statefulsets)
	// Similarly, call getRelatedWorkloads for other workload types...

	relatedWorkloads = common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	return relatedWorkloads
}

// GetConfigMapRelatedWorkloads returns a list of workloads that reference a given config map and are not part of a given list of workloads.
func GetConfigMapRelatedWorkloads(configMapName string, workloads []Workload) (relatedWorkloads []string) {
	for _, workload := range workloads {
		if reflect.ValueOf(workload).IsValid() {
			for _, container := range workload.GetContainers() {
				for _, env := range container.Env {
					if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == configMapName && !slices.Contains(relatedWorkloads, workload.GetName()) {
						relatedWorkloads = append(relatedWorkloads, workload.GetName())
					}
				}
			}
			for _, volume := range workload.GetVolumes() {
				if volume.ConfigMap != nil && volume.ConfigMap.Name == configMapName && !slices.Contains(relatedWorkloads, workload.GetName()) {
					relatedWorkloads = append(relatedWorkloads, workload.GetName())
				}
			}
		}
	}
	return relatedWorkloads
}

// ConfigMapRelatedWorkloads returns a list of workloads that reference a given config map.
func ConfigMapRelatedWorkloads(configMapName string) (relatedWorkloads common.RelatedClusterServices) {
	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetPods() {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetDaemonSets() {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetDeployments() {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetStatefulSets() {
		statefulsets = append(statefulsets, StatefulSet(statefulset))
	}
	relatedPods := GetConfigMapRelatedWorkloads(configMapName, pods)
	relatedDaemonsets := GetConfigMapRelatedWorkloads(configMapName, daemonsets)
	relatedDeployments := GetConfigMapRelatedWorkloads(configMapName, deployments)
	relatedStatefulSets := GetConfigMapRelatedWorkloads(configMapName, statefulsets)

	relatedWorkloads = common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	return relatedWorkloads
}

// GetServiceAccountRelatedWorkloads returns a list of workloads that reference a given service account and are not part of a given list of workloads.
func GetServiceAccountRelatedWorkloads(serviceAccountName string, workloads []Workload) (relatedWorkloads []string) {
	for _, workload := range workloads {
		if reflect.ValueOf(workload).IsValid() && (workload.GetServiceAccountName() == serviceAccountName || workload.GetServiceAccountName() == serviceAccountName) {
			relatedWorkloads = append(relatedWorkloads, workload.GetName())
		}
	}
	return relatedWorkloads
}

// ServiceAccountRelatedWorkloads returns a list of workloads that reference a given service account.
func ServiceAccountRelatedWorkloads(serviceAccountName string) (relatedWorkloads common.RelatedClusterServices) {
	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetPods() {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetDaemonSets() {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetDeployments() {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetStatefulSets() {
		statefulsets = append(statefulsets, StatefulSet(statefulset))
	}
	relatedPods := GetServiceAccountRelatedWorkloads(serviceAccountName, pods)
	relatedDaemonsets := GetServiceAccountRelatedWorkloads(serviceAccountName, daemonsets)
	relatedDeployments := GetServiceAccountRelatedWorkloads(serviceAccountName, deployments)
	relatedStatefulSets := GetServiceAccountRelatedWorkloads(serviceAccountName, statefulsets)

	relatedWorkloads = common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	return relatedWorkloads
}

// ClusterRoleBindingRelatedWorkloads returns a list of workloads that reference a given cluster role binding.
func ClusterRoleBindingRelatedWorkloads(clusterRoleBindingName string) (relatedWorkloads common.RelatedClusterServices) {
	//
	clusterRoleBinding := GetClusterRoleBinding(clusterRoleBindingName)

	// Iterate through the StatefulSets and check for the config map.
	if reflect.ValueOf(clusterRoleBinding).IsValid() {
		for _, clusterRoleBindingSubject := range clusterRoleBinding.Subjects {
			if clusterRoleBindingSubject.Kind == "ServiceAccount" {
				serviceAccountName := clusterRoleBindingSubject.Name
				relatedWorkloads = ServiceAccountRelatedWorkloads(serviceAccountName)
			}
		}

	}

	return relatedWorkloads
}

// ClusterRoleRelatedWorkloads returns a list of workloads that reference a given cluster role.
func ClusterRoleRelatedWorkloads(clusterRoleName string) (relatedWorkloads common.RelatedClusterServices) {

	clusterRoleBindings := GetClusterRoleBindings()

	for _, clusterRoleBinding := range clusterRoleBindings {
		clusterRoleRef := clusterRoleBinding.RoleRef.Name
		if clusterRoleRef == clusterRoleName && reflect.ValueOf(clusterRoleBinding).IsValid() {
			for _, clusterRoleBindingSubject := range clusterRoleBinding.Subjects {
				if clusterRoleBindingSubject.Kind == "ServiceAccount" {
					serviceAccountName := clusterRoleBindingSubject.Name
					relatedWorkloads = ServiceAccountRelatedWorkloads(serviceAccountName)
				}
			}

		}

	}

	return relatedWorkloads
}
