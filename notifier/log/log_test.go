package log

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

func TestAlert(t *testing.T) {
	type TestArg struct {
		expiration time.Time
		ingress    *source.Ingress
		tls        *source.IngressTLS
		opt        notifier.Option
	}

	type TestCase struct {
		args               TestArg
		expectedMatchField zap.Field
		expectedMatchCount int
	}

	tests := []TestCase{
		{
			args: TestArg{
				expiration: time.Now(),
				ingress:    makeTestIngress(t),
				tls:        makeTestIngressTLS(t),
				opt:        notifier.Option{AlertLevel: notifier.AlertLevelWarning},
			},
			expectedMatchField: zap.String("Level", "WARNING"),
			expectedMatchCount: 1,
		},
		{
			args: TestArg{
				expiration: time.Now(),
				ingress:    makeTestIngress(t),
				tls:        makeTestIngressTLS(t),
				opt:        notifier.Option{AlertLevel: notifier.AlertLevelCritical},
			},
			expectedMatchField: zap.String("Level", "CRITICAL"),
			expectedMatchCount: 1,
		},
	}

	for _, test := range tests {
		core, recorded := observer.New(zapcore.InfoLevel)
		l := NewNotifier(zap.New(core))
		l.Alert(test.args.expiration, test.args.ingress, test.args.tls, test.args.opt)

		fields := recorded.FilterField(test.expectedMatchField)
		if fields.Len() != test.expectedMatchCount {
			t.Fatalf("Not found expected value: { %s: %s }", test.expectedMatchField.Key, test.expectedMatchField.String)
		}
	}
}

func makeTestIngress(t *testing.T) *source.Ingress {
	t.Helper()
	return &source.Ingress{
		ClusterName: "DummyClusterName",
		Namespace:   "DummyNamespace",
		Name:        "DummyName",
		TLS:         []*source.IngressTLS{},
	}
}

func makeTestIngressTLS(t *testing.T) *source.IngressTLS {
	t.Helper()
	return &source.IngressTLS{
		Endpoints: []*source.TLSEndpoint{
			source.NewTLSEndpoint("host01.example.com", ""),
			source.NewTLSEndpoint("host02.example.com", ""),
		},
		SecretName: "DummySecretName",
	}
}
