package resources

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"main.go/common"
	"slices"
)

// subjectsRelatedWorkloads returns the workloads of every service account subject, each in its own namespace
func subjectsRelatedWorkloads(subjects []rbacv1.Subject) (relatedWorkloads common.RelatedClusterServices) {
	for _, subject := range subjects {
		if subject.Kind == "ServiceAccount" {
			relatedWorkloads = mergeRelatedWorkloads(relatedWorkloads, referencingWorkloads(serviceAccountIndex, subject.Namespace, subject.Name))
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
