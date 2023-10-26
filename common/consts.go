package common

const (
	EventTypeDeleted  = "DELETED"
	EventTypeModified = "MODIFIED"
	EventTypeAdded    = "ADDED"
)

const (
	DefaultListener = "https://listener.logz.io:8071"
	DefaultLogType  = "logzio-k8s-events"
)
const (
	Metadata           = "metadata"
	ManagedFields      = "managedFields"
	ResourceVersion    = "resourceVersion"
	Annotations        = "annotations"
	DeploymentRevision = "deployment.kubernetes.io/revision"
	Status             = "status"
)
