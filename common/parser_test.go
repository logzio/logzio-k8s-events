package common

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func GetTestEventLog() (eventLog map[string]interface{}) {
	eventLogData := []byte(`{
  "de.dot": "testz",
  "pokemon": null,
  "relatedClusterServices": {
    "statefulsets": [
      "prometheus-kube-prometheus-stack-prometheus-statefulset"
    ],
    "pods": [
      "prometheus-kube-prometheus-stack-prometheus-pod-0"
    ],
    "daemonsets": [
      "prometheus-kube-prometheus-stack-prometheus-daemonset"
    ],
    "deployments": [
      "prometheus-kube-prometheus-stack-prometheus-deployment"
    ]
  },
  "eventType": "MODIFIED",
  "message": "[EVENT] Resource: prometheus-kube-prometheus-stack-prometheus of kind: Secret in namespace: monitoring was updated from version: 26515170 to new version: 27160250.\n",
  "type": "logzio-k8s-events",
  "env_id": "logzio-staging",
  "newObject": {
    "data": "deff2df0cb0dac69be67c39d9e769e0f",
    "metadata": {
      "uid": "c3c91497-6570-48f0-a5df-ea42ad36442b",
      "managedFields": [
        {
          "apiVersion": "v1",
          "fieldsType": "FieldsV1",
          "fieldsV1": {
            "f:data": {
              ".": {},
              "f:prometheus.yaml.gz": {}
            },
            "f:metadata": {
              "f:annotations": {
                ".": {},
                "f:generated": {}
              },
              "f:labels": {
                ".": {},
                "f:managed-by": {}
              },
              "f:ownerReferences": {
                ".": {},
                "k:{\"uid\":\"90d247d0-b3e4-43e2-a2ec-bc0f4d7aad82\"}": {}
              }
            },
            "f:type": {}
          },
          "manager": "PrometheusOperator",
          "operation": "Update",
          "time": "2023-09-20T10:27:04Z"
        }
      ],
      "resourceVersion": "27160250",
      "creationTimestamp": "2023-07-31T12:35:01Z",
      "name": "prometheus-kube-prometheus-stack-prometheus",
      "namespace": "monitoring",
      "annotations": {
        "generated": "true"
      },
      "labels": {
        "managed_by": "prometheus-operator"
      },
      "ownerReferences": [
        {
          "apiVersion": "monitoring.coreos.com/v1",
          "blockOwnerDeletion": true,
          "controller": true,
          "kind": "Prometheus",
          "name": "kube-prometheus-stack-prometheus",
          "uid": "90d247d0-b3e4-43e2-a2ec-bc0f4d7aad82"
        }
      ]
    },
    "apiVersion": "v1",
    "kind": "Secret",
    "type": "Opaque"
  },
  "oldObject": {
    "data": "290f52ba4cf36fb2692f7715d929e02c",
    "metadata": {
      "uid": "c3c91497-6570-48f0-a5df-ea42ad36442b",
      "managedFields": [
        {
          "apiVersion": "v1",
          "fieldsType": "FieldsV1",
          "fieldsV1": {
            "f:data": {
              ".": {},
              "f:prometheus.yaml.gz": {}
            },
            "f:metadata": {
              "f:annotations": {
                ".": {},
                "f:generated": {}
              },
              "f:labels": {
                ".": {},
                "f:managed-by": {}
              },
              "f:ownerReferences": {
                ".": {},
                "k:{\"uid\":\"90d247d0-b3e4-43e2-a2ec-bc0f4d7aad82\"}": {}
              }
            },
            "f:type": {}
          },
          "manager": "PrometheusOperator",
          "operation": "Update",
          "time": "2023-09-19T11:57:26Z"
        }
      ]
    },
    "resourceVersion": "26515170",
    "creationTimestamp": "2023-07-31T12:35:01Z",
    "name": "prometheus-kube-prometheus-stack-prometheus",
    "namespace": "monitoring",
    "annotations": {
      "generated": "true"
    },
    "labels": {
      "managed_by": "prometheus-operator"
    },
    "ownerReferences": [
      {
        "apiVersion": "monitoring.coreos.com/v1",
        "blockOwnerDeletion": true,
        "controller": true,
        "kind": "Prometheus",
        "name": "kube-prometheus-stack-prometheus",
        "uid": "90d247d0-b3e4-43e2-a2ec-bc0f4d7aad82"
      }
    ],
    "apiVersion": "v1",
    "kind": "Secret",
    "type": "Opaque"
  }
}`)
	err := json.Unmarshal(eventLogData, &eventLog)
	if err != nil {
		log.Printf("Error unmarshalling event log: %v", err)
		return nil
	}
	return eventLog
}
func TestParseLogzioLimits(t *testing.T) {
	eventLog := GetTestEventLog()
	parsedLogEvent := eventLog

	eventKind := resourceKind(eventLog)

	for field, value := range eventLog {
		var fieldName string
		var fieldValue interface{}
		t.Run("testRemoveEmptyField", func(t *testing.T) {
			if !reflect.ValueOf(value).IsValid() || value == nil || IsEmptyMap(value) {
				delete(parsedLogEvent, field)
				t.Logf("Successfully removed empty/invalid field %s from the log.", field)
			}
		})

		t.Run("testRenameField", func(t *testing.T) {
			fieldName = FormatFieldName(field)
			if fieldName != field {
				parsedLogEvent[fieldName] = value
				delete(parsedLogEvent, field)
				t.Logf("Successfully renamed badly named field: %s to: %s in the log.", field, fieldName)
			}
		})

		t.Run("testMaskSensitiveData", func(t *testing.T) {
			maskedField, maskedValue := MaskSensitiveData(eventKind, fieldName, value)
			if !reflect.DeepEqual(value, maskedValue) {
				parsedLogEvent[maskedField] = maskedValue
				delete(parsedLogEvent, fieldName)
				t.Logf("Successfully masked field: %s in the log.", field)
			}
		})

		t.Run("testNestedMapOrStruct", func(t *testing.T) {
			nestedField, ok := value.(map[string]interface{})
			if ok {
				t.Logf("The field: %s is a nested map/struct, so parsing its limits.", field)
				parseLogzioLimitsOfKind(nestedField, eventKind)
			}
		})

		t.Run("testFormatFieldValue", func(t *testing.T) {
			fieldValue = FormatFieldValue(value)
			if !reflect.DeepEqual(value, fieldValue) {
				parsedLogEvent[fieldName] = fieldValue
				t.Logf("Successfully formatted field: %s with value: %s to the parsed log.", fieldName, fieldValue)
			}
		})

		t.Run("testFormatFieldOverLimit", func(t *testing.T) {
			fieldOverLimit, truncatedFieldValue := FormatFieldOverLimit(fieldName, fieldValue)
			if fieldOverLimit != fieldName {
				parsedLogEvent[fieldOverLimit] = truncatedFieldValue
				delete(parsedLogEvent, fieldName)
				t.Logf("Successfully truncated field: %s with value: %s to the parsed log.", fieldOverLimit, truncatedFieldValue)
			}
		})
	}
}

func TestParseLogzioLimitsMasksSecretDataUnderConcurrentEvents(t *testing.T) {
	const plaintext = "super-secret-value"
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			parseLogzioLimits(map[string]interface{}{
				"newObject": map[string]interface{}{"kind": "ConfigMap", "data": map[string]interface{}{"a": "b"}},
			})
		}()
		go func() {
			defer wg.Done()
			parsed := parseLogzioLimits(map[string]interface{}{
				"newObject": map[string]interface{}{"kind": "Secret", "data": map[string]interface{}{"password": plaintext}},
			})
			if strings.Contains(string(mustMarshal(t, parsed)), plaintext) {
				t.Errorf("secret data was not masked: %v", parsed)
			}
		}()
	}
	wg.Wait()
}

func TestParseLogzioLimitsMasksSensitiveEnvVars(t *testing.T) {
	parsed := parseLogzioLimits(map[string]interface{}{
		"newObject": map[string]interface{}{
			"kind": "Deployment",
			"spec": map[string]interface{}{"template": map[string]interface{}{"spec": map[string]interface{}{
				"containers": []interface{}{map[string]interface{}{
					"name": "app",
					"env": []interface{}{
						map[string]interface{}{"name": "DB_PASSWORD", "value": "hunter2"},
						map[string]interface{}{"name": "LOG_LEVEL", "value": "debug"},
						map[string]interface{}{"name": "API_TOKEN", "valueFrom": map[string]interface{}{"secretKeyRef": map[string]interface{}{"name": "s"}}},
					},
				}},
			}}},
		},
	})

	containers, ok := parsed["newObject"].(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].(string)
	if !ok {
		t.Fatalf("containers should stay a JSON string, got %T", parsed)
	}
	if strings.Contains(containers, "hunter2") {
		t.Errorf("sensitive env value leaked: %s", containers)
	}
	if !strings.Contains(containers, `"value_hashed":"`+hashData("hunter2")+`"`) {
		t.Errorf("expected hashed value in %s", containers)
	}
	if !strings.Contains(containers, `"value":"debug"`) || !strings.Contains(containers, "secretKeyRef") {
		t.Errorf("non-sensitive env entries must be untouched: %s", containers)
	}
}

func TestParseLogzioLimitsMasksEnvVarsInLastAppliedConfiguration(t *testing.T) {
	annotation := `{"spec":{"containers":[{"env":[{"name":"DB_PASSWORD","value":"hunter2"},{"name":"MONKEY_MODE","value":"on"}]}]}}`
	parsed := parseLogzioLimits(map[string]interface{}{
		"oldObject": map[string]interface{}{"kind": "Deployment", "metadata": map[string]interface{}{
			"annotations": map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": annotation},
		}},
	})

	masked := string(mustMarshal(t, parsed))
	if strings.Contains(masked, "hunter2") {
		t.Errorf("env value leaked through last-applied annotation: %s", masked)
	}
	if !strings.Contains(masked, "MONKEY_MODE") || !strings.Contains(masked, `\"value\":\"on\"`) {
		t.Errorf("non-sensitive env var must be untouched: %s", masked)
	}
}

func TestMaskSensitiveDataHashesSecretLastAppliedConfiguration(t *testing.T) {
	field, value := MaskSensitiveData("Secret", "kubectl_kubernetes_io/last_applied_configuration", `{"data":{"password":"cGFzcw=="}}`)

	if field != "kubectl_kubernetes_io/last_applied_configuration_hashed" || strings.Contains(value.(string), "cGFzcw") {
		t.Errorf("secret last-applied configuration was not hashed: %s=%v", field, value)
	}
}

func TestParseLogzioLimitsDoesNotLeakMaskedFields(t *testing.T) {
	longValue := strings.Repeat("x", 40000)
	parsed := parseLogzioLimits(map[string]interface{}{
		"newObject": map[string]interface{}{"kind": "Secret", "metadata": map[string]interface{}{
			"annotations": map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": longValue},
		}},
		"token": []interface{}{map[string]interface{}{"a": "plaintext-token"}},
	})

	masked := string(mustMarshal(t, parsed))
	if strings.Contains(masked, "xxxx") || strings.Contains(masked, "plaintext-token") || strings.Contains(masked, "_overLimit") {
		t.Errorf("masked field leaked next to its hash: %.300s", masked)
	}
	if !strings.Contains(masked, "last_applied_configuration_hashed") || !strings.Contains(masked, "token_hashed") {
		t.Errorf("expected hashed fields in %.300s", masked)
	}
}

func TestParseLogzioLimitsHashesUnparsableLastAppliedConfiguration(t *testing.T) {
	parsed := parseLogzioLimits(map[string]interface{}{
		"newObject": map[string]interface{}{"kind": "ConfigMap", "metadata": map[string]interface{}{
			"annotations": map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": `{"password":"plaintext`},
		}},
	})

	if masked := string(mustMarshal(t, parsed)); strings.Contains(masked, "plaintext") {
		t.Errorf("unparsable last-applied configuration must not be shipped as is: %s", masked)
	}
}

func TestParseLogzioLimitsKeepsLastAppliedConfigurationWithoutSensitiveEnvVars(t *testing.T) {
	annotation := `{"spec":{"strategy":{"rollingUpdate":{"maxUnavailable":"25%"}},"a":1,"<b>":2}}`
	parsed := parseLogzioLimits(map[string]interface{}{
		"newObject": map[string]interface{}{"kind": "Deployment", "metadata": map[string]interface{}{
			"annotations": map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": annotation},
		}},
	})

	got := parsed["newObject"].(map[string]interface{})["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["kubectl_kubernetes_io/last_applied_configuration"]
	if got != annotation {
		t.Errorf("annotation without sensitive env vars must be shipped byte for byte, got %v", got)
	}
}

func TestBuildSensitiveEnvName(t *testing.T) {
	custom, err := buildSensitiveEnvName(" internal_id|^ssn$|^x{1,3}$ ")
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{
		"DB_PASSWORD": true, // default pattern is kept
		"INTERNAL_ID": true,
		"ssn":         true,
		"MY_SSN":      false, // anchored custom pattern
		"XX":          true,  // commas inside a quantifier are not separators
		"LOG_LEVEL":   false,
	} {
		if got := custom.MatchString(name); got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}

	for _, invalid := range []string{"(unclosed", "x)"} {
		if _, err := buildSensitiveEnvName(invalid); err == nil {
			t.Errorf("expected an error for the invalid regex %q", invalid)
		}
	}
}

func TestConfigureMasking(t *testing.T) {
	defaultPattern, defaultKey := sensitiveEnvName, hashKey
	t.Cleanup(func() { sensitiveEnvName, hashKey = defaultPattern, defaultKey })
	t.Setenv("SENSITIVE_ENV_NAMES", "internal_id")
	t.Setenv("MASKING_HASH_KEY", "stable-key")

	ConfigureMasking()

	if !sensitiveEnvName.MatchString("INTERNAL_ID") || !sensitiveEnvName.MatchString("DB_PASSWORD") {
		t.Error("expected both the configured and the default pattern to match")
	}
	// HMAC-SHA256("stable-key", "hunter2"), so hashes stay comparable across restarts with the same key
	if got := hashData("hunter2"); got != "b236aaef4ae1938c91cfd114ef4d8bd6a9d9562df002c7c8aa0e8c74effebdd6" {
		t.Errorf("hashData with a configured key = %s", got)
	}
}

func TestHashDataIsKeyed(t *testing.T) {
	defaultKey := hashKey
	t.Cleanup(func() { hashKey = defaultKey })

	first := hashData("hunter2")
	if first != hashData("hunter2") {
		t.Error("hashing must be deterministic for the same key, so old and new values can be compared")
	}
	if len(first) != 64 {
		t.Errorf("expected a hex SHA-256 digest, got %q", first)
	}

	hashKey = []byte("another-key")
	if hashData("hunter2") == first {
		t.Error("a different key must give a different hash, otherwise the hash can be reversed with a dictionary")
	}
}

func TestParseLogzioLimitsHashesEachSensitiveFieldOnce(t *testing.T) {
	for i := 0; i < 50; i++ {
		parsed := parseLogzioLimits(map[string]interface{}{"tls.crt": "cert", "ca.crt": "ca"})

		if len(parsed) != 2 || parsed["tls_crt_hashed"] == nil || parsed["ca_crt_hashed"] == nil {
			t.Fatalf("expected each field to be hashed exactly once, got %v", parsed)
		}
	}
}

func TestParseLogzioLimitsMasksFieldsAlreadyNamedHashed(t *testing.T) {
	parsed := parseLogzioLimits(map[string]interface{}{"tls.crt_hashed": "user-value"})

	if strings.Contains(string(mustMarshal(t, parsed)), "user-value") {
		t.Errorf("a field whose name already ends with _hashed must still be masked: %v", parsed)
	}
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
