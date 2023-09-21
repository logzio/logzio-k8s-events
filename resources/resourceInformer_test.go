package resources

import (
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
	"log"
	"testing"
)

func createFakeResourceInformer(gvr schema.GroupVersionResource) cache.SharedIndexInformer {
	fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(fakeDynamicClient, 0, corev1.NamespaceAll, nil)
	fakeResourceInformer := factory.ForResource(gvr).Informer()
	if fakeResourceInformer == nil {
		log.Printf("[ERROR] Resource Informer was not created")
	} else {
		log.Printf("Resource Informer created successfully")
	}
	return fakeResourceInformer
}

func TestCreateResourceInformer(t *testing.T) {
	resourceGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	informer := createFakeResourceInformer(resourceGVR)

	if informer == nil {
		t.Errorf("Failed to create resource informer")
	}
}

// Define an interface that includes the function you want to mock
type InformerCreator interface {
	createFakeResourceInformer(gvr schema.GroupVersionResource, dynamicClient *fakeDynamic.FakeDynamicClient) cache.SharedIndexInformer
}

// Have your mock type implement the interface
type MockInformerCreator struct {
	mock.Mock
}

// Replace createResourceInformer with an instance of the interface

func TestAddEventHandlers(t *testing.T) {
	// Create a new mock informer creator
	fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
	mockInformerCreator := new(MockInformerCreator)
	mockInformer := createFakeResourceInformer(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "deployments"})
	// Define what should be returned when the mock is called
	mockInformerCreator.On("CreateFakeResourceInformer", mock.Anything, mock.Anything).Return(mockInformer)

	// Run the function that you're testing
	AddEventHandlers()

	// Check that the mock was called with the expected parameters
	mockInformerCreator.AssertCalled(t, "CreateFakeResourceInformer", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}, fakeDynamicClient)
	mockInformerCreator.AssertCalled(t, "CreateFakeResourceInformer", schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, fakeDynamicClient)
	// ... add more assertions here ...
}
func (m *MockInformerCreator) CreateFakeResourceInformer(gvr schema.GroupVersionResource, dynamicClient *fakeDynamic.FakeDynamicClient) cache.SharedIndexInformer {
	args := m.Called(gvr, dynamicClient)
	return args.Get(0).(cache.SharedIndexInformer)
}

//func TestAddInformerEventHandler(t *testing.T) {
//	resourceGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
//	mockDynamicClient := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
//
//	informer := CreateFakeResourceInformer(resourceGVR, mockDynamicClient)
//
//	if informer == nil {
//		t.Errorf("Failed to create resource informer")
//	}
//
//	synced := AddInformerEventHandler(informer)
//
//	if !synced {
//		t.Errorf("Failed to add event handler for informer")
//	}
//}
