package common

import (
	"encoding/json"
	"log"
	"reflect"
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
	err = json.Unmarshal(eventLogData, &eventLog)
	if err != nil {
		log.Printf("Error unmarshalling event log: %v", err)
		return nil
	}
	return eventLog
}
func TestParseLogzioLimits(t *testing.T) {
	eventLog := GetTestEventLog()
	parsedLogEvent := eventLog

	t.Run("testNewObjectKind", func(t *testing.T) {
		if eventLog["newObject"] != nil {
			eventI := eventLog["newObject"].(map[string]interface{})
			if eventI["kind"] != nil {
				eventKind = eventI["kind"].(string)
			}
		}
	})

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
				parseLogzioLimits(nestedField)
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
