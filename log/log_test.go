package log

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	if _, err := NewLogger("INFO"); err != nil {
		t.Fatalf("Unexpected fail: %s", err.Error())
	}

	if _, err := NewLogger("DUMMY"); err == nil {
		t.Fatal("Unexpected success")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		arg           string
		expectedValue zapcore.Level
		success       bool
	}{
		{
			arg:           "info",
			expectedValue: zapcore.InfoLevel,
			success:       true,
		},
		{
			arg:           "info",
			expectedValue: zapcore.InfoLevel,
			success:       true,
		},
		{
			arg:     "DUMMY",
			success: false,
		},
	}

	for _, test := range tests {
		actual, err := parseLogLevel(test.arg)

		if (err == nil) != test.success {
			if test.success {
				t.Fatalf("Unexpected fail by error: %s", err.Error())
			} else {
				t.Fatalf("Unexpected success")
			}
		}

		if actual != test.expectedValue {
			t.Fatalf("Unexpected result value: %d", actual)
		}
	}
}
