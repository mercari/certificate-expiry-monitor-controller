package config

import (
	"testing"
	"time"
)

func TestDefaultEnv(t *testing.T) {
	var env Env
	err := env.ParseEnv()

	if err != nil {
		t.Fatal("Failed to parse env that prepared by testcase")
	}
	if env.VerifyInterval != 12*time.Hour {
		t.Fatal("Unexpected default value in INTERVAL")
	}
	if env.AlertThreshold != 336*time.Hour {
		t.Fatal("Unexpected default value in THRESHOLD")
	}
	if env.LogLevel != "INFO" {
		t.Fatal("Unexpected default value in LOG_LEVEL")
	}
	if env.KubeconfigPath != "" {
		t.Fatal("Unexpected default value in KUBE_CONFIG_PATH")
	}
	if len(env.Notifiers) != 1 || env.Notifiers[0] != "log" {
		t.Fatal("Unexpected default value in NOTIFIERS")
	}
	if env.DatadogAPIKey != "" {
		t.Fatal("Unexpected default value in DATADOG_API_KEY")
	}
	if env.DatadogAppKey != "" {
		t.Fatal("Unexpected default value in DATADOG_APPLICATION_KEY")
	}
	if env.TestManager {
		t.Fatal("Unexpected default value in SYNTHETICS_ENABLED")
	}
	if env.CheckInterval != 900 {
		t.Fatal("Unexpected default value in SYNTHETICS_CHECK_INTERVAL")
	}
	if env.AlertMessage != "" {
		t.Fatal("Unexpected default value in SYNTHETICS_ALERT_MESSAGE")
	}
	if env.Tags != nil {
		t.Fatal("Unexpected default value in SYNTHETICS_TAGS")
	}
	if env.AdditionalEndpoints != nil {
		t.Fatal("Unexpected default value in SYNTHETICS_ADDITIONAL_ENDPOINTS")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		env      *Env
		expected bool
	}{
		struct {
			env      *Env
			expected bool
		}{
			env:      &Env{VerifyInterval: time.Minute * 1, AlertThreshold: time.Hour * 24},
			expected: true,
		},
		struct {
			env      *Env
			expected bool
		}{
			env:      &Env{VerifyInterval: time.Second * 59, AlertThreshold: time.Hour * 24},
			expected: false,
		},
		struct {
			env      *Env
			expected bool
		}{
			env:      &Env{VerifyInterval: time.Hour * 25, AlertThreshold: time.Hour * 24},
			expected: false,
		},
		struct {
			env      *Env
			expected bool
		}{
			env:      &Env{VerifyInterval: time.Hour * 24, AlertThreshold: time.Hour * 23},
			expected: false,
		},
	}

	for _, test := range tests {
		result := test.env.validate()
		if (result == nil) != test.expected {
			if test.expected {
				t.Fatalf("Unexpected result error: validation should be success (err: %s)", result)
			} else {
				t.Fatal("Unexpected result error: validation should be failed")
			}
		}
	}
}
