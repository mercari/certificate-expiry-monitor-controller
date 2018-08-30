package slack

import (
	"errors"
	"time"

	libSlack "github.com/nlopes/slack"
	"go.uber.org/ratelimit"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

const (
	// The Slack API has restriction on sending messages. So, notifier send once per second
	// See also: https://api.slack.com/docs/rate-limits
	sendPerSecond = 1

	// notifierName used by pattern match when parse interpret options.
	notifierName = "slack"
)

// API interface defines slack's API behavior.
// API interface defined to wrap the library: github.com/nlopes/slack
type API interface {
	PostMessage(string, string, libSlack.PostMessageParameters) (string, string, error)
}

// Slack struct implements notifier.Notifier interface.
// Slack struct sends alert over RESTful API.
type Slack struct {
	APIClient   API
	ChannelName string
	RateLimiter ratelimit.Limiter
}

// NewNotifier function returns new instance of Slack.
// The destination channel is fixed when initialize.
func NewNotifier(token string, channel string) (notifier.Notifier, error) {
	if token == "" {
		return nil, errors.New("token is missing")
	}

	if channel == "" {
		return nil, errors.New("channel is missing")
	}

	return &Slack{
		APIClient:   libSlack.New(token),
		ChannelName: channel,
		RateLimiter: ratelimit.New(sendPerSecond),
	}, nil
}

// String function used by pattern match when parse interpret options.
func String() string {
	return notifierName
}

// Alert defined by notifier.Notifier interface.
// This implementation post message that includes infromation about ingress and TLS and those deadline.
func (s *Slack) Alert(expiration time.Time, ingress *source.Ingress, tls *source.IngressTLS, opt notifier.Option) error {
	params := newPostParameters(expiration, ingress, tls, opt.AlertLevel)
	return s.postWithRateLimiter(s.ChannelName, "", params)
}

func (s *Slack) postWithRateLimiter(channel string, message string, params libSlack.PostMessageParameters) error {
	s.RateLimiter.Take()
	_, _, err := s.APIClient.PostMessage(channel, message, params)
	return err
}
