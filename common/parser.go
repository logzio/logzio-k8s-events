package common

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"k8s.io/utils/strings/slices"
	"log"
	"reflect"
	"strings"
)

var eventKind string

type KubernetesMetadata struct {
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}
type KubernetesEvent struct {
	Kind               string `json:"kind,omitempty"`
	KubernetesMetadata `json:"metadata,omitempty"`
}
type EventStruct struct {
	EventType string                 `json:"eventType,omitempty"`
	NewObject map[string]interface{} `json:"newObject,omitempty"`
	OldObject map[string]interface{} `json:"oldObject,omitempty"`
}
type LogEvent struct {
	Message                string `json:"message,omitempty"`
	EventStruct            `json:",omitempty"`
	Type                   string                 `json:"type,omitempty"`
	EnvironmentID          string                 `json:"env_id,omitempty"`
	Log                    map[string]interface{} `json:"log,omitempty"`
	RelatedClusterServices `json:"relatedClusterServices,omitempty"`
}

type RelatedClusterServices struct {
	Deployments         []string `json:"deployments,omitempty"`
	DaemonSets          []string `json:"daemonsets,omitempty"`
	StatefulSets        []string `json:"statefulsets,omitempty"`
	Pods                []string `json:"pods,omitempty"`
	Secrets             []string `json:"secrets,omitempty"`
	ServiceAccounts     []string `json:"serviceaccounts,omitempty"`
	ConfigMaps          []string `json:"configmaps,omitempty"`
	ClusterRoles        []string `json:"clusterroles,omitempty"`
	ClusterRoleBindings []string `json:"clusterrolebindings,omitempty"`
}

func isValidList(arrayFieldI []interface{}) (listField []interface{}, isValidArray bool) {
	// Logz.io doesn't support nested array objects well as they contain different data types
	for _, v := range arrayFieldI {
		_, isMap := v.(map[string]interface{})
		if !isMap {
			isValidArray = true
		}
	}
	return arrayFieldI, isValidArray
}

func ParseEventMessage(eventType string, resourceName string, resourceKind string, resourceNamespace string, newResourceVersion string, oldResourceVersions ...string) (msg string) {
	if eventType == "MODIFIED" {
		if len(oldResourceVersions) > 0 {
			oldResourceVersion := oldResourceVersions[0]
			msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s in namespace: %s was updated from version: %s to new version: %s.\n", resourceName, resourceKind, resourceNamespace, oldResourceVersion, newResourceVersion)
		}
	} else if eventType == "DELETED" {
		msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s in namespace: %s with version: %s was deleted.\n", resourceName, resourceKind, resourceNamespace, newResourceVersion)

	} else if eventType == "ADDED" {
		msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s in namespace: %s was added with version: %s.\n", resourceName, resourceKind, resourceNamespace, newResourceVersion)
	} else {
		log.Printf("[ERROR] Failed to parse resource event log message. Unknown eventType: %s.\n", eventType)
	}
	return msg
}

func formatFieldName(field string) (fieldName string) {
	fieldName = field
	// Check if the field contains a dot/slash/hyphen and replace it with underscore
	if strings.ContainsAny(field, "/.-") {
		fieldName = strings.Replace(field, ".", "_", -1)
		fieldName = strings.Replace(fieldName, "/", "_", -1)
		fieldName = strings.Replace(fieldName, "-", "_", -1)
	}
	return fieldName
}
func formatFieldValue(value interface{}) (fieldValue interface{}) {
	fieldValue = value
	// Check if the field value is an array and parse it to a string
	arrayFieldI, ok := value.([]interface{})

	if ok {

		_, isValidArray := isValidList(arrayFieldI)
		if !isValidArray {
			arrayNestedField, err := json.Marshal(arrayFieldI)
			if err != nil {
				log.Printf("\n[ERROR] Failed to parse array nested field. %s\nERROR:\n%v", err)
			}
			// Flatten the array nested field
			fieldValue = string(arrayNestedField)
		}

	}

	return fieldValue
}
func formatFieldOverLimit(fieldName string, fieldValue interface{}) (fieldOverLimit string, truncatedFieldValue interface{}) {
	fieldOverLimit = fieldName
	truncatedFieldValue = fieldValue
	var valueLengthLimit = 32700
	// Check if the field value length is over the limit
	if len(fmt.Sprint(fieldValue)) >= valueLengthLimit && !strings.HasSuffix(fieldName, "_overLimit") {
		// Add the field to the fieldsOverLimit slice, so it will be ignored in the next iteration
		// Truncate the field value to the limit
		truncatedFieldValue = fmt.Sprintf("%s", fmt.Sprint(fieldValue)[:valueLengthLimit])
		// Rename the field if it passes value length limit
		fieldOverLimit = fmt.Sprintf("%s_overLimit", fieldName)
		// Add the field to the fieldsOverLimit slice, so it will be ignored in the next iteration
	}
	return fieldOverLimit, truncatedFieldValue

}
func parseLogzioLimits(eventLog map[string]interface{}) (parsedLogEvent map[string]interface{}) {

	// Declare variables

	// Iterate over the log
	parsedLogEvent = eventLog
	eventI := eventLog["newObject"]
	if eventI != nil {
		eventKind = eventI.(map[string]interface{})["kind"].(string)
	}
	for field, value := range eventLog {
		// Check if the field contains a dot
		fieldName := formatFieldName(field)
		if fieldName != field {
			// Rename the field
			parsedLogEvent[fieldName] = value
			// Remove the original field
			delete(parsedLogEvent, field)
		}
		maskedField, maskedValue := maskSensitiveData(eventKind, fieldName, value)
		if !reflect.DeepEqual(value, maskedValue) {
			parsedLogEvent[maskedField] = maskedValue
			delete(parsedLogEvent, fieldName)
		}

		nestedField, ok := value.(map[string]interface{})

		// Check if the field is a nested map or struct
		if ok {
			parseLogzioLimits(nestedField)
		} else {

			// Check if the field is empty
			if !reflect.ValueOf(value).IsValid() || value == nil || value == struct{}{} {
				// Remove the empty or invalid/nil/struct{} field from the log
				delete(parsedLogEvent, fieldName)
			} else {

				fieldValue := formatFieldValue(value)
				if !reflect.DeepEqual(value, fieldValue) {
					// Add the field value to the parsed log
					parsedLogEvent[fieldName] = fieldValue
				}

				fieldOverLimit, truncatedFieldValue := formatFieldOverLimit(fieldName, fieldValue)
				if fieldOverLimit != fieldName {
					parsedLogEvent[fieldOverLimit] = truncatedFieldValue
					delete(parsedLogEvent, fieldName)
				}

			}

		}

	}

	return parsedLogEvent
}
func hashData(data interface{}) (hashedData string) {
	// Create a new MD5 hash object
	hash := md5.New()

	// Write the secret to the hash object
	hash.Write([]byte(data.(string)))

	// Get the MD5 hash of the secret
	hashSum := hash.Sum(nil)

	// Convert the MD5 hash to a string
	hashedData = fmt.Sprintf("%x", hashSum)

	return hashedData
}
func maskSensitiveData(eventKind string, fieldName string, fieldValue interface{}) (maskedField string, maskedValue interface{}) {
	maskedValue = fieldValue
	maskedField = fieldName
	fieldsToMask := []string{"password", "secret", "token", "key", "access_token", "api_key", "api_secret", "api_token", "api_key_id", "api_secret_id", "api_token_id", "api_key_secret", "api_secret_key", "api_token_secret"}
	if slices.Contains(fieldsToMask, fieldName) || strings.Contains(fieldName, "_crt") || (eventKind == "Secret" && (fieldName == "data" || fieldName == "kubectl_kubernetes_io_last_applied_configuration")) {
		// Mask the field from the log
		stringValue := fmt.Sprintf("%v", fieldValue)
		maskedValue = hashData(stringValue)
		maskedField = fmt.Sprintf("%s_hashed", fieldName)

	}
	return maskedField, maskedValue
}
