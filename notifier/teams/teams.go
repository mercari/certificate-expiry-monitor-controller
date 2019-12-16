package teams

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

const (
	// notifierName used by pattern match when parse interpret options.
	notifierName = "teams"
)

// MsTeams struct implements notifier.Notifier interface.
// MsTeams struct sends alert over RESTful API using http libs.
type MsTeams struct {
	WebhookEndpoint string
}

// NewNotifier function returns new instance of MsTeams.
// The webhook endpoint is fixed when initialize.
func NewNotifier(webhook string) (notifier.Notifier, error) {
	if webhook == "" {
		return nil, errors.New("webhook endpoint is missing")
	}

	return &MsTeams{
		WebhookEndpoint: webhook,
	}, nil
}

// String function used by pattern match when parse interpret options.
func String() string {
	return notifierName
}

// Alert defined by notifier.Notifier interface.
// This implementation post message that includes infromation about ingress and TLS and those deadline.
func (s *MsTeams) Alert(expiration time.Time, ingress *source.Ingress, tls *source.IngressTLS, opt notifier.Option) error {
	return send(s.WebhookEndpoint, newMessageCard(expiration, ingress, tls, opt.AlertLevel))
}

// Http Post request with the MessageCard as payload
func send(endpoint string, msg MessageCard) error {
	enc, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(enc)
	_, err = http.Post(endpoint, "application/json", b)
	if err != nil {
		return err
	}

	return nil
}
