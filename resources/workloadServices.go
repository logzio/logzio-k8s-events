package resources

import (
	"slices"
)

// appendName adds the name if it's set and not in the list yet
func appendName(names []string, name string) []string {
	if name != "" && !slices.Contains(names, name) {
		names = append(names, name)
	}
	return names
}

// GetWorkloadRelatedConfigMaps returns a list of all config maps related to the workload
func GetWorkloadRelatedConfigMaps(workload Workload) (relatedConfigMaps []string) {
	for _, container := range workload.GetContainers() {
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
				relatedConfigMaps = appendName(relatedConfigMaps, env.ValueFrom.ConfigMapKeyRef.Name)
			}
		}
		for _, envFrom := range container.EnvFrom {
			if envFrom.ConfigMapRef != nil {
				relatedConfigMaps = appendName(relatedConfigMaps, envFrom.ConfigMapRef.Name)
			}
		}
	}
	for _, volume := range workload.GetVolumes() {
		if volume.ConfigMap != nil {
			relatedConfigMaps = appendName(relatedConfigMaps, volume.ConfigMap.Name)
		}
	}
	return relatedConfigMaps
}

// GetWorkloadRelatedSecrets returns a list of all secrets related to the workload
func GetWorkloadRelatedSecrets(workload Workload) (relatedSecrets []string) {
	for _, container := range workload.GetContainers() {
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				relatedSecrets = appendName(relatedSecrets, env.ValueFrom.SecretKeyRef.Name)
			}
		}
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				relatedSecrets = appendName(relatedSecrets, envFrom.SecretRef.Name)
			}
		}
	}
	for _, volume := range workload.GetVolumes() {
		if volume.Secret != nil {
			relatedSecrets = appendName(relatedSecrets, volume.Secret.SecretName)
		}
	}
	for _, pullSecret := range workload.GetImagePullSecrets() {
		relatedSecrets = appendName(relatedSecrets, pullSecret.Name)
	}
	return relatedSecrets
}

// GetWorkloadRelatedServiceAccounts returns a list of all service accounts related to the workload
func GetWorkloadRelatedServiceAccounts(workload Workload) (relatedServiceAccounts []string) {
	return appendName(relatedServiceAccounts, workload.GetServiceAccountName())
}
