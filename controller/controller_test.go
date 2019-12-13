package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	synthetics "github.com/lainra/certificate-expiry-monitor-controller/synthetics/datadog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/notifier/log"
	"github.com/mercari/certificate-expiry-monitor-controller/source"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	dummyURL = "sample.mercari.dummy"
)

func TestNewController(t *testing.T) {
	type testArg struct {
		logger      *zap.Logger
		clientSet   kubernetes.Interface
		interval    time.Duration
		threshold   time.Duration
		notifiers   []notifier.Notifier
		testManager *synthetics.TestManager
	}
	tests := []struct {
		arg     testArg
		success bool
	}{
		{
			arg: testArg{
				logger:    zap.NewNop(),
				interval:  10 * time.Hour,
				threshold: 48 * time.Hour,
				notifiers: []notifier.Notifier{},
			},
			success: false,
		},
		{
			arg: testArg{
				clientSet: makeTestClientSet(t, []string{dummyURL}),
				interval:  10 * time.Hour,
				threshold: 48 * time.Hour,
				notifiers: []notifier.Notifier{},
			},
			success: false,
		},
		{
			arg: testArg{
				logger:    zap.NewNop(),
				clientSet: makeTestClientSet(t, []string{dummyURL}),
				interval:  10 * time.Hour,
				threshold: 48 * time.Hour,
				notifiers: []notifier.Notifier{},
			},
			success: true,
		},
	}

	for _, test := range tests {
		_, err := NewController(test.arg.logger, test.arg.clientSet, test.arg.interval, test.arg.threshold, test.arg.notifiers, test.arg.testManager)

		if (err == nil) != test.success {
			t.Fatalf("Unexpected result with error: %s", err.Error())
		}
	}
}

func TestRun(t *testing.T) {
	server := httptest.NewTLSServer(http.NewServeMux())
	defer server.Close()
	u, _ := url.Parse(server.URL)

	// Overwrite default port number to test server.URL
	source.DefaultPortNumber = u.Port()

	// Observe all logs that printed by notifiers
	core, recorded := observer.New(zapcore.InfoLevel)

	stopCh := make(chan struct{}, 1)
	interval := 1 * time.Second
	threshold := 24 * time.Hour
	notifiers := []notifier.Notifier{log.NewNotifier(zap.NewNop())}
	clientSet := makeTestClientSet(t, []string{u.Hostname()})
	testManager, err := synthetics.NewTestManager("api_key", "app_key")
	testManager.Client = nil

	controller, err := NewController(zap.New(core), clientSet, interval, threshold, notifiers, testManager)
	if err != nil {
		t.Fatalf("Unexpected falied to initialize controller: %s", err.Error())
	}

	expectedCount := 2
	expectedMessage := "onInteration"

	onIteration = func() {
		controller.Logger.Info(expectedMessage)
	}

	defer func() {
		onIteration = func() {}
	}()

	go func() {
		for i := 0; i < expectedCount; i++ {
			<-time.After(interval)
		}
		close(stopCh)
	}()

	controller.Run(stopCh)

	fields := recorded.FilterMessage(expectedMessage)
	if fields.Len() != expectedCount {
		t.Fatalf("Unexpected count that triggered WARNING alert: %d", fields.Len())
	}
}

func TestRunOnce(t *testing.T) {
	server := httptest.NewTLSServer(http.NewServeMux())
	defer server.Close()
	u, _ := url.Parse(server.URL)

	// Prefetch certs to get expiration for testing
	expectedCerts, _ := source.NewTLSEndpoint(u.Hostname(), u.Port()).GetCertificates()
	expiration := expectedCerts[0].NotAfter

	// Overwrite default port number to test server.URL
	source.DefaultPortNumber = u.Port()

	interval := 10 * time.Hour
	threshold := 48 * time.Hour
	clientSet := makeTestClientSet(t, []string{u.Hostname()})
	testManager, _ := synthetics.NewTestManager("api_key", "app_key")
	testManager.Client = nil

	tests := []struct {
		arg                time.Time
		expectedMatchField zap.Field
		expectedMatchCount int
	}{
		{
			// expiration has not reached the threshold.
			arg:                expiration.Add(-(threshold + threshold)),
			expectedMatchField: zap.String("Level", "WARNING"),
			expectedMatchCount: 0,
		},
		{
			// expiration has not reached the threshold.
			arg:                expiration.Add(-(threshold + threshold)),
			expectedMatchField: zap.String("Level", "CRITICAL"),
			expectedMatchCount: 0,
		},
		{
			// expiration reached the threshold.
			arg:                expiration,
			expectedMatchField: zap.String("Level", "WARNING"),
			expectedMatchCount: 1,
		},
		{
			// already expired
			arg:                expiration.Add(threshold),
			expectedMatchField: zap.String("Level", "CRITICAL"),
			expectedMatchCount: 1,
		},
	}

	for _, test := range tests {
		core, recorded := observer.New(zapcore.InfoLevel)

		notifiers := []notifier.Notifier{log.NewNotifier(zap.New(core))}

		controller, err := NewController(zap.NewNop(), clientSet, interval, threshold, notifiers, testManager)
		if err != nil {
			t.Fatalf("Unexpected falied to initialize controller: %s", err.Error())
		}

		err = controller.runOnce(test.arg)
		if err != nil {
			t.Fatalf("Unexpected falied to run runOnce: %s", err.Error())
		}

		fields := recorded.FilterField(test.expectedMatchField)
		if fields.Len() != test.expectedMatchCount {
			t.Fatalf("Not found expected value: { %s: %s }", test.expectedMatchField.Key, test.expectedMatchField.String)
		}
	}
}

func makeTestClientSet(t *testing.T, availableHosts []string) kubernetes.Interface {
	t.Helper()

	return fake.NewSimpleClientset(
		&v1beta1.IngressList{
			Items: []v1beta1.Ingress{
				// case: expected
				v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "ingress1",
						Namespace:   "namespace1",
						ClusterName: "clusterName",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{
							{
								Hosts:      availableHosts,
								SecretName: "ingressSecret1",
							},
						},
					},
				},
				// case: empty TLS
				v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "ingress2",
						Namespace:   "namespace2",
						ClusterName: "clusterName",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{},
					},
				},
				// case: unreachable host
				v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "ingress3",
						Namespace:   "namespace3",
						ClusterName: "clusterName",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{
							{
								Hosts:      []string{dummyURL},
								SecretName: "ingressSecret3",
							},
						},
					},
				},
				// case: empty hosts (but TLS enabled)
				v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "ingress4",
						Namespace:   "namespace4",
						ClusterName: "clusterName",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{
							{
								Hosts:      []string{},
								SecretName: "ingressSecret4",
							},
						},
					},
				},
			},
		},
	)
}
