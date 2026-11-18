package resources

import (
	rbacv1 "k8s.io/api/rbac/v1"
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
func SecretRelatedWorkloads(secretName string, namespace string) (relatedWorkloads common.RelatedClusterServices) {
	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetPods(namespace) {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetDaemonSets(namespace) {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetDeployments(namespace) {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetStatefulSets(namespace) {
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
func ConfigMapRelatedWorkloads(configMapName string, namespace string) (relatedWorkloads common.RelatedClusterServices) {
	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetPods(namespace) {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetDaemonSets(namespace) {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetDeployments(namespace) {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetStatefulSets(namespace) {
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
func ServiceAccountRelatedWorkloads(serviceAccountName string, namespace string) (relatedWorkloads common.RelatedClusterServices) {
	var pods []Workload
	var daemonsets []Workload
	var deployments []Workload
	var statefulsets []Workload
	for _, pod := range GetPods(namespace) {
		pods = append(pods, Pod(pod))
	}
	for _, daemonset := range GetDaemonSets(namespace) {
		daemonsets = append(daemonsets, DaemonSet(daemonset))
	}
	for _, deployment := range GetDeployments(namespace) {
		deployments = append(deployments, Deployment(deployment))
	}
	for _, statefulset := range GetStatefulSets(namespace) {
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
	return subjectsRelatedWorkloads(GetClusterRoleBinding(clusterRoleBindingName).Subjects)
}

// ClusterRoleRelatedWorkloads returns a list of workloads that reference a given cluster role.
func ClusterRoleRelatedWorkloads(clusterRoleName string) (relatedWorkloads common.RelatedClusterServices) {
	for _, clusterRoleBinding := range GetClusterRoleBindings() {
		if clusterRoleBinding.RoleRef.Name == clusterRoleName {
			relatedWorkloads = mergeRelatedWorkloads(relatedWorkloads, subjectsRelatedWorkloads(clusterRoleBinding.Subjects))
		}
	}
	return relatedWorkloads
}

// subjectsRelatedWorkloads returns the workloads of every service account subject, each in its own namespace
func subjectsRelatedWorkloads(subjects []rbacv1.Subject) (relatedWorkloads common.RelatedClusterServices) {
	for _, subject := range subjects {
		if subject.Kind == "ServiceAccount" {
			relatedWorkloads = mergeRelatedWorkloads(relatedWorkloads, ServiceAccountRelatedWorkloads(subject.Name, subject.Namespace))
		}
	}
	return relatedWorkloads
}

// mergeRelatedWorkloads adds the workloads in other to related, without duplicates
func mergeRelatedWorkloads(related, other common.RelatedClusterServices) common.RelatedClusterServices {
	related.Deployments = appendMissing(related.Deployments, other.Deployments)
	related.DaemonSets = appendMissing(related.DaemonSets, other.DaemonSets)
	related.StatefulSets = appendMissing(related.StatefulSets, other.StatefulSets)
	related.Pods = appendMissing(related.Pods, other.Pods)
	return related
}

func appendMissing(names, more []string) []string {
	for _, name := range more {
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	return names
}
