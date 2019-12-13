package synthetics

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/zorkian/go-datadog-api"
)

type fakeClient struct {
	t *testing.T

	validateCreateSyntheticsTestFunc  func(t *testing.T, syntheticsTest *datadog.SyntheticsTest) (*datadog.SyntheticsTest, error)
	validateGetSyntheticsTestsFunc    func(t *testing.T) []datadog.SyntheticsTest
	validateDeleteSyntheticsTestsFunc func(t *testing.T, publicIds []string) error
}

func (f *fakeClient) CreateSyntheticsTest(syntheticsTest *datadog.SyntheticsTest) (*datadog.SyntheticsTest, error) {
	test, _ := f.validateCreateSyntheticsTestFunc(f.t, syntheticsTest)
	return test, nil
}

func (f *fakeClient) GetSyntheticsTests() ([]datadog.SyntheticsTest, error) {
	tests := f.validateGetSyntheticsTestsFunc(f.t)
	return tests, nil
}

func (f *fakeClient) DeleteSyntheticsTests(publicIds []string) error {
	error := f.validateDeleteSyntheticsTestsFunc(f.t, publicIds)
	return error
}

func TestCreateSyntheticsTest(t *testing.T) {
	client := &fakeClient{
		t: t,
		validateCreateSyntheticsTestFunc: func(t *testing.T, syntheticsTest *datadog.SyntheticsTest) (*datadog.SyntheticsTest, error) {
			if len(syntheticsTest.GetConfig().Assertions) != 2 {
				t.Fatalf("got %v, want %v", len(syntheticsTest.GetConfig().Assertions), 2)
			}
			if got, want := *syntheticsTest.GetOptions().AcceptSelfSigned, false; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := *syntheticsTest.GetOptions().TickEvery, 60; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Assertions[0].GetType(), "certificate"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Assertions[0].GetOperator(), "isInMoreThan"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Assertions[0].Target, 12; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Assertions[1].GetType(), "property"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Assertions[1].GetOperator(), "is"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Assertions[1].GetProperty(), "issuer.CN"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Assertions[1].Target, "Let's Encrypt Authority X3"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Request.GetPort(), 443; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetConfig().Request.GetHost(), "example.com"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetType(), "api"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := syntheticsTest.GetSubtype(), "ssl"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if len(syntheticsTest.Locations) != 1 {
				t.Fatalf("got %v, want %v", len(syntheticsTest.Locations), 1)
			}
			if got, want := syntheticsTest.Locations[0], "aws:ap-northeast-1"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if len(syntheticsTest.Tags) != 1 {
				t.Fatalf("got %v, want %v", len(syntheticsTest.Tags), 1)
			}
			if got, want := syntheticsTest.Tags[0], "managed-by-cert-exp-mon"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			return syntheticsTest, nil
		},
	}

	apiKey := "api_key"
	appKey := "app_key"
	tm, err := NewTestManager(apiKey, appKey)
	if err != nil {
		t.Fatalf("want nil, got %s", err)
	}
	tm.CheckInterval = 60
	tm.AlertMessage = ""
	tm.Client = client
	tm.DefaultTag = "managed-by-cert-exp-mon"
	endpoint := "example.com"
	port := 443
	tm.createManagedSyntheticsTest(endpoint, port)
}

func TestGetSyntheticsTests(t *testing.T) {
	client := &fakeClient{
		t: t,
		validateGetSyntheticsTestsFunc: func(t *testing.T) []datadog.SyntheticsTest {
			test := new(datadog.SyntheticsTest)
			test2 := new(datadog.SyntheticsTest)
			tests := &[]datadog.SyntheticsTest{*test, *test2}
			return *tests
		},
	}

	apiKey := "api_key"
	appKey := "app_key"
	tm, err := NewTestManager(apiKey, appKey)
	if err != nil {
		t.Fatalf("want nil, got %s", err)
	}
	tm.Client = client
	if got, _ := tm.Client.GetSyntheticsTests(); got == nil {
		t.Fatal("want []datadog.SyntheticsTest, got nil")
	}

}

func TestNewTestManager(t *testing.T) {
	apiKey := "api_key"
	appKey := "app_key"
	if got, _ := NewTestManager(apiKey, appKey); got == nil {
		t.Fatal("want *NewTestManager, got nil")
	}
}

func TestCreateManagedSyntheticsTests(t *testing.T) {
	client := &fakeClient{
		t: t,
		validateGetSyntheticsTestsFunc: func(t *testing.T) []datadog.SyntheticsTest {
			test := new(datadog.SyntheticsTest)
			test.SetName("example.com")
			test.Tags = append(test.Tags, "managed-by-cert-expiry-mon")
			test2 := new(datadog.SyntheticsTest)
			test2.SetName("example2.com")
			tests := &[]datadog.SyntheticsTest{*test, *test2}
			return *tests
		},
		validateCreateSyntheticsTestFunc: func(t *testing.T, syntheticsTest *datadog.SyntheticsTest) (*datadog.SyntheticsTest, error) {
			return syntheticsTest, nil
		},
	}

	apiKey := "api_key"
	appKey := "app_key"
	tm, err := NewTestManager(apiKey, appKey)
	if err != nil {
		t.Fatalf("want nil, got %s", err)
	}
	tm.Client = client
	tm.DefaultTag = "managed-by-cert-expiry-mon"

	if got, _ := tm.Client.GetSyntheticsTests(); got == nil {
		t.Fatal("want []datadog.SyntheticsTest, got nil")
	}

	if got := captureOutput(func() {
		endpointList := []string{"example.com"}
		tm.CreateManagedSyntheticsTests(endpointList)
	}); strings.Contains(got, "Test is already existing for") == false {
		t.Fatalf("want `Test is already existing for example.com and Ingress exists`, got %s", got)
	}
	if got := captureOutput(func() {
		endpointList := []string{"nonexistinguri.com"}
		tm.CreateManagedSyntheticsTests(endpointList)
	}); strings.Contains(got, "Creating new test for Ingress endpoint nonexistinguri.com") == false {
		t.Fatalf("want test, got %s", got)
	}

}

func TestDeleteManagedSyntheticsTests(t *testing.T) {
	client := &fakeClient{
		t: t,
		validateGetSyntheticsTestsFunc: func(t *testing.T) []datadog.SyntheticsTest {
			test := new(datadog.SyntheticsTest)
			test.SetName("example.com")
			test.SetPublicId("aaa-aaa-aaa")
			test.Tags = append(test.Tags, "managed-by-cert-expiry-mon")

			test2 := new(datadog.SyntheticsTest)
			test.SetPublicId("bbb-bbb-bbb")
			test2.SetName("example2.com")

			test3 := new(datadog.SyntheticsTest)
			test3.SetPublicId("ccc-ccc-ccc")
			test3.SetName("example3.com")
			test3.Tags = append(test.Tags, "managed-by-cert-expiry-mon")

			tests := &[]datadog.SyntheticsTest{*test, *test2, *test3}
			return *tests
		},
		validateDeleteSyntheticsTestsFunc: func(t *testing.T, publicIds []string) error {
			return nil
		},
	}

	apiKey := "api_key"
	appKey := "app_key"
	tm, _ := NewTestManager(apiKey, appKey)
	tm.Client = client
	tm.DefaultTag = "managed-by-cert-expiry-mon"

	// Case 1: Only example3.com should be deleted, example.com is in the Ingress endpoint list and example2 doesn't have the managed tag
	if got := captureOutput(func() {
		endpointList := []string{"example.com"}
		tm.DeleteManagedSyntheticsTests(endpointList)
	}); strings.Contains(got, "Managed test ccc-ccc-ccc, with hostname example3.com, doesn't have any matching ingress, adding to delete list") == false {
		t.Fatalf("want `Managed test ccc-ccc-ccc, with hostname example3.com, doesn't have any matching ingress, adding to delete list`, got %s", got)
	}
	// Case 2: example.com and example3.com should be deleted, example2.com doesn't have the tag
	if got := captureOutput(func() {
		endpointList := []string{}
		tm.DeleteManagedSyntheticsTests(endpointList)
	}); strings.Contains(got, "Deleting 2 managed tests") == false {
		t.Fatalf("want `Deleting 2 managed tests`, got %s", got)
	}
	// Case 3: Nothing should be deleted, expect no output
	if got := captureOutput(func() {
		endpointList := []string{"example.com", "example3.com"}
		tm.DeleteManagedSyntheticsTests(endpointList)
	}); strings.Contains(got, "No test candidate for deletion") == false {
		t.Fatalf("want `No test candidate for deletion`, got %s", got)
	}
	// Case 4: example2.com should not be deleted as it doesn't have the managed tag, expect no output
	if got := captureOutput(func() {
		endpointList := []string{"example2.com"}
		tm.DeleteManagedSyntheticsTests(endpointList)
	}); strings.Contains(got, "re") == false {
		t.Fatalf("want `Deleting 2 managed tests`, got %s", got)
	}
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}
