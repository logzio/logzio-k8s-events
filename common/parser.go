package common

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"k8s.io/utils/strings/slices"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// The renamed annotation key keeps its slash: FormatFieldName only rewrites dots, hyphens and backslashes
const lastAppliedConfigField = "kubectl_kubernetes_io/last_applied_configuration"

const defaultSensitiveEnvNames = `password|passwd|secret|token|credential|api[_-]?key|private[_-]?key|access[_-]?key|connection[_-]?string|(^|[_-])(key|pass|pwd|auth|dsn)($|[_-])`

var sensitiveEnvName = mustBuildSensitiveEnvName("")

// buildSensitiveEnvName extends the default sensitive env var name pattern with an additional regex
func buildSensitiveEnvName(additionalPattern string) (*regexp.Regexp, error) {
	pattern := defaultSensitiveEnvNames
	if additionalPattern = strings.TrimSpace(additionalPattern); additionalPattern != "" {
		pattern += "|(?:" + additionalPattern + ")"
	}
	return regexp.Compile("(?i)" + pattern)
}

func mustBuildSensitiveEnvName(additionalPattern string) *regexp.Regexp {
	pattern, err := buildSensitiveEnvName(additionalPattern)
	if err != nil {
		log.Fatalf("\n[FATAL] Invalid regex in SENSITIVE_ENV_NAMES environment variable.\nERROR: %v\n", err)
	}
	return pattern
}

// hashKey keys the HMAC of masked values, so they can't be reversed with a dictionary without it
var hashKey = randomHashKey()

func randomHashKey() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("\n[FATAL] Failed to generate a hash key for masking sensitive data.\nERROR: %v\n", err)
	}
	return key
}

// ConfigureMasking adds the regex in SENSITIVE_ENV_NAMES to the env var names whose values are masked,
// and sets MASKING_HASH_KEY as the hash key, which keeps hashes comparable across restarts
func ConfigureMasking() {
	sensitiveEnvName = mustBuildSensitiveEnvName(os.Getenv("SENSITIVE_ENV_NAMES"))
	if key := os.Getenv("MASKING_HASH_KEY"); key != "" {
		hashKey = []byte(key)
	}
}

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
	// More accurate message for cluster level resources
	inNamespaceMsg := ""
	if resourceNamespace != "" {
		inNamespaceMsg = " in namespace: " + resourceNamespace
	}

	if eventType == EventTypeModified {
		if len(oldResourceVersions) > 0 {
			oldResourceVersion := oldResourceVersions[0]
			msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s%s was updated from version: %s to new version: %s.", resourceName, resourceKind, inNamespaceMsg, oldResourceVersion, newResourceVersion)
		}
	} else if eventType == EventTypeDeleted {
		msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s%s with version: %s was deleted.", resourceName, resourceKind, inNamespaceMsg, newResourceVersion)

	} else if eventType == EventTypeAdded {
		msg = fmt.Sprintf("[EVENT] Resource: %s of kind: %s%s was added with version: %s.", resourceName, resourceKind, inNamespaceMsg, newResourceVersion)
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

// resourceKind returns the kind of the event's new object, or an empty string if it has none
func resourceKind(eventLog map[string]interface{}) string {
	for _, objectField := range []string{"newObject", "oldObject"} {
		object, _ := eventLog[objectField].(map[string]interface{})
		if kind, _ := object["kind"].(string); kind != "" {
			return kind
		}
	}
	return ""
}

// maskLastAppliedEnvVars masks env vars inside the last-applied-configuration annotation, which embeds the whole manifest as JSON
func maskLastAppliedEnvVars(fieldName string, value interface{}) (maskedValue interface{}, changed bool) {
	raw, ok := value.(string)
	if !ok || fieldName != lastAppliedConfigField {
		return value, false
	}
	var manifest interface{}
	if err := json.Unmarshal([]byte(raw), &manifest); err != nil {
		return hashData(raw), true
	}
	if !maskEnvVars(manifest) {
		return value, false
	}
	masked, err := json.Marshal(manifest)
	if err != nil {
		return hashData(raw), true
	}
	return string(masked), true
}

// maskEnvVars hashes the value of sensitive-named env vars nested in arrays, before the arrays are flattened to strings.
// It reports whether anything was masked.
func maskEnvVars(value interface{}) (masked bool) {
	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			masked = maskEnvVars(item) || masked
		}
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok && sensitiveEnvName.MatchString(name) {
			if envValue, ok := v["value"]; ok {
				v["value_hashed"] = hashData(envValue)
				delete(v, "value")
				masked = true
			}
		}
		for _, nested := range v {
			masked = maskEnvVars(nested) || masked
		}
	}
	return masked
}

// parseLogzioLimits parses the logzio limits and replaces the fields with the new ones
func parseLogzioLimits(eventLog map[string]interface{}) map[string]interface{} {
	return parseLogzioLimitsOfKind(eventLog, resourceKind(eventLog))
}

func parseLogzioLimitsOfKind(eventLog map[string]interface{}, eventKind string) (parsedLogEvent map[string]interface{}) {
	parsedLogEvent = eventLog

	// Snapshot the keys: ranging over the map while renaming fields could revisit the added ones, e.g. hash tls_crt_hashed again
	fields := make([]string, 0, len(eventLog))
	for field := range eventLog {
		fields = append(fields, field)
	}
	for _, field := range fields {
		value := eventLog[field]
		// Check if the field contains a dot/slash/hyphen and replace it with underscore
		// Check if the field is empty

		if !reflect.ValueOf(value).IsValid() || value == nil || IsEmptyMap(value) {
			// Remove the empty or invalid/nil/struct{} field from the log
			delete(parsedLogEvent, field)
			continue
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
			continue
		}

		nestedField, ok := value.(map[string]interface{})

		// Check if the field is a nested map or struct
		if ok {
			parseLogzioLimitsOfKind(nestedField, eventKind)
		} else {
			{
				maskEnvVars(value)
				if eventKind != "Secret" {
					if maskedValue, changed := maskLastAppliedEnvVars(fieldName, value); changed {
						value = maskedValue
						parsedLogEvent[fieldName] = value
					}
				}
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

// hashData hashes sensitive data with HMAC-SHA256
func hashData(data interface{}) string {
	mac := hmac.New(sha256.New, hashKey)
	mac.Write([]byte(fmt.Sprint(data)))
	return hex.EncodeToString(mac.Sum(nil))
}

// MaskSensitiveData masks sensitive data in the log event
func MaskSensitiveData(eventKind string, fieldName string, fieldValue interface{}) (maskedField string, maskedValue interface{}) {
	maskedValue = fieldValue // Initialize maskedValue to original fieldValue
	maskedField = fieldName  // Initialize maskedField to original fieldName

	// Array of field names to consider sensitive
	fieldsToMask := []string{"password", "secret", "token", "key", "access_token", "api_key", "api_secret", "api_token", "api_key_id", "api_secret_id", "api_token_id", "api_key_secret", "api_secret_key", "api_token_secret"}

	// Check if the field name is in the list of fields to mask, or has "_crt" in it, or is a secret data or last applied configuration
	if slices.Contains(fieldsToMask, fieldName) || strings.Contains(fieldName, "_crt") || (eventKind == "Secret" && (fieldName == "data" || fieldName == lastAppliedConfigField)) {
		// If the field is sensitive, mask the field value by hashing it
		stringValue := fmt.Sprintf("%v", fieldValue)
		maskedValue = hashData(stringValue)
		maskedField = fmt.Sprintf("%s_hashed", fieldName) // Append "_hashed" to the field name
	}

	// Return the masked field name and value
	return maskedField, maskedValue
}
