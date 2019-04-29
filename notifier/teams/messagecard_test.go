package teams

import (
	_ "errors"
	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	expectedDays = 5
)

func TestNewMessageCard(t *testing.T) {
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
				expiration: time.Now().AddDate(0, 0, expectedDays).Add(time.Hour * 12),
				ingress:    makeTestIngress(t),
				tls:        makeTestIngressTLS(t),
				alertLevel: notifier.AlertLevelWarning,
			},
			expected: TestExpect{
				color:      warningColor,
				days:       expectedDays,
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
				color:      dangerColor,
				days:       expectedDays,
				subPreText: "[CRITICAL]",
			},
		},
	}

	for _, test := range tests {
		actual := newMessageCard(test.args.expiration, test.args.ingress, test.args.tls, test.args.alertLevel)

		if !strings.Contains(actual.Title, test.expected.subPreText) {
			t.Fatalf("Title doesn't include %s: %s", test.expected.subPreText, actual.Title)
		}

		if !strings.Contains(actual.Summary, test.expected.subPreText) {
			t.Fatalf("Summary doesn't include %s: %s", test.expected.subPreText, actual.Summary)
		}

		if !strings.Contains(actual.Title, strconv.Itoa(test.expected.days)) {
			t.Fatalf("Title doesn't include expected days %d: %s", test.expected.days, actual.Title)
		}

		if !strings.Contains(actual.Summary, strconv.Itoa(test.expected.days)) {
			t.Fatalf("Summary doesn't include expected days %d: %s", test.expected.days, actual.Summary)
		}

		if actual.ThemeColor != test.expected.color {
			t.Fatalf("Unexpected Alert color %s, expected %s", actual.ThemeColor, test.expected.color)
		}
	}
}

func TestNewAttachmentFields(t *testing.T) {
	expectedFieldCount := 6

	tests := []struct {
		expiration time.Time
		ingress    *source.Ingress
		tls        *source.IngressTLS
	}{
		{
			ingress:    makeTestIngress(t),
			tls:        makeTestIngressTLS(t),
			expiration: time.Now().AddDate(0, 0, expectedDays).Add(time.Hour * 12),
		},
	}

	for _, test := range tests {
		actual := newFactAttachments(test.ingress, test.tls, test.expiration)

		if len(actual) != expectedFieldCount {
			t.Fatalf("Unexpected number of fields: %d", len(actual))
		}
	}
}
