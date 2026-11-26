package common

import (
	"k8s.io/apimachinery/pkg/runtime"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"testing"
)

// CreateDynamicFakeClient creates a fake dynamic client.
func CreateDynamicFakeClient() (mockDynamicClient *fakeDynamic.FakeDynamicClient) {
	scheme := runtime.NewScheme()

	mockDynamicClient = fakeDynamic.NewSimpleDynamicClient(scheme)

	return mockDynamicClient
}

// TestFakeDynamicClient demonstrates how to use a fake dynamic client with SharedInformerFactory in tests.
func TestFakeDynamicClient(t *testing.T) {
	// Create the fake client.
	fakeDynamicClient := CreateDynamicFakeClient()
	if fakeDynamicClient == nil {
		t.Error("Failed to create fake dynamic client")
	} else {
		t.Log("Created fake dynamic client")
	}

}
