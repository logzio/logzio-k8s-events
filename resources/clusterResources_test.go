package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func workloadTestEnvVars() (workloadEnvVars []corev1.EnvVar) {
	workloadEnvVars = []corev1.EnvVar{
		{
			Name: "TOKEN_ENV_VAR",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-secret",
					},
				},
			},
		},
		{
			Name: "TOKEN_ENV_VAR",
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-configmap",
					},
				},
			},
		},
	}
	return workloadEnvVars
}
func workloadTestVolumes() (workloadVolumes []corev1.Volume) {

	workloadVolumes = []corev1.Volume{
		{
			Name: "test-secret-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "test-secret",
				},
			},
		},
		{
			Name: "test-configmap-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-configmap",
					},
				},
			},
		},
	}
	return workloadVolumes
}

func GetTestClusterRoleBinding(clusterRoleBindingName string) (clusterRoleBinding rbacv1.ClusterRoleBinding) {

	clusterRoleBinding = rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "test-serviceaccount",
				Namespace: "default",
				APIGroup:  "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "test-clusterrole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	return clusterRoleBinding
}
func GetTestPod() (pod corev1.Pod) {
	pod = corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "container-nginx",
					Image: "container-image-nginx",
					Env:   workloadTestEnvVars(),
				},
			},
			Volumes:            workloadTestVolumes(),
			ServiceAccountName: "test-serviceaccount",
		},
	}
	return pod
}
func GetTestDeployment() (deployment appsv1.Deployment) {
	deployment = appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: new(int32), // use new(int32) to create a pointer to an int32
			Selector: &metav1.LabelSelector{ // label selector for pods
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-nginx",
							Image: "container-image-nginx",
							Env:   workloadTestEnvVars(),
						},
					},
					Volumes:            workloadTestVolumes(),
					ServiceAccountName: "test-serviceaccount",
				},
			},
		}}
	return deployment
}
func GetTestDaemonSet() (relatedDaemonsets appsv1.DaemonSet) {
	daemonSet := appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-daemonset",
			Namespace: "default",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{ // label selector for pods
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-nginx",
							Image: "container-image-nginx",
							Env:   workloadTestEnvVars(),
						},
					},
					Volumes:            workloadTestVolumes(),
					ServiceAccountName: "test-serviceaccount",
				},
			},
		},
	}
	return daemonSet
}
func GetTestStatefulSet() (relatedStatefulsets appsv1.StatefulSet) {
	statefulSet := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-statefulset",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    new(int32), // use new(int32) to create a pointer to an int32
			ServiceName: "nginx",    // a service that governs this StatefulSet
			Selector: &metav1.LabelSelector{ // label selector for pods
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-nginx",
							Image: "container-image-nginx",
							Env:   workloadTestEnvVars(),
						},
					},
					Volumes:            workloadTestVolumes(),
					ServiceAccountName: "test-serviceaccount",
				},
			},
		},
		Status: appsv1.StatefulSetStatus{
			Replicas: *new(int32),
		},
	}
	return statefulSet
}

// Resources lists
func GetTestDaemonSets() (relatedDaemonsets []appsv1.DaemonSet) {
	daemonSet := GetTestDaemonSet()
	testDaemonSet := daemonSet
	testDaemonSet.Name = "test-daemonset-2"
	relatedDaemonsets = append(relatedDaemonsets, daemonSet)
	relatedDaemonsets = append(relatedDaemonsets, testDaemonSet)

	return relatedDaemonsets
}
func GetTestStatefulSets() (relatedStatefulsets []appsv1.StatefulSet) {
	statefulSet := GetTestStatefulSet()
	testStatefulSet := statefulSet
	testStatefulSet.Name = "test-statefulset-2"
	relatedStatefulsets = append(relatedStatefulsets, statefulSet)
	relatedStatefulsets = append(relatedStatefulsets, testStatefulSet)

	return relatedStatefulsets
}
func GetTestPods() (relatedPods []corev1.Pod) {
	pod := GetTestPod()
	testPod := pod
	testPod.Name = "test-pod-2"

	relatedPods = append(relatedPods, pod)
	relatedPods = append(relatedPods, testPod)

	return relatedPods
}
func GetTestDeployments() (relatedDeployments []appsv1.Deployment) {
	deployment := GetTestDeployment()
	testDeployment := deployment
	testDeployment.Name = "test-deployment-2"
	relatedDeployments = append(relatedDeployments, deployment)
	relatedDeployments = append(relatedDeployments, testDeployment)

	return relatedDeployments
}
func GetTestClusterRoleBindings() (relatedClusterRoleBindings []rbacv1.ClusterRoleBinding) {
	clusterRoleBinding := GetTestClusterRoleBinding("test-clusterrolebinding")
	testClusterRoleBinding := clusterRoleBinding
	testClusterRoleBinding.Name = "test-clusterrolebinding-2"
	relatedClusterRoleBindings = append(relatedClusterRoleBindings, clusterRoleBinding)
	relatedClusterRoleBindings = append(relatedClusterRoleBindings, testClusterRoleBinding)

	return relatedClusterRoleBindings
}
