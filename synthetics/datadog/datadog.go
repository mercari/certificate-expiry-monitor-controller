package synthetics

import (
	"errors"
	"fmt"
	"github.com/zorkian/go-datadog-api"
	"go.uber.org/zap"
	"log"
)

// TestManager synchronize synthetics tests in Datadog with existing Kubernetes Ingress Endpoints
type TestManager struct {
	Client              Client
	Logger              *zap.Logger
	AlertMessage        string
	CheckInterval       int
	DefaultTag          string
	DefaultLocations    []string
	Tags                []string
	AdditionalEndpoints []string
	Enabled             bool
}

// Client is an interface that clients implement to manage synthetic tests in Datadog.
type Client interface {
	CreateSyntheticsTest(syntheticsTest *datadog.SyntheticsTest) (*datadog.SyntheticsTest, error)
	GetSyntheticsTests() ([]datadog.SyntheticsTest, error)
	DeleteSyntheticsTests(publicIds []string) error
}

type SyntheticEndpoint struct {
	Hostname string
	Port int
}

type SyntheticEndpoints map[string]SyntheticEndpoint

// NewTestManager creates a new datadog.Client
func NewTestManager(apiKey string, appKey string) (*TestManager, error) {
	if apiKey == "" {
		return &TestManager{}, errors.New("datadog api key is required")
	}
	if appKey == "" {
		return &TestManager{}, errors.New("datadog application key is required")
	}
	return &TestManager{
		Client: datadog.NewClient(apiKey, appKey),
	}, nil
}

// Contains return whether a slice contains a specific value
func Contains(slice []string, val string) bool {
	for _, n := range slice {
		if val == n {
			return true
		}
	}
	return false
}

// getManagedSyntheticsTests returns all synthetics tests matching the default tag
func (tm *TestManager) getManagedSyntheticsTests() ([]datadog.SyntheticsTest, error) {
	// Return error if default tag is not set
	if tm.DefaultTag == "" {
		err := fmt.Errorf("No default tag is set for synthetics tests, aborting creation process")
		return nil, err
	}
	var managedTests []datadog.SyntheticsTest
	// Get all existing synthetic tests
	tests, err := tm.Client.GetSyntheticsTests()
	if err != nil {
		log.Printf("Failed to get synthetics tests from Datadog: %s\n", err.Error())
		return nil, err
	}
	for _, test := range tests {
		// Only deal with tests having auto-generated tag
		if Contains(test.Tags, tm.DefaultTag) {
			if _, exists := test.GetNameOk(); exists {
				managedTests = append(managedTests, test)
			}
		}
	}

	return managedTests, nil
}

// createManagedSyntheticsTest configures and create a new synthetics test in Datadog
func (tm *TestManager) createManagedSyntheticsTest(name string, endpoint SyntheticEndpoint) (*datadog.SyntheticsTest, error) {
	newOptions := &datadog.SyntheticsOptions{}
	newOptions.SetAcceptSelfSigned(false)
	newOptions.SetTickEvery(tm.CheckInterval)

	expiryAssertion := &datadog.SyntheticsAssertion{}
	expiryAssertion.SetType("certificate")
	expiryAssertion.SetOperator("isInMoreThan")
	expiryAssertion.Target = 12

	newRequest := &datadog.SyntheticsRequest{}
	newRequest.SetHost(endpoint.Hostname)
	newRequest.SetPort(endpoint.Port)

	newConfig := &datadog.SyntheticsConfig{}
	newConfig.Assertions = []datadog.SyntheticsAssertion{*expiryAssertion}
	newConfig.SetRequest(*newRequest)

	tags := tm.Tags
	tags = append(tags, tm.DefaultTag)

	newTest := &datadog.SyntheticsTest{Locations: tm.DefaultLocations, Tags: tags}
	newTest.SetName(name)
	newTest.SetType("api")
	newTest.SetSubtype("ssl")
	newTest.SetConfig(*newConfig)

	newTest.SetMessage(tm.AlertMessage)
	newTest.SetOptions(*newOptions)

	test, err := tm.Client.CreateSyntheticsTest(newTest)
	if err != nil {
		return nil, err
	}
	return test, nil
}

// CreateManagedSyntheticsTests creates synthetics test according to the endpointList provided
func (tm *TestManager) CreateManagedSyntheticsTests(endpoints SyntheticEndpoints) error {
	// Get all existing synthetic tests
	tests, err := tm.getManagedSyntheticsTests()
	if err != nil {
		log.Printf("Failed to get synthetics tests: %s\n", err.Error())
		return err
	}
	for name, endpoint := range endpoints {
		var matched bool

		// Normalize endpoint names from SYNTHETIC_ADDITIONAL_ENDPOINTS as they might have a defined port
		for _, test := range tests {
			if name == test.GetName() {
				matched = true
			}
		}
		if matched {
			log.Printf("Test is already existing for %s:%d and Ingress exists", endpoint.Hostname, endpoint.Port)
		} else {
			log.Printf("Creating new test for Ingress endpoint %s:%d", endpoint.Hostname, endpoint.Port)
			_, err := tm.createManagedSyntheticsTest(name, endpoint)
			if err != nil {
				log.Printf("Couldn't create the synthetic test for Ingress endpoint %s:%d: %s\n", endpoint.Hostname, endpoint.Port, err.Error())
			}
		}
	}
	return nil
}

// DeleteManagedSyntheticsTests removes managed synthetics test not matching the endpointList provided
func (tm *TestManager) DeleteManagedSyntheticsTests(endpoints SyntheticEndpoints) error {
	// Get all existing synthetic tests
	tests, err := tm.getManagedSyntheticsTests()
	if err != nil {
		log.Printf("Failed to get synthetics tests: %s\n", err.Error())
		return err
	}
	// Slice containing all tests publicIds to delete
	var toDelete []string
	for _, test := range tests {
		if _, ok := endpoints[test.GetName()]; !ok {
			log.Printf("warn: Managed test %s, with hostname %s, doesn't have any matching ingress, adding to delete list", test.GetPublicId(), test.GetName())
			toDelete = append(toDelete, test.GetPublicId())
		}
	}
	// Delete only when there are candidates to deletion
	if toDelete != nil && len(toDelete) >= 1 {
		log.Printf("Deleting %d managed tests", len(toDelete))
		err := tm.Client.DeleteSyntheticsTests(toDelete)
		if err != nil {
			log.Printf("Failed to delete managed tests: %s\n", err.Error())
			return err
		}
	} else {
		log.Printf("No test candidate for deletion")
	}
	return nil
}
