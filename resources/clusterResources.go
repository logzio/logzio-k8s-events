package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Workload is an interface that provides a common API for Kubernetes workloads (Pod, Deployment, etc.)
type Workload interface {
	GetName() string
	GetNamespace() string
	GetContainers() []corev1.Container
	GetVolumes() []corev1.Volume
	GetServiceAccountName() string
	GetImagePullSecrets() []corev1.LocalObjectReference
}

type Pod corev1.Pod
type Deployment appsv1.Deployment
type DaemonSet appsv1.DaemonSet
type StatefulSet appsv1.StatefulSet

func (p Pod) GetName() string                                    { return p.Name }
func (p Pod) GetNamespace() string                               { return p.Namespace }
func (p Pod) GetContainers() []corev1.Container                  { return p.Spec.Containers }
func (p Pod) GetVolumes() []corev1.Volume                        { return p.Spec.Volumes }
func (p Pod) GetServiceAccountName() string                      { return p.Spec.ServiceAccountName }
func (p Pod) GetImagePullSecrets() []corev1.LocalObjectReference { return p.Spec.ImagePullSecrets }

func (d Deployment) GetName() string                   { return d.Name }
func (d Deployment) GetNamespace() string              { return d.Namespace }
func (d Deployment) GetContainers() []corev1.Container { return d.Spec.Template.Spec.Containers }
func (d Deployment) GetVolumes() []corev1.Volume       { return d.Spec.Template.Spec.Volumes }
func (d Deployment) GetServiceAccountName() string     { return d.Spec.Template.Spec.ServiceAccountName }
func (d Deployment) GetImagePullSecrets() []corev1.LocalObjectReference {
	return d.Spec.Template.Spec.ImagePullSecrets
}

func (d DaemonSet) GetName() string                   { return d.Name }
func (d DaemonSet) GetNamespace() string              { return d.Namespace }
func (d DaemonSet) GetContainers() []corev1.Container { return d.Spec.Template.Spec.Containers }
func (d DaemonSet) GetVolumes() []corev1.Volume       { return d.Spec.Template.Spec.Volumes }
func (d DaemonSet) GetServiceAccountName() string     { return d.Spec.Template.Spec.ServiceAccountName }
func (d DaemonSet) GetImagePullSecrets() []corev1.LocalObjectReference {
	return d.Spec.Template.Spec.ImagePullSecrets
}

func (s StatefulSet) GetName() string                   { return s.Name }
func (s StatefulSet) GetNamespace() string              { return s.Namespace }
func (s StatefulSet) GetContainers() []corev1.Container { return s.Spec.Template.Spec.Containers }
func (s StatefulSet) GetVolumes() []corev1.Volume       { return s.Spec.Template.Spec.Volumes }
func (s StatefulSet) GetServiceAccountName() string     { return s.Spec.Template.Spec.ServiceAccountName }
func (s StatefulSet) GetImagePullSecrets() []corev1.LocalObjectReference {
	return s.Spec.Template.Spec.ImagePullSecrets
}
