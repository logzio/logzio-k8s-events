# Logz.io Kubernetes Events

The logzio-k8s-events integration was made to send data about deployment events in the cluster, and how they affect the resources in the cluster.

It uses the Kubernetes informer official SDK to watch for deployment events in the cluster.
The events are getting parsed and enriched using Kubernetes SDK to correlate them with resources that are being effected by the deployment. 
They are then sent to Logz.io using Logz.io GoLang SDK. 

Currently supported resource kinds are Deployment, Daemonset, Statefulset, ConfigMap, Secret, Service Account, Cluster Role & Cluster Role Binding.

It can be deployed using the [logzio-k8s-events Helm chart](https://github.com/logzio/logzio-helm/tree/master/charts/logzio-k8s-events).

# Tests

Each package has test files that are relevant to each functionality, running tests can be done using the following command:
```
go test .
```

The [tests.yml](https://github.com/logzio/logzio-k8s-events/blob/master/.github/workflows/tests.yml) workflow runs when opening a pull request to validate the tests passes. 

# Architecture 
[![Architecture](./K8SDeploymentEvents.svg)

## Change log
 - **0.0.1**:
    - Initial release.
