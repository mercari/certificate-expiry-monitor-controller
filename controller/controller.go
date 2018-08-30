package controller

import (
	"crypto/x509"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"

	"k8s.io/client-go/kubernetes"
)

// onIteration is called when the starting runOnce. Used in testing.
var onIteration = func() {}

// Controller identifies an instance of a controller.
// It created by calling NewController function.
type Controller struct {
	Logger         *zap.Logger
	Source         *source.Source
	VerifyInterval time.Duration
	AlertThreshold time.Duration
	Notifiers      []notifier.Notifier
}

// NewController function validates arguments and
// returns new controller pointer constructed by arguments.
func NewController(
	logger *zap.Logger,
	clientSet kubernetes.Interface,
	interval time.Duration,
	threshold time.Duration,
	notifiers []notifier.Notifier,
) (*Controller, error) {
	if clientSet == nil {
		return nil, errors.New("clientSet must be non nil value")
	}

	if logger == nil {
		return nil, errors.New("logger must be non nil value")
	}

	return &Controller{
		Logger:         logger,
		Source:         source.NewSource(clientSet),
		VerifyInterval: interval,
		AlertThreshold: threshold,
		Notifiers:      notifiers,
	}, nil
}

// Run function starts execution loop that executes runOnce at VerifyInterval.
// If stopCh receives message, Run function terminates execution loop.
func (c *Controller) Run(stopCh chan struct{}) {
	c.Logger.Info("Starting controller...")

	for {
		onIteration()
		currentTime := time.Now()
		err := c.runOnce(currentTime)
		if err != nil {
			c.Logger.Error("Failed to run runOnce: %s", zap.Error(err))
		}

		select {
		case <-time.After(c.VerifyInterval):
		case <-stopCh:
			c.Logger.Info("Terminating controller...")
			return
		}
	}
}

func (c *Controller) runOnce(currentTime time.Time) error {
	ingresses, err := c.Source.Ingresses()
	if err != nil {
		return err
	}

	thresholdTime := currentTime.Add(c.AlertThreshold)

	for _, ingress := range ingresses {
		for _, tls := range ingress.TLS {

			var certificates []*x509.Certificate
			for _, e := range tls.Endpoints {
				certificates, err = e.GetCertificates()
				if err != nil {
					c.Logger.Warn("Detect error when GetCertificates()", zap.String("host", e.Hostname+":"+e.Port), zap.Error(err))
					continue
				}

				// Controller assumes that IngressTLS has one certificate chain and all endpoints associated it.
				// So, if detect certificate chain the first time, break loop and verify it.
				break
			}

			if len(certificates) == 0 {
				c.Logger.Warn("Remote endpoints has no certificates, but endpoints enabled TLS")
				continue
			}

			// certs[0] is end-user certificate.
			// TODO: able to verify root and intermediate certificate by option
			expiration := certificates[0].NotAfter

			opt := notifier.Option{}
			if expiration.Before(currentTime) {
				// If certificate has been expired.
				opt.AlertLevel = notifier.AlertLevelCritical
			} else if expiration.Before(thresholdTime) {
				// If certificates has been reached the thresholdTime.
				opt.AlertLevel = notifier.AlertLevelWarning
			} else {
				// This expiration has not reached the threshold.
				continue
			}

			// Send Alert to all notifiers.
			for _, notifier := range c.Notifiers {
				err := notifier.Alert(expiration, ingress, tls, opt)

				if err != nil {
					c.Logger.Warn("Failed to send Alert", zap.Error(err))
				}
			}
		}
	}

	return nil
}
