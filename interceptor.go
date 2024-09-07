package interceptor

import (
	"net/http"
	"net/url"
)

// Pipeline is a wrapper around an http.RoundTripper (also known as a transport)
// that executes a series of Interceptors added via the Use method.
type Pipeline struct {
	// interceptors is a stack of interceptors that are called on every request.
	interceptors []Interceptor

	// Transport is the underlying http.RoundTripper. If nil, http.DefaultTransport is used.
	Transport http.RoundTripper
}

// RoundTrip executes the request using the Pipeline's interceptors and the
// underlying Transport. It implements the http.RoundTripper interface.
func (t *Pipeline) RoundTrip(req *http.Request) (*http.Response, error) {
	var transport http.RoundTripper = t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Wrap transport in reverse order so that execution is in original order
	for i := len(t.interceptors) - 1; i >= 0; i-- {
		transport = t.interceptors[i](transport)
	}

	return transport.RoundTrip(req)
}

// Use appends one or more Interceptors to the Pipeline, allowing them to
// modify or inspect requests before passing them to the underlying transport.
func (t *Pipeline) Use(interceptors ...Interceptor) {
	t.interceptors = append(t.interceptors, interceptors...)
}

// Interceptor defines a function that wraps an http.RoundTripper,
// allowing custom behavior to be injected into the request lifecycle.
type Interceptor func(http.RoundTripper) http.RoundTripper

// RoundTripperFunc is an adapter to allow the use of ordinary functions
// as http.RoundTripper. If f is a function with the appropriate signature,
// RoundTripperFunc(f) is an http.RoundTripper that calls f.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip calls f(req), making RoundTripperFunc implement http.RoundTripper.
func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// BaseURL returns an Interceptor that ensures all outgoing requests use
// the given baseURL. If the request URL already has a scheme, it is left unchanged.
func BaseURL(baseURL url.URL) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			// If the request URL has a scheme, leave it unchanged.
			if req.URL.Scheme != "" {
				return next.RoundTrip(req)
			}
			// Modify the request URL to include the base URL.
			req.URL.Path = baseURL.JoinPath(req.URL.Path).Path
			req.URL = baseURL.ResolveReference(req.URL)
			return next.RoundTrip(req)
		})
	}
}

// Header returns an Interceptor that adds or overrides a header with
// the specified key and value on each request.
func Header(key string, value string) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			// Set the header if the key is not empty.
			if key != "" {
				req.Header.Set(key, value)
			}
			return next.RoundTrip(req)
		})
	}
}
