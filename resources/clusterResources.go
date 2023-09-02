package resources

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"main.go/common"
	"reflect"
)

func GetClusterRoleBindings() (relatedClusterRoleBindings []rbacv1.ClusterRoleBinding) {

	// List clusterRoleBinding
	clusterRoleBindingsClient := common.K8sClient.RbacV1().ClusterRoleBindings()
	clusterRoleBindings, err := clusterRoleBindingsClient.List(context.TODO(), metav1.ListOptions{})
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

func GetDeployments() (relatedDeployments []appsv1.Deployment) {

	//		// List Deployments
	deploymentsClient := common.K8sClient.AppsV1().Deployments("")
	deployments, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// Handle error by common the error and returning an empty list of related DaemonSets.
		log.Printf("[ERROR] Error listing Deployments: %v", err)
		return
	}
	// Create a map of DaemonSet names to DaemonSet objects.
	for _, deployment := range deployments.Items {
		if reflect.ValueOf(deployment).IsValid() {
			relatedDeployments = append(relatedDeployments, deployment)
		}
	}

	// Iterate through the DaemonSets and check for the config map.

	return relatedDeployments
}

func GetPods() (relatedPods []corev1.Pod) {

	// List DaemonSets
	podsClient := common.K8sClient.CoreV1().Pods("")
	pods, err := podsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// Handle error by common the error and returning an empty list of related Pods.
		log.Printf("[ERROR] Error listing Pods: %v", err)
		return
	}

	// Create a map of Pods names to DaemonSet objects.
	for _, pod := range pods.Items {
		if reflect.ValueOf(pod).IsValid() {
			relatedPods = append(relatedPods, pod)
		}
	}

	return relatedPods
}

func GetDaemonSets() (relatedDaemonSets []appsv1.DaemonSet) {

	// List DaemonSets
	daemonSetsClient := common.K8sClient.AppsV1().DaemonSets("")
	daemonSets, err := daemonSetsClient.List(context.TODO(), metav1.ListOptions{})
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

func GetStatefulSets() (relatedStatefulSets []appsv1.StatefulSet) {

	// List statefulSet
	statefulSetsClient := common.K8sClient.AppsV1().StatefulSets("")
	statefulSets, err := statefulSetsClient.List(context.TODO(), metav1.ListOptions{})
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

func GetDeployment(deploymentName string, namespace string) (relatedDeployment appsv1.Deployment) {

	deploymentsClient := common.K8sClient.AppsV1().Deployments(namespace)
	deployment, err := deploymentsClient.Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		log.Printf("[ERROR] Error getting Deployment: %s \nError: %v", deploymentName, err)
		return
	}
	if reflect.ValueOf(deployment).IsValid() {
		relatedDeployment = *deployment
	}

	return relatedDeployment
}

func GetDaemonSet(daemonSetName string, namespace string) (relatedDaemonSet appsv1.DaemonSet) {

	daemonSetsClient := common.K8sClient.AppsV1().DaemonSets(namespace)
	daemonSet, err := daemonSetsClient.Get(context.TODO(), daemonSetName, metav1.GetOptions{})
	if err != nil {
		log.Printf("[ERROR] Error getting DaemonSet: %s \nError: %v", daemonSetName, err)
		return
	}
	if reflect.ValueOf(daemonSet).IsValid() {
		relatedDaemonSet = *daemonSet
	}

	return relatedDaemonSet
}

func GetStatefulSet(statefulSetName string, namespace string) (relatedStatefulSet appsv1.StatefulSet) {

	statefulSetsClient := common.K8sClient.AppsV1().StatefulSets(namespace)
	statefulSet, err := statefulSetsClient.Get(context.TODO(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		log.Printf("[ERROR] Error getting statefulSet: %s \nError: %v", statefulSetName, err)
		return
	}
	if reflect.ValueOf(statefulSet).IsValid() {
		relatedStatefulSet = *statefulSet
	}

	return relatedStatefulSet
}

func GetClusterRoleBinding(clusterRoleBindingName string) (relatedClusterRoleBinding rbacv1.ClusterRoleBinding) {

	clusterRoleBindingsClient := common.K8sClient.RbacV1().ClusterRoleBindings()
	clusterRoleBinding, err := clusterRoleBindingsClient.Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
	if err != nil {
		// Handle error by common the error and returning an empty list of related ClusterRoleBindings.
		log.Printf("[ERROR] Error getting clusterRoleBinding: %v", err)
		return
	}

	if reflect.ValueOf(clusterRoleBinding).IsValid() {
		relatedClusterRoleBinding = *clusterRoleBinding
	}

	return relatedClusterRoleBinding
}

func GetClusterRelatedResources(resourceKind string, resourceName string, namespace string) (relatedClusterServices common.RelatedClusterServices) {
	//
	log.Printf("[DEBUG] Attemping to parse Resource: %s of kind: %s related cluster services.\n", resourceName, resourceKind)

	common.CreateClusterClient()

	switch resourceKind {
	case "ConfigMap":
		relatedClusterServices = GetConfigMapRelatedWorkloads(resourceName)
	case "Secret":
		relatedClusterServices = GetSecretRelatedWorkloads(resourceName)
	case "ClusterRoleBinding":
		relatedClusterServices = GetClusterRoleBindingRelatedWorkloads(resourceName)
	case "ServiceAccount":
		relatedClusterServices = GetServiceAccountRelatedWorkloads(resourceName)
	case "ClusterRole":
		relatedClusterServices = GetClusterRoleRelatedWorkloads(resourceName)
	case "Deployment":
		relatedClusterServices = GetDeploymentRelatedResources(resourceName, namespace)
	case "DaemonSet":
		relatedClusterServices = GetDaemonSetRelatedResources(resourceName, namespace)
	case "StatefulSet":
		relatedClusterServices = GetStatefulSetRelatedResources(resourceName, namespace)
	default:
		log.Printf("[ERROR] Unknown resource kind %s", resourceKind)
	}

	return relatedClusterServices
}
