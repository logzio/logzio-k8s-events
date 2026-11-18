package resources

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"main.go/common"
)

// deploymentWith mounts a Secret and a ConfigMap that are both named secret
func deploymentWith(name, namespace, serviceAccount, secret string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccount,
			Volumes: []corev1.Volume{
				{Name: "s", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: secret}}},
				{Name: "c", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: secret}}}},
			},
		}}},
	}
}

func clusterRoleBinding(name, role string, subjects ...rbacv1.Subject) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: name}, RoleRef: rbacv1.RoleRef{Name: role}, Subjects: subjects}
}

func serviceAccountSubject(name, namespace string) rbacv1.Subject {
	return rbacv1.Subject{Kind: "ServiceAccount", Name: name, Namespace: namespace}
}

func useFakeClient(t *testing.T, objects ...runtime.Object) {
	t.Helper()
	// Run the one-time client setup first, so it can't replace the fake when the tests run inside a cluster
	common.CreateClusterClient()
	previous := common.K8sClient
	t.Cleanup(func() { common.K8sClient = previous })
	common.K8sClient = fake.NewSimpleClientset(objects...)
}

func TestRelatedWorkloadsAreLimitedToTheResourceNamespace(t *testing.T) {
	useFakeClient(t,
		deploymentWith("app-a", "team-a", "default", "db"),
		deploymentWith("app-b", "team-b", "default", "db"),
	)

	for kind, related := range map[string]common.RelatedClusterServices{
		"Secret":         GetClusterRelatedResources("Secret", "db", "team-a"),
		"ConfigMap":      GetClusterRelatedResources("ConfigMap", "db", "team-a"),
		"ServiceAccount": GetClusterRelatedResources("ServiceAccount", "default", "team-a"),
	} {
		if !reflect.DeepEqual(related.Deployments, []string{"app-a"}) {
			t.Errorf("%s in team-a: expected only app-a, got %v", kind, related.Deployments)
		}
	}
}

func TestClusterRoleRelatedWorkloadsMergesAllSubjects(t *testing.T) {
	useFakeClient(t,
		deploymentWith("app-a", "team-a", "reader", ""),
		deploymentWith("app-b", "team-b", "reader", ""),
		deploymentWith("app-c", "team-c", "reader", ""),
		clusterRoleBinding("readers", "view", serviceAccountSubject("reader", "team-a"), serviceAccountSubject("reader", "team-b")),
		clusterRoleBinding("other", "edit", serviceAccountSubject("reader", "team-c")),
	)

	for kind, related := range map[string]common.RelatedClusterServices{
		"ClusterRole":        GetClusterRelatedResources("ClusterRole", "view", ""),
		"ClusterRoleBinding": GetClusterRelatedResources("ClusterRoleBinding", "readers", ""),
	} {
		if !reflect.DeepEqual(related.Deployments, []string{"app-a", "app-b"}) {
			t.Errorf("%s: expected the workloads of both subjects and not team-c's, got %v", kind, related.Deployments)
		}
	}
}

func TestWorkloadRelatedClusterRolesMatchTheServiceAccountNamespace(t *testing.T) {
	useFakeClient(t,
		deploymentWith("app-a", "team-a", "default", ""),
		clusterRoleBinding("team-a-view", "view", serviceAccountSubject("default", "team-a")),
		clusterRoleBinding("team-b-admin", "admin", serviceAccountSubject("default", "team-b")),
	)

	related := GetClusterRelatedResources("Deployment", "app-a", "team-a")

	if !reflect.DeepEqual(related.ClusterRoleBindings, []string{"team-a-view"}) || !reflect.DeepEqual(related.ClusterRoles, []string{"view"}) {
		t.Errorf("expected only team-a's binding and role, got bindings %v and roles %v", related.ClusterRoleBindings, related.ClusterRoles)
	}
}

func TestDeletedClusterRoleBindingHasNoRelatedWorkloads(t *testing.T) {
	useFakeClient(t, deploymentWith("app-a", "team-a", "reader", ""))

	related := GetClusterRelatedResources("ClusterRoleBinding", "deleted", "")

	if !reflect.DeepEqual(related, common.RelatedClusterServices{}) {
		t.Errorf("expected no related workloads for a deleted binding, got %+v", related)
	}
}
