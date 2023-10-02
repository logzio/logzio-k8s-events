package common

import (
	"k8s.io/apimachinery/pkg/runtime"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

// CreateFakeClient creates a fake client.
func CreateFakeClient() (mockClient *fake.Clientset) {

	mockClient = fake.NewSimpleClientset()

	return mockClient
}

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

// TestFakeClusterClient demonstrates how to use a fake client with SharedInformerFactory in tests.
func TestFakeClusterClient(t *testing.T) {
	// Create the fake client.
	fakeK8sClient := CreateFakeClient()

	if fakeK8sClient == nil {
		t.Error("Failed to create fake client")
	} else {
		t.Log("Created fake client")
	}
}
