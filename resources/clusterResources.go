package resources

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"main.go/common"
	"reflect"
)

// Workload is an interface that provides a common API for Kubernetes workloads (Pod, Deployment, etc.)
type Workload interface {
	GetName() string
	GetContainers() []corev1.Container
	GetVolumes() []corev1.Volume
	GetServiceAccountName() string
}

type Pod corev1.Pod
type Deployment appsv1.Deployment
type DaemonSet appsv1.DaemonSet
type StatefulSet appsv1.StatefulSet

func (p Pod) GetName() string                   { return p.Name }
func (p Pod) GetContainers() []corev1.Container { return p.Spec.Containers }
func (p Pod) GetVolumes() []corev1.Volume       { return p.Spec.Volumes }
func (p Pod) GetServiceAccountName() string     { return p.Spec.ServiceAccountName }

func (d Deployment) GetName() string                   { return d.Name }
func (d Deployment) GetContainers() []corev1.Container { return d.Spec.Template.Spec.Containers }
func (d Deployment) GetVolumes() []corev1.Volume       { return d.Spec.Template.Spec.Volumes }
func (d Deployment) GetServiceAccountName() string     { return d.Spec.Template.Spec.ServiceAccountName }

func (d DaemonSet) GetServiceAccountName() string     { return d.Spec.Template.Spec.ServiceAccountName }
func (d DaemonSet) GetName() string                   { return d.Name }
func (d DaemonSet) GetContainers() []corev1.Container { return d.Spec.Template.Spec.Containers }
func (d DaemonSet) GetVolumes() []corev1.Volume       { return d.Spec.Template.Spec.Volumes }

func (s StatefulSet) GetName() string                   { return s.Name }
func (s StatefulSet) GetContainers() []corev1.Container { return s.Spec.Template.Spec.Containers }
func (s StatefulSet) GetVolumes() []corev1.Volume       { return s.Spec.Template.Spec.Volumes }
func (s StatefulSet) GetServiceAccountName() string     { return s.Spec.Template.Spec.ServiceAccountName }

// GetClusterRoleBindings retrieves all ClusterRoleBindings in the cluster
func GetClusterRoleBindings() (relatedClusterRoleBindings []rbacv1.ClusterRoleBinding) {

	// List ClusterRoleBindings
	clusterRoleBindingsClient := common.K8sClient.RbacV1().ClusterRoleBindings()
	clusterRoleBindings, err := clusterRoleBindingsClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		// Handle error by common the error and returning an empty list of related ClusterRoleBindings.
		log.Printf("[ERROR] Error listing ClusterRoleBindings: %v", err)
		return
	}

	for _, clusterRoleBinding := range clusterRoleBindings.Items {
		if reflect.ValueOf(clusterRoleBinding).IsValid() {
			relatedClusterRoleBindings = append(relatedClusterRoleBindings, clusterRoleBinding)
		}
	}

	return relatedClusterRoleBindings
}

// GetDeployments retrieves all Deployments in the cluster
func GetDeployments() (relatedDeployments []appsv1.Deployment) {
	// List Deployments
	deploymentsClient := common.K8sClient.AppsV1().Deployments("")
	deployments, err := deploymentsClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		// Handle error by common the error and returning an empty list of related deployments.
		log.Printf("[ERROR] Error listing Deployments: %v", err)
		return
	}
	// Create a map of deployment names to deployment objects.
	for _, deployment := range deployments.Items {
		if reflect.ValueOf(deployment).IsValid() {
			relatedDeployments = append(relatedDeployments, deployment)
		}
	}

	return relatedDeployments
}

// GetPods retrieves all Pods in the cluster
func GetPods() (relatedPods []corev1.Pod) {
	// List Pods
	podsClient := common.K8sClient.CoreV1().Pods("")
	pods, err := podsClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		// Handle error by common the error and returning an empty list of related Pods.
		log.Printf("[ERROR] Error listing Pods: %v", err)
		return
	}

	// Create a map of Pods names to Pod objects.
	for _, pod := range pods.Items {
		if reflect.ValueOf(pod).IsValid() {
			relatedPods = append(relatedPods, pod)
		}
	}

	return relatedPods
}

// GetDaemonSets retrieves all DaemonSets in the cluster
func GetDaemonSets() (relatedDaemonSets []appsv1.DaemonSet) {
	// List DaemonSets
	daemonSetsClient := common.K8sClient.AppsV1().DaemonSets("")
	daemonSets, err := daemonSetsClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		// Handle error by common the error and returning an empty list of related DaemonSets.
		log.Printf("[ERROR] Error listing DaemonSets: %v", err)
		return
	}

	// Create a map of DaemonSet names to DaemonSet objects.
	for _, daemonSet := range daemonSets.Items {
		if reflect.ValueOf(daemonSet).IsValid() {

			relatedDaemonSets = append(relatedDaemonSets, daemonSet)
		}
	}

	return relatedDaemonSets
}

// GetStatefulSets retrieves all StatefulSets in the cluster
func GetStatefulSets() (relatedStatefulSets []appsv1.StatefulSet) {
	// List statefulSet
	statefulSetsClient := common.K8sClient.AppsV1().StatefulSets("")
	statefulSets, err := statefulSetsClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		// Handle error by common the error and returning an empty list of related statefulSets.
		log.Printf("[ERROR] Error listing StatefulSets: %v", err)
		return
	}

	for _, statefulSet := range statefulSets.Items {
		if reflect.ValueOf(statefulSet).IsValid() {
			relatedStatefulSets = append(relatedStatefulSets, statefulSet)
		}
	}

	return relatedStatefulSets
}

// GetDeployment retrieves a specific Deployment by name and namespace
func GetDeployment(deploymentName string, namespace string) (relatedDeployment appsv1.Deployment) {

	deploymentsClient := common.K8sClient.AppsV1().Deployments(namespace)
	deployment, err := deploymentsClient.Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		resourceNotFoundErr := errors.IsNotFound(err)
		// Ignore errors of resource not found, as the resource may not exist in the cluster in deletion events.
		if !resourceNotFoundErr {
			log.Printf("[ERROR] Failed to get Deployment: %s in namespace %s\nError: %v", deploymentName, namespace, err)
			return
		}

	}
	if reflect.ValueOf(deployment).IsValid() {
		relatedDeployment = *deployment
	}

	return relatedDeployment
}

// GetDaemonSet retrieves a specific DaemonSet by name and namespace
func GetDaemonSet(daemonSetName string, namespace string) (relatedDaemonSet appsv1.DaemonSet) {

	daemonSetsClient := common.K8sClient.AppsV1().DaemonSets(namespace)
	daemonSet, err := daemonSetsClient.Get(context.Background(), daemonSetName, metav1.GetOptions{})
	if err != nil {
		resourceNotFoundErr := errors.IsNotFound(err)
		// Ignore errors of resource not found, as the resource may no longer exist in the cluster in deletion events.
		if !resourceNotFoundErr {
			log.Printf("[ERROR] Failed to get DaemonSet: %s in namespace %s\nError: %v", daemonSetName, namespace, err)
			return
		}

	}
	if reflect.ValueOf(daemonSet).IsValid() {
		relatedDaemonSet = *daemonSet
	}

	return relatedDaemonSet
}

// GetStatefulSet retrieves a specific StatefulSet by name and namespace
func GetStatefulSet(statefulSetName string, namespace string) (relatedStatefulSet appsv1.StatefulSet) {

	statefulSetsClient := common.K8sClient.AppsV1().StatefulSets(namespace)
	statefulSet, err := statefulSetsClient.Get(context.Background(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		resourceNotFoundErr := errors.IsNotFound(err)
		// Ignore errors of resource not found, as the resource may not exist in the cluster in deletion events.
		if !resourceNotFoundErr {
			log.Printf("[ERROR] Failed to get StatefulSet: %s in namespace %s\nError: %v", statefulSetName, namespace, err)
			return
		}

	}
	if reflect.ValueOf(statefulSet).IsValid() {
		relatedStatefulSet = *statefulSet
	}

	return relatedStatefulSet
}

// GetClusterRoleBinding retrieves a specific ClusterRoleBinding by name and namespace
func GetClusterRoleBinding(clusterRoleBindingName string) (relatedClusterRoleBinding rbacv1.ClusterRoleBinding) {

	clusterRoleBindingsClient := common.K8sClient.RbacV1().ClusterRoleBindings()
	clusterRoleBinding, err := clusterRoleBindingsClient.Get(context.Background(), clusterRoleBindingName, metav1.GetOptions{})
	if err != nil {
		resourceNotFoundErr := errors.IsNotFound(err)
		// Ignore errors of resource not found, as the resource may not exist in the cluster in deletion events.
		if !resourceNotFoundErr {
			log.Printf("[ERROR] Failed to get ClusterRoleBinding: %s\nError: %v", clusterRoleBindingName, err)
			return
		}

	}

	if reflect.ValueOf(clusterRoleBinding).IsValid() {
		relatedClusterRoleBinding = *clusterRoleBinding
	}

	return relatedClusterRoleBinding
}

// GetClusterRelatedResources retrieves all related resources for a given resource kind, name and namespace.
func GetClusterRelatedResources(resourceKind string, resourceName string, namespace string) (relatedClusterServices common.RelatedClusterServices) {

	common.CreateClusterClient()

	if common.K8sClient != nil {
		switch resourceKind {
		case "ConfigMap":
			relatedClusterServices = ConfigMapRelatedWorkloads(resourceName)
		case "Secret":
			relatedClusterServices = SecretRelatedWorkloads(resourceName)
		case "ClusterRoleBinding":
			relatedClusterServices = ClusterRoleBindingRelatedWorkloads(resourceName)
		case "ServiceAccount":
			relatedClusterServices = ServiceAccountRelatedWorkloads(resourceName)
		case "ClusterRole":
			relatedClusterServices = ClusterRoleRelatedWorkloads(resourceName)
		case "Deployment":
			relatedClusterServices = DeploymentRelatedResources(resourceName, namespace)
		case "DaemonSet":
			relatedClusterServices = DaemonSetRelatedResources(resourceName, namespace)
		case "StatefulSet":
			relatedClusterServices = StatefulSetRelatedResources(resourceName, namespace)
		default:
			log.Printf("[ERROR] Unknown resource kind %s", resourceKind)
		}

	} else {
		log.Printf("Failed to parse Resource: %s of kind: %s related cluster services, couldn't create a K8S client.\n", resourceName, resourceKind)
	}

	return relatedClusterServices
}
