package runner

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

const healthzCheckTimeout = 1 * time.Second

// HTTPResponseChecker is a function that can validate service health
// from the provided response
type HTTPResponseChecker func(response io.Reader) error

// HTTPHealthzChecker is a Checker that can validate service health over HTTP
type HTTPHealthzChecker struct {
	name    string
	URL     string
	client  *http.Client
	checker HTTPResponseChecker
}

// Name returns the name of this checker
func (r *HTTPHealthzChecker) Name() string { return r.name }

// Check runs an HTTP check and reports errors to the specified Reporter
func (r *HTTPHealthzChecker) Check(ctx context.Context, reporter Reporter) {
	req, err := http.NewRequest("GET", r.URL, nil)
	if err != nil {
		reporter.Add(NewProbeFromErr(r.name, noErrorDetail, fmt.Errorf("failed to create request: %v", err)))
		return
	}
	req = req.WithContext(ctx)
	resp, err := r.client.Do(req)
	if err != nil {
		reporter.Add(NewProbeFromErr(r.name, noErrorDetail, fmt.Errorf("healthz check failed: %v", err)))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		reporter.Add(&Probe{
			Checker: r.name,
			Status:  ProbeFailed,
			Error: fmt.Errorf("unexpected HTTP status: %s",
				http.StatusText(resp.StatusCode)).Error(),
			Code: strconv.Itoa(resp.StatusCode),
		})
		return
	}
	if err = r.checker(resp.Body); err != nil {
		reporter.Add(NewProbeFromErr(r.name, noErrorDetail, err))
		return
	}
	reporter.Add(&Probe{
		Checker: r.name,
		Status:  ProbeRunning,
	})
}

// NewHTTPHealthzChecker creates a Checker for an HTTP health endpoint
// using the specified URL and a custom response checker
func NewHTTPHealthzChecker(name, URL string, checker HTTPResponseChecker) Checker {
	defaultTransport := http.RoundTripper(nil)
	return NewHTTPHealthzCheckerWithTransport(name, URL, defaultTransport, checker)
}

// NewUnixSocketHealthzChecker returns a new Checker that tests
// the specified unix domain socket path and URL
func NewUnixSocketHealthzChecker(name, URL, socketPath string, checker HTTPResponseChecker) Checker {
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
	return NewHTTPHealthzCheckerWithTransport(name, URL, transport, checker)
}

// NewHTTPHealthzCheckerWithTransport creates a Checker for an HTTP health endpoint
// using the specified transport, URL and a custom response checker
func NewHTTPHealthzCheckerWithTransport(name, URL string, transport http.RoundTripper, checker HTTPResponseChecker) Checker {
	client := &http.Client{
		Transport: transport,
		Timeout:   healthzCheckTimeout,
	}
	return &HTTPHealthzChecker{
		name:    name,
		URL:     URL,
		client:  client,
		checker: checker,
	}
}
