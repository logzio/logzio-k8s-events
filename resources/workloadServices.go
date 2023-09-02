package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/strings/slices"
	"main.go/common"
	"reflect"
)

// Deployment Kind

func getDeploymentRelatedConfigMaps(deployment appsv1.Deployment) (relatedConfigMaps []string) {

	for _, container := range deployment.Spec.Template.Spec.Containers {
		containerEnv := container.Env
		if containerEnv != nil {
			for _, env := range containerEnv {
				if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name != "" && !slices.Contains(relatedConfigMaps, env.ValueFrom.ConfigMapKeyRef.Name) {
					configMapName := env.ValueFrom.ConfigMapKeyRef.Name
					relatedConfigMaps = append(relatedConfigMaps, configMapName)
				}
			}
		}
	}
	for _, volume := range deployment.Spec.Template.Spec.Volumes {

		if reflect.ValueOf(volume).IsValid() && volume.ConfigMap != nil && volume.ConfigMap.Name != "" && !slices.Contains(relatedConfigMaps, volume.ConfigMap.Name) {
			configMapName := volume.ConfigMap.Name
			relatedConfigMaps = append(relatedConfigMaps, configMapName)
		}
	}

	return relatedConfigMaps
}

func getDeploymentRelatedSecrets(deployment appsv1.Deployment) (relatedSecrets []string) {

	for _, container := range deployment.Spec.Template.Spec.Containers {
		containerEnv := container.Env
		if containerEnv != nil {
			for _, env := range containerEnv {
				if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name != "" && !slices.Contains(relatedSecrets, env.ValueFrom.SecretKeyRef.Name) {
					secretName := env.ValueFrom.SecretKeyRef.Name
					relatedSecrets = append(relatedSecrets, secretName)
				}
			}
		}
	}
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if reflect.ValueOf(volume).IsValid() && volume.Secret != nil && volume.Secret.SecretName != "" && !slices.Contains(relatedSecrets, volume.Secret.SecretName) {
			secretName := volume.Secret.SecretName
			relatedSecrets = append(relatedSecrets, secretName)
		}
	}

	return relatedSecrets
}

func getDeploymentRelatedServiceAccounts(deployment appsv1.Deployment) (relatedServiceAccounts []string) {

	if deployment.Spec.Template.Spec.ServiceAccountName != "" && !slices.Contains(relatedServiceAccounts, deployment.Spec.Template.Spec.ServiceAccountName) {
		serviceAccountName := deployment.Spec.Template.Spec.ServiceAccountName
		relatedServiceAccounts = append(relatedServiceAccounts, serviceAccountName)
	}
	return relatedServiceAccounts
}

func getDeploymentRelatedClusterRoleBindings(deployment appsv1.Deployment) (relatedClusterRoleBindings []string) {

	if serviceAccountName := deployment.Spec.Template.Spec.ServiceAccountName; serviceAccountName != "" {
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

func getDeploymentRelatedClusterRoles(deployment appsv1.Deployment) (relatedClusterRoles []string) {

	if serviceAccountName := deployment.Spec.Template.Spec.ServiceAccountName; serviceAccountName != "" {
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

func GetDeploymentRelatedResources(deploymentName string, namespace string) (relatedResources common.RelatedClusterServices) {
	//
	deployment := GetDeployment(deploymentName, namespace)
	if reflect.ValueOf(deployment).IsValid() {
		relatedConfigMaps := getDeploymentRelatedConfigMaps(deployment)
		relatedSecrets := getDeploymentRelatedSecrets(deployment)
		relatedServiceAccounts := getDeploymentRelatedServiceAccounts(deployment)
		relatedClusterRoleBindings := getDeploymentRelatedClusterRoleBindings(deployment)
		relatedClusterRoles := getDeploymentRelatedClusterRoles(deployment)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}

	return relatedResources
}

// DaemonSet Kind
func getDaemonSetRelatedConfigMaps(daemonSet appsv1.DaemonSet) (relatedConfigMaps []string) {
	//

	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		containerEnv := container.Env
		if containerEnv != nil {
			for _, env := range containerEnv {
				if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name != "" && !slices.Contains(relatedConfigMaps, env.ValueFrom.ConfigMapKeyRef.Name) {
					configMapName := env.ValueFrom.ConfigMapKeyRef.Name
					relatedConfigMaps = append(relatedConfigMaps, configMapName)
				}
			}
		}
	}
	for _, volume := range daemonSet.Spec.Template.Spec.Volumes {

		if reflect.ValueOf(volume).IsValid() && volume.ConfigMap != nil && volume.ConfigMap.Name != "" && !slices.Contains(relatedConfigMaps, volume.ConfigMap.Name) {
			configMapName := volume.ConfigMap.Name
			relatedConfigMaps = append(relatedConfigMaps, configMapName)
		}
	}

	return relatedConfigMaps
}

func getDaemonSetRelatedSecrets(daemonSet appsv1.DaemonSet) (relatedSecrets []string) {
	//

	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		containerEnv := container.Env
		if containerEnv != nil {
			for _, env := range containerEnv {
				if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name != "" && !slices.Contains(relatedSecrets, env.ValueFrom.SecretKeyRef.Name) {
					secretName := env.ValueFrom.SecretKeyRef.Name
					relatedSecrets = append(relatedSecrets, secretName)
				}
			}
		}
	}
	for _, volume := range daemonSet.Spec.Template.Spec.Volumes {
		if reflect.ValueOf(volume).IsValid() && volume.Secret != nil && volume.Secret.SecretName != "" && !slices.Contains(relatedSecrets, volume.Secret.SecretName) {
			secretName := volume.Secret.SecretName
			relatedSecrets = append(relatedSecrets, secretName)
		}
	}

	return relatedSecrets
}

func getDaemonSetRelatedServiceAccounts(daemonSet appsv1.DaemonSet) (relatedServiceAccounts []string) {
	//

	if daemonSet.Spec.Template.Spec.ServiceAccountName != "" && !slices.Contains(relatedServiceAccounts, daemonSet.Spec.Template.Spec.ServiceAccountName) {
		serviceAccountName := daemonSet.Spec.Template.Spec.ServiceAccountName
		relatedServiceAccounts = append(relatedServiceAccounts, serviceAccountName)
	}
	return relatedServiceAccounts
}

func getDaemonSetRelatedClusterRoleBindings(daemonSet appsv1.DaemonSet) (relatedClusterRoleBindings []string) {
	//

	if serviceAccountName := daemonSet.Spec.Template.Spec.ServiceAccountName; serviceAccountName != "" {
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

func getDaemonSetRelatedClusterRoles(daemonSet appsv1.DaemonSet) (relatedClusterRoles []string) {
	//

	if serviceAccountName := daemonSet.Spec.Template.Spec.ServiceAccountName; serviceAccountName != "" {
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

func GetDaemonSetRelatedResources(daemonSetName string, namespace string) (relatedResources common.RelatedClusterServices) {
	//

	daemonSet := GetDaemonSet(daemonSetName, namespace)
	if reflect.ValueOf(daemonSet).IsValid() {
		relatedConfigMaps := getDaemonSetRelatedConfigMaps(daemonSet)
		relatedSecrets := getDaemonSetRelatedSecrets(daemonSet)
		relatedServiceAccounts := getDaemonSetRelatedServiceAccounts(daemonSet)
		relatedClusterRoleBindings := getDaemonSetRelatedClusterRoleBindings(daemonSet)
		relatedClusterRoles := getDaemonSetRelatedClusterRoles(daemonSet)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}

	return relatedResources
}

// StatefulSet Kind
func getStatefulSetRelatedConfigMaps(statefulSet appsv1.StatefulSet) (relatedConfigMaps []string) {
	//
	for _, container := range statefulSet.Spec.Template.Spec.Containers {
		containerEnv := container.Env
		if containerEnv != nil {
			for _, env := range containerEnv {
				if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name != "" && !slices.Contains(relatedConfigMaps, env.ValueFrom.ConfigMapKeyRef.Name) {
					configMapName := env.ValueFrom.ConfigMapKeyRef.Name
					relatedConfigMaps = append(relatedConfigMaps, configMapName)
				}
			}
		}
	}
	for _, volume := range statefulSet.Spec.Template.Spec.Volumes {

		if reflect.ValueOf(volume).IsValid() && volume.ConfigMap != nil && volume.ConfigMap.Name != "" && !slices.Contains(relatedConfigMaps, volume.ConfigMap.Name) {
			configMapName := volume.ConfigMap.Name
			relatedConfigMaps = append(relatedConfigMaps, configMapName)
		}
	}

	return relatedConfigMaps
}

func getStatefulSetRelatedSecrets(statefulSet appsv1.StatefulSet) (relatedSecrets []string) {
	//

	for _, container := range statefulSet.Spec.Template.Spec.Containers {
		containerEnv := container.Env
		if containerEnv != nil {
			for _, env := range containerEnv {
				if reflect.ValueOf(env).IsValid() && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name != "" && !slices.Contains(relatedSecrets, env.ValueFrom.SecretKeyRef.Name) {
					secretName := env.ValueFrom.SecretKeyRef.Name
					relatedSecrets = append(relatedSecrets, secretName)
				}
			}
		}
	}
	for _, volume := range statefulSet.Spec.Template.Spec.Volumes {
		if reflect.ValueOf(volume).IsValid() && volume.Secret != nil && volume.Secret.SecretName != "" && !slices.Contains(relatedSecrets, volume.Secret.SecretName) {
			secretName := volume.Secret.SecretName
			relatedSecrets = append(relatedSecrets, secretName)
		}
	}

	return relatedSecrets
}

func getStatefulSetRelatedServiceAccounts(statefulSet appsv1.StatefulSet) (relatedServiceAccounts []string) {
	//
	if statefulSet.Spec.Template.Spec.ServiceAccountName != "" && !slices.Contains(relatedServiceAccounts, statefulSet.Spec.Template.Spec.ServiceAccountName) {
		serviceAccountName := statefulSet.Spec.Template.Spec.ServiceAccountName
		relatedServiceAccounts = append(relatedServiceAccounts, serviceAccountName)
	}
	return relatedServiceAccounts
}

func getStatefulSetRelatedClusterRoleBindings(statefulSet appsv1.StatefulSet) (relatedClusterRoleBindings []string) {
	//

	if serviceAccountName := statefulSet.Spec.Template.Spec.ServiceAccountName; serviceAccountName != "" {
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

func getStatefulSetRelatedClusterRoles(statefulSet appsv1.StatefulSet) (relatedClusterRoles []string) {
	//

	if serviceAccountName := statefulSet.Spec.Template.Spec.ServiceAccountName; serviceAccountName != "" {
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

func GetStatefulSetRelatedResources(statefulSetName string, namespace string) (relatedResources common.RelatedClusterServices) {
	//

	statefulSet := GetStatefulSet(statefulSetName, namespace)
	if reflect.ValueOf(statefulSet).IsValid() {
		relatedConfigMaps := getStatefulSetRelatedConfigMaps(statefulSet)
		relatedSecrets := getStatefulSetRelatedSecrets(statefulSet)
		relatedServiceAccounts := getStatefulSetRelatedServiceAccounts(statefulSet)
		relatedClusterRoleBindings := getStatefulSetRelatedClusterRoleBindings(statefulSet)
		relatedClusterRoles := getStatefulSetRelatedClusterRoles(statefulSet)
		relatedResources = common.RelatedClusterServices{ConfigMaps: relatedConfigMaps, Secrets: relatedSecrets, ServiceAccounts: relatedServiceAccounts, ClusterRoleBindings: relatedClusterRoleBindings, ClusterRoles: relatedClusterRoles}
	}

	return relatedResources
}
