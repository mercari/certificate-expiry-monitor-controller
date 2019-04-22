package teams

import (
	"fmt"
	"strings"
	"time"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

const (
	messageCardContext = "https://schema.org/extensions"
	messageCardType    = "MessageCard"
	dangerColor        = "FF0000"
	warningColor       = "FFA500"
)

// MessageCard struct for representing information in a rich format.
type MessageCard struct {
	Context    string    `json:"@context"`
	Type       string    `json:"@type"`
	Title      string    `json:"Title"`
	Summary    string    `json:"Summary"`
	ThemeColor string    `json:"themeColor"`
	Sections   []Section `json:"sections"`
	Version    string    `json:"version" default:"1.0"`
}

type Section struct {
	Facts []MessageFact `json:"facts"`
}

type MessageFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func newMessageCard(expiration time.Time, ingress *source.Ingress, tls *source.IngressTLS, alertLevel notifier.AlertLevel) MessageCard {
	var color, preText string

	switch alertLevel {
	case notifier.AlertLevelCritical:
		color = dangerColor
		days := int64(time.Since(expiration).Hours() / 24)
		preText = fmt.Sprintf("[CRITICAL] TLS certificate already expired at %d days ago", days)
	case notifier.AlertLevelWarning:
		color = warningColor
		days := int64(time.Until(expiration).Hours() / 24)
		preText = fmt.Sprintf("[WARNING] TLS certificate will expire within %d days", days)
	}

	hosts := make([]string, len(tls.Endpoints))
	for i, e := range tls.Endpoints {
		hosts[i] = e.Hostname + ":" + e.Port
	}

	facts := []MessageFact{
		{Name: "Cluster", Value: ingress.ClusterName},
		{Name: "Namespace", Value: ingress.Namespace},
		{Name: "Ingress", Value: ingress.Name},
		{Name: "TLS secret name", Value: tls.SecretName},
		{Name: "Expiration", Value: expiration.Format(time.RFC822)},
		{Name: "Hosts", Value: strings.Join(hosts, ", ")},
	}

	return MessageCard{
		Context:    messageCardContext,
		Type:       messageCardType,
		ThemeColor: color,
		Title:      preText,
		Summary:    preText,
		Sections: []Section{
			{
				Facts: facts,
			},
		},
	}
}
