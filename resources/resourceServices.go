package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/strings/slices"
	"main.go/common"
	"reflect"
)

// Secret Kind
func getSecretRelatedPods(secretName string) (relatedPods []string) {

	// List Pods
	pods := GetPods()
	// Create a map of Pods names to Pod objects.
	podsMap := map[string]corev1.Pod{}
	for _, pod := range pods {
		if reflect.ValueOf(pod).IsValid() {
			podsMap[pod.Name] = pod
		}
	}

	// Iterate through the Pods and check for the secret name.
	for podName, pod := range podsMap {
		for _, container := range pod.Spec.Containers {
			containerEnv := container.Env
			if containerEnv != nil {
				for _, env := range containerEnv {
					if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secretName && !slices.Contains(relatedPods, podName) {
						relatedPods = append(relatedPods, podName)
					}
				}
			}

		}
		for _, volume := range pod.Spec.Volumes {
			if reflect.ValueOf(volume).IsValid() && volume.Secret != nil && volume.Secret.SecretName == secretName && !slices.Contains(relatedPods, podName) {
				relatedPods = append(relatedPods, podName)

			}
		}
	}

	return relatedPods
}
func getSecretRelatedDaemonSets(secretName string) (relatedDaemonSets []string) {

	// List DaemonSets
	daemonSets := GetDaemonSets()
	// Create a map of DaemonSet names to DaemonSet objects.
	daemonSetsMap := map[string]appsv1.DaemonSet{}
	for _, daemonSet := range daemonSets {
		if reflect.ValueOf(daemonSet).IsValid() {

			daemonSetsMap[daemonSet.Name] = daemonSet
		}
	}

	// Iterate through the DaemonSets and check for the config map.
	for daemonSetName, daemonSet := range daemonSetsMap {
		for _, container := range daemonSet.Spec.Template.Spec.Containers {
			containerEnv := container.Env
			if containerEnv != nil {
				for _, env := range containerEnv {
					if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secretName && !slices.Contains(relatedDaemonSets, daemonSetName) {
						relatedDaemonSets = append(relatedDaemonSets, daemonSetName)
					}
				}
			}
		}
		for _, volume := range daemonSet.Spec.Template.Spec.Volumes {
			if reflect.ValueOf(volume).IsValid() && volume.Secret != nil && volume.Secret.SecretName == secretName && !slices.Contains(relatedDaemonSets, daemonSetName) {
				relatedDaemonSets = append(relatedDaemonSets, daemonSetName)
			}
		}
	}

	return relatedDaemonSets
}
func getSecretRelatedDeployments(secretName string) (relatedDeployments []string) {

	// List Deployments
	deployments := GetDeployments()
	// Create a map of Deployment names to DaemonSet objects.
	deploymentsMap := map[string]appsv1.Deployment{}
	for _, deployment := range deployments {
		if reflect.ValueOf(deployment).IsValid() {
			deploymentsMap[deployment.Name] = deployment
		}
	}

	// Iterate through the Deployments and check for the config map.
	for deploymentName, deployment := range deploymentsMap {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			containerEnv := container.Env
			if containerEnv != nil {
				for _, env := range containerEnv {
					if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secretName && !slices.Contains(relatedDeployments, deploymentName) {
						relatedDeployments = append(relatedDeployments, deploymentName)
					}
				}
			}
		}
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if reflect.ValueOf(volume).IsValid() && volume.Secret != nil && volume.Secret.SecretName == secretName && !slices.Contains(relatedDeployments, deploymentName) {
				relatedDeployments = append(relatedDeployments, deploymentName)
			}
		}
	}

	return relatedDeployments
}
func getSecretRelatedStatefulSets(secretName string) (relatedStatefulSets []string) {

	// List StatefulSets
	statefulSets := GetStatefulSets()
	// Create a map of StatefulSet names to StatefulSet objects.
	statefulSetsMap := map[string]appsv1.StatefulSet{}
	for _, statefulSet := range statefulSets {
		if reflect.ValueOf(statefulSet).IsValid() {

			statefulSetsMap[statefulSet.Name] = statefulSet
		}
	}

	// Iterate through the StatefulSets and check for the config map.
	for statefulSetName, statefulSet := range statefulSetsMap {
		for _, container := range statefulSet.Spec.Template.Spec.Containers {
			containerEnv := container.Env
			if containerEnv != nil {
				for _, env := range containerEnv {
					if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secretName && !slices.Contains(relatedStatefulSets, statefulSetName) {
						relatedStatefulSets = append(relatedStatefulSets, statefulSetName)
					}
				}
			}
		}
		for _, volume := range statefulSet.Spec.Template.Spec.Volumes {
			if reflect.ValueOf(volume).IsValid() && volume.Secret != nil && volume.Secret.SecretName == secretName && !slices.Contains(relatedStatefulSets, statefulSetName) {
				relatedStatefulSets = append(relatedStatefulSets, statefulSetName)
			}
		}
	}

	return relatedStatefulSets
}
func GetSecretRelatedWorkloads(secretName string) (relatedWorkloads common.RelatedClusterServices) {

	relatedDeployments := getSecretRelatedDeployments(secretName)
	relatedDaemonsets := getSecretRelatedDaemonSets(secretName)
	relatedStatefulSets := getSecretRelatedStatefulSets(secretName)
	relatedPods := getSecretRelatedPods(secretName)

	relatedWorkloads = common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	return relatedWorkloads
}

// ConfigMap Kind
func getConfigMapRelatedPods(configMapName string) (relatedPods []string) {

	// List Pods
	pods := GetPods()
	// Create a map of Pods names to DaemonSet objects.
	podsMap := map[string]corev1.Pod{}
	for _, pod := range pods {
		if reflect.ValueOf(pod).IsValid() {
			podsMap[pod.Name] = pod
		}
	}

	// Iterate through the Pods and check for the config map.
	for podName, pod := range podsMap {
		for _, container := range pod.Spec.Containers {
			containerEnv := container.Env
			if containerEnv != nil {
				for _, env := range containerEnv {
					if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == configMapName && !slices.Contains(relatedPods, podName) {
						relatedPods = append(relatedPods, podName)
					}
				}
			}

		}
		for _, volume := range pod.Spec.Volumes {
			if reflect.ValueOf(volume).IsValid() && !slices.Contains(relatedPods, podName) {
				if volume.ConfigMap != nil {
					if volume.ConfigMap.Name == configMapName {
						relatedPods = append(relatedPods, podName)
					}
				}

			}
		}
	}

	return relatedPods
}
func getConfigMapRelatedDaemonSets(configMapName string) (relatedDaemonSets []string) {

	// List DaemonSets
	daemonSets := GetDaemonSets()
	// Create a map of DaemonSet names to DaemonSet objects.
	daemonSetsMap := map[string]appsv1.DaemonSet{}
	for _, daemonSet := range daemonSets {
		if reflect.ValueOf(daemonSet).IsValid() {

			daemonSetsMap[daemonSet.Name] = daemonSet
		}
	}

	// Iterate through the DaemonSets and check for the config map.
	for daemonSetName, daemonSet := range daemonSetsMap {
		for _, container := range daemonSet.Spec.Template.Spec.Containers {
			containerEnv := container.Env
			if containerEnv != nil {
				for _, env := range containerEnv {
					if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == configMapName && !slices.Contains(relatedDaemonSets, daemonSetName) {
						relatedDaemonSets = append(relatedDaemonSets, daemonSetName)
					}
				}
			}
		}
		for _, volume := range daemonSet.Spec.Template.Spec.Volumes {
			if reflect.ValueOf(volume).IsValid() && volume.ConfigMap != nil && volume.ConfigMap.Name == configMapName && !slices.Contains(relatedDaemonSets, daemonSetName) {
				relatedDaemonSets = append(relatedDaemonSets, daemonSetName)
			}
		}
	}

	return relatedDaemonSets
}
func getConfigMapRelatedDeployments(configMapName string) (relatedDeployments []string) {

	// List Deployments
	deployments := GetDeployments()
	// Create a map of Deployment names to DaemonSet objects.
	deploymentsMap := map[string]appsv1.Deployment{}
	for _, deployment := range deployments {
		if reflect.ValueOf(deployment).IsValid() {
			deploymentsMap[deployment.Name] = deployment
		}
	}

	// Iterate through the Deployments and check for the config map.
	for deploymentName, deployment := range deploymentsMap {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			containerEnv := container.Env
			if containerEnv != nil {
				for _, env := range containerEnv {
					if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == configMapName && !slices.Contains(relatedDeployments, deploymentName) {
						relatedDeployments = append(relatedDeployments, deploymentName)
					}
				}
			}
		}
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if reflect.ValueOf(volume).IsValid() && volume.ConfigMap != nil && volume.ConfigMap.Name == configMapName && !slices.Contains(relatedDeployments, deploymentName) {
				relatedDeployments = append(relatedDeployments, deploymentName)
			}
		}
	}

	return relatedDeployments
}
func getConfigMapRelatedStatefulSets(configMapName string) (relatedStatefulSets []string) {

	// List StatefulSets
	statefulSets := GetStatefulSets()
	// Create a map of StatefulSet names to StatefulSet objects.
	statefulSetsMap := map[string]appsv1.StatefulSet{}
	for _, statefulSet := range statefulSets {
		if reflect.ValueOf(statefulSet).IsValid() {

			statefulSetsMap[statefulSet.Name] = statefulSet
		}
	}

	// Iterate through the StatefulSets and check for the config map.
	for statefulSetName, statefulSet := range statefulSetsMap {
		for _, container := range statefulSet.Spec.Template.Spec.Containers {
			containerEnv := container.Env
			if containerEnv != nil {
				for _, env := range containerEnv {
					if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == configMapName && !slices.Contains(relatedStatefulSets, statefulSetName) {
						relatedStatefulSets = append(relatedStatefulSets, statefulSetName)
					}
				}
			}
		}
		for _, volume := range statefulSet.Spec.Template.Spec.Volumes {
			if reflect.ValueOf(volume).IsValid() && volume.ConfigMap != nil && volume.ConfigMap.Name == configMapName && !slices.Contains(relatedStatefulSets, statefulSetName) {
				relatedStatefulSets = append(relatedStatefulSets, statefulSetName)
			}
		}
	}

	return relatedStatefulSets
}
func GetConfigMapRelatedWorkloads(configMapName string) (relatedWorkloads common.RelatedClusterServices) {

	relatedDeployments := getConfigMapRelatedDeployments(configMapName)
	relatedDaemonsets := getConfigMapRelatedDaemonSets(configMapName)
	relatedStatefulSets := getConfigMapRelatedStatefulSets(configMapName)
	relatedPods := getConfigMapRelatedPods(configMapName)

	relatedWorkloads = common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	return relatedWorkloads
}

// ServiceAccount Kind
func getServiceAccountRelatedPods(serviceAccountName string) (relatedPods []string) {

	// List Pods
	pods := GetPods()
	// Create a map of Pods names to DaemonSet objects.
	podsMap := map[string]corev1.Pod{}
	for _, pod := range pods {
		if reflect.ValueOf(pod).IsValid() {
			podsMap[pod.Name] = pod
		}
	}

	// Iterate through the Pods and check for the config map.
	for podName, pod := range podsMap {
		if pod.Spec.ServiceAccountName == serviceAccountName || pod.Spec.DeprecatedServiceAccount == serviceAccountName {
			relatedPods = append(relatedPods, podName)
		}
	}

	return relatedPods
}
func getServiceAccountRelatedDaemonSets(serviceAccountName string) (relatedDaemonSets []string) {

	// List DaemonSets
	daemonSets := GetDaemonSets()
	// Create a map of DaemonSet names to DaemonSet objects.
	daemonSetsMap := map[string]appsv1.DaemonSet{}
	for _, daemonSet := range daemonSets {
		if reflect.ValueOf(daemonSet).IsValid() {

			daemonSetsMap[daemonSet.Name] = daemonSet
		}
	}

	// Iterate through the DaemonSets and check for the config map.
	for daemonSetName, daemonSet := range daemonSetsMap {
		if daemonSet.Spec.Template.Spec.ServiceAccountName == serviceAccountName || daemonSet.Spec.Template.Spec.DeprecatedServiceAccount == serviceAccountName {
			relatedDaemonSets = append(relatedDaemonSets, daemonSetName)
		}

	}

	return relatedDaemonSets
}
func getServiceAccountRelatedDeployments(serviceAccountName string) (relatedDeployments []string) {

	// List Deployments
	deployments := GetDeployments()
	// Create a map of Deployment names to DaemonSet objects.
	deploymentsMap := map[string]appsv1.Deployment{}
	for _, deployment := range deployments {
		if reflect.ValueOf(deployment).IsValid() {
			deploymentsMap[deployment.Name] = deployment
		}
	}

	// Iterate through the Deployments and check for the config map.
	for deploymentName, deployment := range deploymentsMap {
		if deployment.Spec.Template.Spec.ServiceAccountName == serviceAccountName || deployment.Spec.Template.Spec.DeprecatedServiceAccount == serviceAccountName {

			relatedDeployments = append(relatedDeployments, deploymentName)
		}
	}

	return relatedDeployments
}
func getServiceAccountRelatedStatefulSets(serviceAccountName string) (relatedStatefulSets []string) {

	// List StatefulSets
	statefulSets := GetStatefulSets()
	// Create a map of StatefulSet names to StatefulSet objects.
	statefulSetsMap := map[string]appsv1.StatefulSet{}
	for _, statefulSet := range statefulSets {
		if reflect.ValueOf(statefulSet).IsValid() {

			statefulSetsMap[statefulSet.Name] = statefulSet
		}
	}

	// Iterate through the StatefulSets and check for the config map.
	for statefulSetName, statefulSet := range statefulSetsMap {
		if statefulSet.Spec.Template.Spec.ServiceAccountName == serviceAccountName || statefulSet.Spec.Template.Spec.DeprecatedServiceAccount == serviceAccountName {

			relatedStatefulSets = append(relatedStatefulSets, statefulSetName)
		}
	}

	return relatedStatefulSets
}
func GetServiceAccountRelatedWorkloads(serviceAccountName string) (relatedWorkloads common.RelatedClusterServices) {
	relatedDeployments := getServiceAccountRelatedDeployments(serviceAccountName)
	relatedDaemonsets := getServiceAccountRelatedDaemonSets(serviceAccountName)
	relatedStatefulSets := getServiceAccountRelatedStatefulSets(serviceAccountName)
	relatedPods := getServiceAccountRelatedPods(serviceAccountName)

	relatedWorkloads = common.RelatedClusterServices{Deployments: relatedDeployments, DaemonSets: relatedDaemonsets, StatefulSets: relatedStatefulSets, Pods: relatedPods}

	return relatedWorkloads
}

// ClusterRoleBinding Kind

func GetClusterRoleBindingRelatedWorkloads(clusterRoleBindingName string) (relatedWorkloads common.RelatedClusterServices) {
	//
	clusterRoleBinding := GetClusterRoleBinding(clusterRoleBindingName)

	// Iterate through the StatefulSets and check for the config map.
	if reflect.ValueOf(clusterRoleBinding).IsValid() {
		for _, clusterRoleBindingSubject := range clusterRoleBinding.Subjects {
			if clusterRoleBindingSubject.Kind == "ServiceAccount" {
				serviceAccountName := clusterRoleBindingSubject.Name
				relatedWorkloads = GetServiceAccountRelatedWorkloads(serviceAccountName)
			}
		}

	}

	return relatedWorkloads
}

// ClusterRole Kind

func GetClusterRoleRelatedWorkloads(clusterRoleName string) (relatedWorkloads common.RelatedClusterServices) {

	clusterRoleBindings := GetClusterRoleBindings()

	for _, clusterRoleBinding := range clusterRoleBindings {
		clusterRoleRef := clusterRoleBinding.RoleRef.Name
		if clusterRoleRef == clusterRoleName && reflect.ValueOf(clusterRoleBinding).IsValid() {
			for _, clusterRoleBindingSubject := range clusterRoleBinding.Subjects {
				if clusterRoleBindingSubject.Kind == "ServiceAccount" {
					serviceAccountName := clusterRoleBindingSubject.Name
					relatedWorkloads = GetServiceAccountRelatedWorkloads(serviceAccountName)
				}
			}

		}

	}

	return relatedWorkloads
}
