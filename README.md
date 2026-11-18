# Logz.io Kubernetes Events

The logzio-k8s-events integration was made to send data about deployment events in the cluster, and how they affect the resources in the cluster.

It uses the Kubernetes informer official SDK to watch for deployment events in the cluster.
The events are getting parsed and enriched using Kubernetes SDK to correlate them with resources that are being effected by the deployment. 
They are then sent to Logz.io using Logz.io GoLang SDK. 

Currently supported resource kinds are Deployment, Daemonset, Statefulset, ConfigMap, Secret, Service Account, Cluster Role & Cluster Role Binding.

It can be deployed using the [logzio-k8s-events Helm chart](https://github.com/logzio/logzio-helm/tree/master/charts/logzio-k8s-events).

# Configuration

The integration is configured with environment variables:

| Variable | Description | Default |
|---|---|---|
| `LOGZIO_TOKEN` | **Required.** Logz.io logs shipping token. | |
| `LOGZIO_LISTENER` | Logz.io listener URL. | `https://listener.logz.io:8071` |
| `ENV_ID` | Environment ID added to each log as `env_id`. | |
| `LOG_TYPE` | Log type added to each log as `type`. | `logzio-k8s-events` |
| `SENSITIVE_ENV_NAMES` | Regex of additional env var names whose values are hashed, e.g. `internal_id\|^ssn$`. It extends the built-in pattern (`password`, `secret`, `token`, `api_key`, ...) and never replaces it. | |
| `MASKING_HASH_KEY` | Key for hashing sensitive values with HMAC-SHA256. Set it to keep hashes comparable across restarts. | Random key per process |
| `DEBUG` | Print each full event to stdout. By default only a one-line summary is printed. | `false` |

# Tests

Each package has test files that are relevant to each functionality, running tests can be done using the following command:
```
go test ./...
```

The [tests.yaml](https://github.com/logzio/logzio-k8s-events/blob/master/.github/workflows/tests.yaml) workflow runs when opening a pull request to validate the tests pass. 

# Architecture 
![Architecture](./architecture.svg)

## Change log
 - **1.0.0**:
   - **Breaking changes**:
     - `metadata.managedFields` is no longer shipped.
     - Masked values are hashed with HMAC-SHA256 (64 hex characters) instead of MD5. Set `MASKING_HASH_KEY` for hashes that stay comparable across restarts.
     - Only a one-line summary of each event is printed to stdout. Set `DEBUG=true` to print the full event.
   - Fix a race that could ship Secret `data` unmasked.
   - Hash the values of sensitive container env vars, including inside the `kubectl.kubernetes.io/last-applied-configuration` annotation.
   - Hash a Secret's `last-applied-configuration` annotation.
   - Mask Secrets deleted while the watch was down.
   - Add `SENSITIVE_ENV_NAMES` to extend the masked env var names.
   - Run the Docker image as a non-root user (UID `10001`).
   - Publish the Docker image with SBOM and provenance attestations.
   - Upgrade `golang.org/x/net` to `v0.62.0`.
   - Shut down gracefully on `SIGTERM`, flushing the remaining logs.
   - Match related workloads in the resource's namespace, and merge the workloads of all subjects of a ClusterRole or ClusterRoleBinding.
   - Fix data races on shared globals.
   - Set `@timestamp` to the time of the event.
   - Batch logs instead of sending each event separately.
   - Upgrade GoLang version to `v1.27.4`, and docker images to `golang:1.27.4-alpine3.24` and `alpine:3.24`.
   - Upgrade `k8s.io/client-go`, `k8s.io/api` and `k8s.io/apimachinery` to `v0.37.2`.
   - Upgrade `github.com/logzio/logzio-go` to `v1.0.10`.
   - Upgrade GitHub Actions, and add `go vet`, race detection and `govulncheck` to CI.
 - **0.0.4**:
   - Upgrade `github.com/logzio/logzio-go` to `v1.0.9`
   - Upgrade GoLang version to `v1.23.0`
   - Upgrade GoLang docker image to `golang:1.23.0-alpine3.20`   
 - **0.0.3**:
    - Upgrade GoLang version to `v1.22.3`
    - Upgrade docker image to `alpine:3.20`
    - Upgrade GoLang docker image to `golang:1.22.3-alpine3.20`
 - **0.0.2**:
    - Ignore internal event changes.
 - **0.0.1**:
    - Initial release.