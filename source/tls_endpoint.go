package source

import (
	"crypto/tls"
	"crypto/x509"
	"strings"
)

var (
	// Allow certificate that signed by unknown authority.
	// Controller only concerns expiration of certificate.
	defaultTLSConfig = tls.Config{InsecureSkipVerify: true}

	// DefaultPortNumber exposes default port number to testing
	// TODO: Support port numbers other than :443
	DefaultPortNumber = "443"
)

// TLSEndpoint expressses https endpoint that using TLS.
type TLSEndpoint struct {
	Hostname string
	Port     string
}

// NewTLSEndpoint creates new TLSEndpoint instance.
// If port number is empty, set DefaultPortNumber instead.
func NewTLSEndpoint(host string, port string) *TLSEndpoint {
	if port == "" {
		port = DefaultPortNumber
	}

	return &TLSEndpoint{
		Hostname: host,
		Port:     port,
	}
}

// GetCertificates tries to get certificates from endpoint using tls.Dial
func (e *TLSEndpoint) GetCertificates() ([]*x509.Certificate, error) {

	// We cannot connect to Hostnames with wildcards, so replacing with www.
	hostName := strings.Replace(e.Hostname, "*", "www", -1)
	conn, err := tls.Dial("tcp", hostName+":"+e.Port, &defaultTLSConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.ConnectionState().PeerCertificates, nil
}
