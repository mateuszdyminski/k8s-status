package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// ETCDConfig defines a set of configuration parameters for accessing
// etcd endpoints
type ETCDConfig struct {
	// Endpoints lists etcd server endpoints
	Endpoints []string
	// CAFile is an SSL Certificate Authority file used to secure
	// communication with etcd
	CAFile string
	// CertFile is an SSL certificate file used to secure
	// communication with etcd
	CertFile string
	// KeyFile is an SSL key file used to secure communication with etcd
	KeyFile string
	// InsecureSkipVerify controls whether a client verifies the
	// server's certificate chain and host name.
	InsecureSkipVerify bool
}

// defaultTLSHandshakeTimeout specifies the default maximum amount of time
// spent waiting to for a TLS handshake
const defaultTLSHandshakeTimeout = 10 * time.Second

// defaultDialTimeout is the default maximum amount of time a dial will wait for
// a connect to complete.
const defaultDialTimeout = 30 * time.Second

// defaultKeepAlive specifies the default keep-alive period for an active
// network connection.
const defaultKeepAlivePeriod = 30 * time.Second

// EtcdChecker is an HTTPResponseChecker that interprets results from
// an etcd HTTP-based healthz end-point.
func EtcdChecker(response io.Reader) error {
	payload, err := ioutil.ReadAll(response)
	if err != nil {
		return err
	}

	healthy, err := etcdStatus(payload)
	if err != nil {
		return err
	}

	if !healthy {
		return fmt.Errorf("unexpected etcd status: %s", payload)
	}
	return nil
}

// etcdStatus determines if the specified etcd status value
// indicates a healthy service
func etcdStatus(payload []byte) (healthy bool, err error) {
	result := struct{ Health string }{}
	nresult := struct{ Health bool }{}
	err = json.Unmarshal(payload, &result)
	if err != nil {
		err = json.Unmarshal(payload, &nresult)
	}
	if err != nil {
		return false, err
	}

	return (result.Health == "true" || nresult.Health == true), nil
}

// NewHTTPTransport creates a new http.Transport from the specified
// set of attributes.
// The resulting transport can be used to create an http.Client
func (r *ETCDConfig) NewHTTPTransport() (*http.Transport, error) {
	tlsConfig, err := r.clientConfig()
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   defaultDialTimeout,
			KeepAlive: defaultKeepAlivePeriod,
		}).Dial,
		TLSHandshakeTimeout: defaultTLSHandshakeTimeout,
		TLSClientConfig:     tlsConfig,
	}

	return transport, nil
}

// clientConfig generates a tls.Config object for use by an HTTP client.
func (r *ETCDConfig) clientConfig() (*tls.Config, error) {
	if r.empty() {
		return nil, nil
	}
	cert, err := ioutil.ReadFile(r.CertFile)
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(r.KeyFile)
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		MinVersion:         tls.VersionTLS10,
		InsecureSkipVerify: r.InsecureSkipVerify,
	}
	config.RootCAs, err = newCertPool([]string{r.CAFile})
	if err != nil {
		return nil, err
	}
	return config, nil
}

// Empty determines if the configuration is empty
func (r *ETCDConfig) empty() bool {
	return r.CAFile == "" && r.CertFile == "" && r.KeyFile == ""
}

// newCertPool creates x509 certPool with provided CA files.
func newCertPool(CAFiles []string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()

	for _, CAFile := range CAFiles {
		pemByte, err := ioutil.ReadFile(CAFile)
		if err != nil {
			return nil, err
		}

		for {
			var block *pem.Block
			block, pemByte = pem.Decode(pemByte)
			if block == nil {
				break
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			certPool.AddCert(cert)
		}
	}

	return certPool, nil
}
