package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const (
	// The following values used by validate function.
	lowerIntervalMinutes = 1  // INTERVAL must be more than 1 minute
	upperIntervalHours   = 24 // INTERVAL must be less than 24 hours
	lowerThresholdHours  = 24 // THRESHOLD must be more than 24 hours
)

// Env struct defines configuration of controller that provided by ENV.
type Env struct {
	// Original configurations
	LogLevel       string        `envconfig:"LOG_LEVEL" default:"INFO"`
	KubeconfigPath string        `envconfig:"KUBE_CONFIG_PATH"`
	VerifyInterval time.Duration `envconfig:"INTERVAL" default:"12h"`
	AlertThreshold time.Duration `envconfig:"THRESHOLD" default:"336h"`
	Notifiers      []string      `envconfig:"NOTIFIERS" default:"log"`

	// Configration for Slack
	SlackToken   string `envconfig:"SLACK_TOKEN"`
	SlackChannel string `envconfig:"SLACK_CHANNEL"`

	// Configration for Teams
	TeamsWebhook string `envconfig:"TEAMS_WEBHOOK"`
}

// ParseEnv function sets to Env struct and verify it.
// If varify failed, ParseEnv function returns the error immediately.
func (e *Env) ParseEnv() error {
	if err := envconfig.Process("", e); err != nil {
		return err
	}
	if err := e.validate(); err != nil {
		return err
	}
	return nil
}

// validate validates upper and lower limit of configurations.
func (e *Env) validate() error {
	validations := []struct {
		proposition bool
		message     string
	}{
		{
			e.VerifyInterval.Minutes() >= lowerIntervalMinutes,
			fmt.Sprintf("INTERVAL must be more than %d minutes", lowerIntervalMinutes),
		},
		{
			e.VerifyInterval.Hours() <= upperIntervalHours,
			fmt.Sprintf("INTERVAL must be less than %d hours", upperIntervalHours),
		},
		{
			e.AlertThreshold.Hours() >= lowerThresholdHours,
			fmt.Sprintf("THRESHOLD must be more than %d hours", lowerThresholdHours),
		},
	}

	for _, v := range validations {
		if !v.proposition {
			return fmt.Errorf(v.message)
		}
	}

	return nil
}
