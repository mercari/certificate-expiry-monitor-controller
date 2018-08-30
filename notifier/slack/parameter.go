package slack

import (
	"fmt"
	"strings"
	"time"

	libSlack "github.com/nlopes/slack"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

// newPostParameters creates params to pass PostMessage function.
func newPostParameters(expiration time.Time, ingress *source.Ingress, tls *source.IngressTLS, alertLevel notifier.AlertLevel) libSlack.PostMessageParameters {
	var color, preText string

	switch alertLevel {
	case notifier.AlertLevelCritical:
		color = "danger"
		days := int64(time.Since(expiration).Hours() / 24)
		preText = fmt.Sprintf("[CRITICAL] TLS certificate already expired at %d days ago", days)
	case notifier.AlertLevelWarning:
		color = "warning"
		days := int64(time.Until(expiration).Hours() / 24)
		preText = fmt.Sprintf("[WARNING] TLS certificate will expire within %d days", days)
	}

	return libSlack.PostMessageParameters{
		Username: "Certificate Expiry Monitor",
		Attachments: []libSlack.Attachment{
			libSlack.Attachment{
				Color:   color,
				Pretext: preText,
				Fields:  newAttachmentFields(ingress.ClusterName, ingress.Namespace, ingress.Name, tls.SecretName, expiration, tls.Endpoints),
			},
		},
	}
}

// newAttachmentFields creates attachment filed slice that may included in PostParameter.
func newAttachmentFields(cluster string, namespace string, name string, secret string, expiration time.Time, endpoints []*source.TLSEndpoint) []libSlack.AttachmentField {
	hosts := make([]string, len(endpoints))
	for i, e := range endpoints {
		hosts[i] = e.Hostname + ":" + e.Port
	}

	return []libSlack.AttachmentField{
		libSlack.AttachmentField{Title: "Cluster", Value: cluster},
		libSlack.AttachmentField{Title: "Namespace", Value: namespace},
		libSlack.AttachmentField{Title: "Ingress", Value: name},
		libSlack.AttachmentField{Title: "TLS secret name", Value: secret},
		libSlack.AttachmentField{Title: "Expiration", Value: expiration.Format(time.RFC822)},
		libSlack.AttachmentField{Title: "Hosts", Value: strings.Join(hosts, "\n")},
	}
}
