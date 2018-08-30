package log

import (
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

const (
	// alertLogLevel used when create new log notifier
	alertLogLevel = "ERROR"

	// notifierName used by pattern match when parse interpret options.
	notifierName = "log"
)

// Log struct implements notifier.Notifier interface.
// Log struct output alert information using application logger.
type Log struct {
	Logger *zap.Logger
}

// NewNotifier function returns new instance of Log.
func NewNotifier(logger *zap.Logger) notifier.Notifier {
	return &Log{
		Logger: logger,
	}
}

// String function used by pattern match when parse interpret options.
func String() string {
	return notifierName
}

// AlertLogLevel called when create new log notifier
func AlertLogLevel() string {
	return alertLogLevel
}

// Alert defined by notifier.Notifier interface.
// This function create and print fields using log package.
func (log *Log) Alert(expiration time.Time, ingress *source.Ingress, tls *source.IngressTLS, opt notifier.Option) error {
	fields := loggingFields(ingress.ClusterName, ingress.Namespace, ingress.Name, tls.SecretName, expiration, tls.Endpoints, opt.AlertLevel)
	log.Logger.Error("ALERT", fields...)
	return nil
}

func loggingFields(
	cluster string,
	namespace string,
	name string,
	secret string,
	expiration time.Time,
	endpoints []*source.TLSEndpoint,
	alertLevel notifier.AlertLevel,
) []zapcore.Field {

	var level string
	switch alertLevel {
	case notifier.AlertLevelWarning:
		level = "WARNING"
	case notifier.AlertLevelCritical:
		level = "CRITICAL"
	}

	hosts := make([]string, len(endpoints))
	for i, e := range endpoints {
		hosts[i] = e.Hostname + ":" + e.Port
	}

	return []zapcore.Field{
		zap.String("Level", level),
		zap.String("ClusterName", cluster),
		zap.String("Namespace", namespace),
		zap.String("Ingress", name),
		zap.String("TLS Secret name", secret),
		zap.String("Expiration", expiration.Format(time.RFC822)),
		zap.String("Hosts", strings.Join(hosts, ",")),
	}
}
