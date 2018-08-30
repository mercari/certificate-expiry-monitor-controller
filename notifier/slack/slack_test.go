package slack

import (
	"errors"
	"testing"
	"time"

	libSlack "github.com/nlopes/slack"
	"go.uber.org/ratelimit"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

const (
	stubClientChannelName = "random"
	dummyToken            = "dummy_token"
)

type fakeClient struct {
	Token string
}

func (client *fakeClient) PostMessage(channel string, message string, params libSlack.PostMessageParameters) (string, string, error) {
	if channel != stubClientChannelName {
		return "", "", errors.New("Unexpected channel name")
	}
	return "", "", nil
}

func TestNewNotifier(t *testing.T) {
	tests := []struct {
		arg struct {
			token   string
			channel string
		}
		success bool
	}{
		{
			arg: struct {
				token   string
				channel string
			}{
				token:   "token",
				channel: "channel",
			},
			success: true,
		},
		{
			arg: struct {
				token   string
				channel string
			}{
				token:   "",
				channel: "channel",
			},
			success: false,
		},
		{
			arg: struct {
				token   string
				channel string
			}{
				token:   "token",
				channel: "",
			},
			success: false,
		},
	}

	for _, test := range tests {
		_, err := NewNotifier(test.arg.token, test.arg.channel)

		if (err == nil) != test.success {
			if test.success {
				t.Fatalf("Unexpected failed to initialize notifier: %s", err.Error())
			} else {
				t.Fatalf("Unexpected successed to initialize notifier")
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
	type TestArg struct {
		backend    *Slack
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
				makeTestSlack(t, dummyToken, stubClientChannelName),
				time.Now(),
				makeTestIngress(t),
				makeTestIngressTLS(t),
				notifier.Option{AlertLevel: notifier.AlertLevelWarning},
			},
			success: true,
		},
		{
			args: TestArg{
				makeTestSlack(t, dummyToken, "Dummy"+stubClientChannelName),
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

func TestPostWithRateLimiter(t *testing.T) {
	s := makeTestSlack(t, dummyToken, stubClientChannelName)

	err := s.postWithRateLimiter(stubClientChannelName, "", libSlack.PostMessageParameters{})
	if err != nil {
		t.Fatal("Raise error when testing postWithRateLimiter")
	}

	before := time.Now()

	err = s.postWithRateLimiter(stubClientChannelName, "", libSlack.PostMessageParameters{})
	if err != nil {
		t.Fatal("Raise error when testing postWithRateLimiter")
	}

	after := time.Now()

	if after.Sub(before).Seconds() <= float64(sendPerSecond) {
		t.Fatalf("Rate limit is not being observed in %d per second. Actual: %f", sendPerSecond, after.Sub(before).Seconds())
	}
}

func makeTestSlack(t *testing.T, token string, channel string) *Slack {
	t.Helper()
	return &Slack{
		APIClient:   &fakeClient{Token: token}, // Replace client to fake client
		ChannelName: channel,
		RateLimiter: ratelimit.New(sendPerSecond),
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
