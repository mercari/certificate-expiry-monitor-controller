package teams

import (
	"encoding/json"
	_ "errors"
	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	dummyEndpoint = "http://dummy_endpoint"
)

func TestNewNotifier(t *testing.T) {
	tests := []struct {
		arg struct {
			webhook string
		}
		success bool
	}{
		{
			arg: struct {
				webhook string
			}{
				webhook: "token",
			},
			success: true,
		},
		{
			arg: struct {
				webhook string
			}{
				webhook: "",
			},
			success: false,
		},
	}

	for _, test := range tests {
		_, err := NewNotifier(test.arg.webhook)

		if (err == nil) != test.success {
			if test.success {
				t.Fatalf("Unexpected failed to initialize notifier: %s", err.Error())
			} else {
				t.Fatalf("Unexpected succeeded to initialize notifier")
			}
		}
	}
}

func TestString(t *testing.T) {
	if String() != notifierName {
		t.Fatal("Unmatch return value of String() with notifierName")
	}
}

func TestAlert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockHttpRequest(t)))
	defer server.Close()

	type TestArg struct {
		backend    *MsTeams
		expiration time.Time
		ingress    *source.Ingress
		tls        *source.IngressTLS
		opt        notifier.Option
	}

	type TestCase struct {
		args    TestArg
		success bool
	}

	tests := []TestCase{
		{
			args: TestArg{
				makeTestTeams(t, server.URL),
				time.Now(),
				makeTestIngress(t),
				makeTestIngressTLS(t),
				notifier.Option{AlertLevel: notifier.AlertLevelWarning},
			},
			success: true,
		},
		{
			args: TestArg{
				makeTestTeams(t, dummyEndpoint),
				time.Now(),
				makeTestIngress(t),
				makeTestIngressTLS(t),
				notifier.Option{AlertLevel: notifier.AlertLevelCritical},
			},
			success: false,
		},
	}

	for _, test := range tests {
		err := test.args.backend.Alert(test.args.expiration, test.args.ingress, test.args.tls, test.args.opt)

		if (err == nil) && test.success == false {
			t.Fatal("Unexpected result: Alert should be fail")
		}

		if (err != nil) && test.success == true {
			t.Fatalf("Unexpected result: %s", err.Error())
		}
	}
}

func makeTestTeams(t *testing.T, webhook string) *MsTeams {
	t.Helper()
	return &MsTeams{
		WebhookEndpoint: webhook,
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

func mockHttpRequest(t *testing.T) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var card MessageCard
		_ = decoder.Decode(&card)

		if card.Type != messageCardType {
			t.Fatal("Unexpected result: Alert should be fail")
		}
	}
}
