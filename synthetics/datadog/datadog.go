package synthetics

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/zorkian/go-datadog-api"
	"go.uber.org/zap"
)

// TestManager synchronize synthetics tests in Datadog with existing Kubernetes Ingress Endpoints
type TestManager struct {
	Client              Client
	Logger              *zap.Logger
	AlertMessage        string
	CheckInterval       int
	DefaultTag          string
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

// ContainsColon return whether a slice contains a specific value before the first colon
func ContainsColon(slice []string, val string) bool {
	for _, n := range slice {
		if strings.ContainsAny(n, ":") {
			s := strings.Split(n, ":")
			n = s[0]
		}
		if val == n {
			return true
		}
	}
	return false
}

// createManagedSyntheticsTest configures and create a new synthetics test in Datadog
func (tm *TestManager) createManagedSyntheticsTest(endpoint string, port int) (*datadog.SyntheticsTest, error) {
	newOptions := &datadog.SyntheticsOptions{}
	newOptions.SetAcceptSelfSigned(false)
	newOptions.SetTickEvery(tm.CheckInterval)

	expiryAssertion := &datadog.SyntheticsAssertion{}
	expiryAssertion.SetType("certificate")
	expiryAssertion.SetOperator("isInMoreThan")
	expiryAssertion.Target = 12

	newRequest := &datadog.SyntheticsRequest{}
	newRequest.SetHost(endpoint)
	newRequest.SetPort(port)

	newConfig := &datadog.SyntheticsConfig{}
	newConfig.Assertions = []datadog.SyntheticsAssertion{*expiryAssertion}
	newConfig.SetRequest(*newRequest)

	tags := tm.Tags
	tags = append(tags, tm.DefaultTag)

	newTest := &datadog.SyntheticsTest{Locations: []string{"aws:ap-northeast-1"}, Tags: tags}
	newTest.SetName(endpoint)
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
func (tm *TestManager) CreateManagedSyntheticsTests(endpointList []string) error {
	if tm.DefaultTag == "" {
		err := fmt.Errorf("No default tag is set for synthetics tests, aborting creation process")
		return err
	}
	// Get all existing synthetic tests
	tests, err := tm.Client.GetSyntheticsTests()
	if err != nil {
		log.Printf("Couldn't get synthetics tests from Datadog: %s\n", err.Error())
		return err
	}
	for _, endpoint := range endpointList {
		var matched bool
		port := 443
		if strings.ContainsAny(endpoint, ":") {
			s := strings.Split(endpoint, ":")
			endpoint = s[0]
			i, err := strconv.Atoi(s[1])
			if err != nil {
				err := fmt.Errorf("The port number is not a valid numeral")
				return err
			}
			port = i
		}
		// Normalize endpoint names from SYNTHETIC_ADDITIONAL_ENDPOINTS as they might have a defined port
		for _, test := range tests {
			// Only deal with tests having auto-generated tag
			if Contains(test.Tags, tm.DefaultTag) {
				endpointName, exists := test.GetNameOk()
				// Use matched to determine whether a skip/create operation is required
				if exists {
					log.Printf("warn: endpointName (Test name) is %s, endpoint (Ingress name) is %s", endpointName, endpoint)
					if endpoint == endpointName {
						matched = true
						break
					}
				} else {
					log.Printf("warn: Test %s doesn't have any name, skipping", test.GetPublicId())
				}
			}
		}
		if matched {
			log.Printf("Test is already existing for %s and Ingress exists", endpoint)
		} else {
			log.Printf("Creating new test for Ingress endpoint %s", endpoint)
			_, err := tm.createManagedSyntheticsTest(endpoint, port)
			if err != nil {
				log.Printf("Couldn't create the synthetic test for Ingress endpoint %s: %s\n", endpoint, err.Error())
			}
		}
	}
	return nil
}

// DeleteManagedSyntheticsTests removes managed synthetics test not matching the endpointList provided
func (tm *TestManager) DeleteManagedSyntheticsTests(endpointList []string) error {
	if tm.DefaultTag == "" {
		err := fmt.Errorf("No default tag is set for synthetics tests, aborting deletion process")
		return err
	}
	// Get all existing synthetic tests
	tests, err := tm.Client.GetSyntheticsTests()
	if err != nil {
		log.Printf("Couldn't get synthetics tests from Datadog: %s\n", err.Error())
	}
	// Slice containing all tests publicIds to delete
	toDelete := []string{}
	for _, test := range tests {
		// Only deal with tests having auto-generated tag
		if Contains(test.Tags, tm.DefaultTag) {
			if endpointName, exists := test.GetNameOk(); exists {
				if !ContainsColon(endpointList, endpointName) {
					log.Printf("warn: Managed test %s, with hostname %s, doesn't have any matching ingress, adding to delete list", test.GetPublicId(), endpointName)
					toDelete = append(toDelete, test.GetPublicId())
				}
			} else {
				log.Printf("warn: Test %s doesn't have any name, skipping", test.GetPublicId())
			}
		}
	}
	// Delete only when there are candidates to deletion
	if len(toDelete) >= 1 {
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
