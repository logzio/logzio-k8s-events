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
	ResourceObjects
}
type ResourceObjects struct {
	NewObject map[string]interface{} `json:"newObject,omitempty"`
	OldObject map[string]interface{} `json:"oldObject,omitempty"`
}
type EventStruct struct {
	EventType string `json:"eventType,omitempty"`
	KubernetesEvent
}
type LogEvent struct {
	Message                string `json:"message,omitempty"`
	EventStruct            `json:",omitempty"`
	Type                   string `json:"type,omitempty"`
	EnvironmentID          string `json:"env_id,omitempty"`
	RelatedClusterServices `json:"relatedClusterServices,omitempty"`
	ExtraFields            map[string]interface{} `json:"-"`
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

// IsValidList checks if an array is valid
func IsValidList(arrayFieldI []interface{}) (listField []interface{}, isValidArray bool) {
	// Logz.io doesn't support nested array objects well as they contain different data types
	for _, v := range arrayFieldI {
		_, isMap := v.(map[string]interface{})
		if !isMap {
			isValidArray = true
		}
	}
	return arrayFieldI, isValidArray
}

// ParseEventMessage parses event messages from the kubernetes event log
func ParseEventMessage(eventType string, resourceName string, resourceKind string, resourceNamespace string, newResourceVersion string, oldResourceVersions ...string) (msg string) {
	// Support cluster level resources
	namespacePart := ""
	if resourceNamespace != "" {
		namespacePart = " in namespace: " + resourceNamespace
	}

	if eventType == EventTypeModified {
		if len(oldResourceVersions) > 0 {
			oldResourceVersion := oldResourceVersions[0]
			msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s%s was updated from version: %s to new version: %s.", resourceName, resourceKind, namespacePart, oldResourceVersion, newResourceVersion)
		}
	} else if eventType == EventTypeDeleted {
		msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s%s with version: %s was deleted.", resourceName, resourceKind, namespacePart, newResourceVersion)

	} else if eventType == EventTypeAdded {
		msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s%s was added with version: %s.", resourceName, resourceKind, namespacePart, newResourceVersion)
	} else {
		log.Printf("[ERROR] Failed to parse resource event log message. Unknown eventType: %s.\n", eventType)
	}

	return msg
}

// FormatFieldName formats field name
func FormatFieldName(field string) (fieldName string) {
	fieldName = field
	// Check if the field contains a dot/slash/hyphen and replace it with underscore
	if strings.ContainsAny(field, "\\.-") {
		fieldName = strings.ReplaceAll(fieldName, ".", "_")
		fieldName = strings.ReplaceAll(fieldName, "\\", "_")
		fieldName = strings.ReplaceAll(fieldName, "-", "_")
	}
	return fieldName
}

// FormatFieldValue formats field value
func FormatFieldValue(value interface{}) (fieldValue interface{}) {
	fieldValue = value
	// Check if the field value is an array and parse it to a string
	arrayFieldI, ok := value.([]interface{})

	if ok {

		_, isValidArray := IsValidList(arrayFieldI)
		if !isValidArray {
			arrayNestedField, err := json.Marshal(arrayFieldI)
			if err != nil {
				log.Printf("\n[ERROR] Failed to parse array nested field: %s\nERROR:\n%v", arrayNestedField, err)
			}
			// Flatten the array nested field
			fieldValue = string(arrayNestedField)
		}

	}

	return fieldValue
}

// FormatFieldOverLimit formats field value if it is over limit
func FormatFieldOverLimit(fieldName string, fieldValue interface{}) (fieldOverLimit string, truncatedFieldValue interface{}) {
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

// IsEmptyMap checks if the map is empty
func IsEmptyMap(value interface{}) bool {
	isEmpty := false
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Map && v.Len() == 0 {
		isEmpty = true
	}
	return isEmpty
}

// parseLogzioLimits parses the logzio limits and replaces the fields with the new ones
func parseLogzioLimits(eventLog map[string]interface{}) (parsedLogEvent map[string]interface{}) {

	// Declare variables

	// Iterate over the log
	parsedLogEvent = eventLog

	if eventLog["newObject"] != nil {
		eventI := eventLog["newObject"].(map[string]interface{})
		if eventI["kind"] != nil {
			eventKind = eventI["kind"].(string)
		}
	}
	for field, value := range eventLog {
		// Check if the field contains a dot/slash/hyphen and replace it with underscore
		// Check if the field is empty

		if !reflect.ValueOf(value).IsValid() || value == nil || IsEmptyMap(value) {
			// Remove the empty or invalid/nil/struct{} field from the log
			delete(parsedLogEvent, field)
		}
		fieldName := FormatFieldName(field)
		if fieldName != field {
			// Rename the field
			parsedLogEvent[fieldName] = value
			// Remove the original field
			delete(parsedLogEvent, field)
		}
		maskedField, maskedValue := MaskSensitiveData(eventKind, fieldName, value)
		if !reflect.DeepEqual(value, maskedValue) {
			parsedLogEvent[maskedField] = maskedValue
			delete(parsedLogEvent, fieldName)
		}

		nestedField, ok := value.(map[string]interface{})

		// Check if the field is a nested map or struct
		if ok {
			parseLogzioLimits(nestedField)
		} else {
			{

				fieldValue := FormatFieldValue(value)
				if !reflect.DeepEqual(value, fieldValue) {
					// Add the field value to the parsed log
					parsedLogEvent[fieldName] = fieldValue
				}

				fieldOverLimit, truncatedFieldValue := FormatFieldOverLimit(fieldName, fieldValue)
				if fieldOverLimit != fieldName {
					parsedLogEvent[fieldOverLimit] = truncatedFieldValue
					delete(parsedLogEvent, fieldName)
				}

			}

		}

	}

	return parsedLogEvent
}

// hashData uses MD5 to hash the secret and return the hashed secret
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

// MaskSensitiveData masks sensitive data in the log event
func MaskSensitiveData(eventKind string, fieldName string, fieldValue interface{}) (maskedField string, maskedValue interface{}) {
	maskedValue = fieldValue // Initialize maskedValue to original fieldValue
	maskedField = fieldName  // Initialize maskedField to original fieldName

	// Array of field names to consider sensitive
	fieldsToMask := []string{"password", "secret", "token", "key", "access_token", "api_key", "api_secret", "api_token", "api_key_id", "api_secret_id", "api_token_id", "api_key_secret", "api_secret_key", "api_token_secret"}

	// Check if the field name is in the list of fields to mask, or has "_crt" in it, or is a secret data or last applied configuration
	if slices.Contains(fieldsToMask, fieldName) || strings.Contains(fieldName, "_crt") || (eventKind == "Secret" && (fieldName == "data" || fieldName == "kubectl_kubernetes_io_last_applied_configuration")) {
		// If the field is sensitive, mask the field value by hashing it
		stringValue := fmt.Sprintf("%v", fieldValue)
		maskedValue = hashData(stringValue)
		maskedField = fmt.Sprintf("%s_hashed", fieldName) // Append "_hashed" to the field name
	}

	// Return the masked field name and value
	return maskedField, maskedValue
}
