package datadog

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/source"
)

type Datadog struct {
	token    string
	address  string
	tags     []string
	gaugekey string
}

// NewNotifier function returns new instance of Log.
func NewNotifier(token, address string, tagstr string, gaugekey string) (notifier.Notifier, error) {
	if token == "" {
		return nil, errors.New("datadog token is missing")
	}

	if address == "" {
		return nil, errors.New("datadog address is missing")
	}
	tags := strings.Split(tagstr, ",")
	return &Datadog{token: token, address: address, tags: tags, gaugekey: gaugekey}, nil
}

// String function used by pattern match when parse interpret options.
func String() string {
	return "datadog"
}

func pretty(s []string) string {
	return strings.Replace(strings.ToLower(strings.Join(s, " ")), " ", "_", -1)
}

// Alert defined by notifier.Notifier interface.
func (dd *Datadog) Alert(expiration time.Time, ingress *source.Ingress, tls *source.IngressTLS, opt notifier.Option) error {
	c, err := statsd.New(dd.address)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	for _, ep := range tls.Endpoints {
		certs, err := ep.GetCertificates()
		if err != nil {
			log.Fatal(err)
		}
		cert := certs[0] // only report the first cert in the chain
		s := fmt.Sprintf("/c:%v/st:%v/l:%v/o:%v/ou:%v/cn:%v",
			pretty(cert.Subject.Country),
			pretty(cert.Subject.Province),
			pretty(cert.Subject.Locality),
			pretty(cert.Subject.Organization),
			pretty(cert.Subject.OrganizationalUnit),
			strings.ToLower(cert.Subject.CommonName))
		expire := int(expiration.Sub(time.Now()) / 24)

		c.Tags = append(c.Tags, "certificate:"+s)
		c.Tags = append(c.Tags, dd.tags...)
		err = c.Gauge(dd.gaugekey, float64(expire), nil, 1)
		if err != nil {
			log.Fatal(err)
		}

	}
	return nil
}
