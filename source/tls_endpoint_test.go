package source

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetCertificates(t *testing.T) {
	testWithTLSServer(func(server *httptest.Server) {
		u, _ := url.Parse(server.URL)

		// When using available endpoint
		availableEndpoint := NewTLSEndpoint(u.Hostname(), u.Port())
		certs, err := availableEndpoint.GetCertificates()
		if err != nil || len(certs) == 0 {
			t.Fatalf("Cannot get certificate when using available endpoint %s", u.Hostname()+":"+u.Port())
		}

		// When using unavailable endpoint
		unavailableEndpoint := NewTLSEndpoint("dummy.localhost.local", "443")
		certs, err = unavailableEndpoint.GetCertificates()
		if err == nil || len(certs) != 0 {
			t.Fatalf("Unexpected result when using unavailable endpoint %s", u.Hostname()+":"+u.Port())
		}
	})
}

func testWithTLSServer(f func(server *httptest.Server)) {
	server := httptest.NewTLSServer(http.NewServeMux())
	defer server.Close()
	f(server)
}

func TestNewTLSEndpoint(t *testing.T) {
	type testArg struct {
		Hostname string
		Port     string
	}

	tests := []struct {
		args         testArg
		expectedPort string
	}{
		{
			args:         testArg{Hostname: "example.com", Port: "5512"},
			expectedPort: "5512",
		},
		{
			args:         testArg{Hostname: "*.example.com", Port: ""},
			expectedPort: DefaultPortNumber,
		},
	}

	for _, test := range tests {
		e := NewTLSEndpoint(test.args.Hostname, test.args.Port)
		if test.expectedPort != e.Port {
			t.Fatalf("Unexpected port number: %s", e.Port)
		}
	}
}
