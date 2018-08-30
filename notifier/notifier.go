package notifier

import (
	"time"

	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

// AlertLevel expresses notification level when uses in Alert() function
// notification level hierarky : (high) CRITICAL > WARNING (low)
type AlertLevel int

const (
	// AlertLevelWarning express that notification level is `WARNING`.
	AlertLevelWarning AlertLevel = iota
	// AlertLevelCritical express that notification level is `CRITICAL`.
	AlertLevelCritical
)

// Option struct provides configration about notification.
type Option struct {
	AlertLevel AlertLevel
}

// Notifier interface expresses the notification services that able to send Alert.
// If controller triggers Alert, Notifier send details about certificate's expirarion to own service.
type Notifier interface {
	Alert(time.Time, *source.Ingress, *source.IngressTLS, Option) error
}
