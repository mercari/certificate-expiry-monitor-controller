package slack

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

func TestNewPostParameters(t *testing.T) {
	expectedDays := 5

	type TestArg struct {
		expiration time.Time
		ingress    *source.Ingress
		tls        *source.IngressTLS
		alertLevel notifier.AlertLevel
	}

	type TestExpect struct {
		color      string
		days       int
		subPreText string
	}

	type TestCase struct {
		args     TestArg
		expected TestExpect
	}

	tests := []TestCase{
		{
			args: TestArg{
				// To consider effect of time lapse: `.Add(time.Hour * 12)`
				expiration: time.Now().AddDate(0, 0, expectedDays).Add(time.Hour * 12),
				ingress:    makeTestIngress(t),
				tls:        makeTestIngressTLS(t),
				alertLevel: notifier.AlertLevelWarning,
			},
			expected: TestExpect{
				color:      "warning",
				days:       5,
				subPreText: "[WARNING]",
			},
		},
		{
			args: TestArg{
				expiration: time.Now().AddDate(0, 0, -expectedDays),
				ingress:    makeTestIngress(t),
				tls:        makeTestIngressTLS(t),
				alertLevel: notifier.AlertLevelCritical,
			},
			expected: TestExpect{
				color:      "danger",
				days:       5,
				subPreText: "[CRITICAL]",
			},
		},
	}

	for _, test := range tests {
		actual := newPostParameters(test.args.expiration, test.args.ingress, test.args.tls, test.args.alertLevel)
		actualAttachment := actual.Attachments[0]

		if !strings.Contains(actualAttachment.Pretext, test.expected.subPreText) {
			t.Fatalf("Pretext not includes %s: %s", test.expected.subPreText, actualAttachment.Pretext)
		}

		if actualAttachment.Color != test.expected.color {
			t.Fatalf("Unexpected Alert color %s, expected %s", actualAttachment.Color, test.expected.color)
		}

		if !strings.Contains(actualAttachment.Pretext, strconv.Itoa(test.expected.days)) {
			t.Fatalf("Pretext not includes expected days %d: %s", test.expected.days, actualAttachment.Pretext)
		}
	}
}

func TestNewAttachmentFields(t *testing.T) {
	expectedFieldCount := 6

	tests := []struct {
		cluster    string
		namespace  string
		name       string
		secret     string
		expiration time.Time
		endpoints  []*source.TLSEndpoint
	}{
		{
			cluster:    "dummyCluster",
			namespace:  "dummyNamespace",
			name:       "dummyName",
			secret:     "dummySecret",
			expiration: time.Now(),
			endpoints:  []*source.TLSEndpoint{},
		},
	}

	for _, test := range tests {
		actual := newAttachmentFields(test.cluster, test.namespace, test.name, test.secret, test.expiration, test.endpoints)

		if len(actual) != expectedFieldCount {
			t.Fatalf("Unexpected number of fields: %d", len(actual))
		}
	}
}
