# HTTP Interceptors Go

`interceptor` is a Go package that provides a flexible way to chain and apply custom HTTP interceptors to modify outbound requests and their respective responses. You can think of it as middleware for the `http.Client` instead of the `http.Server`.

The package wraps around the standard library's `http.RoundTripper`, allowing you to easily create and use custom interceptors for modifying HTTP request/response behavior.

This package was inspired by middleware functionality for HTTP servers, like in the Go standard library and third-party options like [chi](https://github.com/go-chi/chi). It is also conceptually similar to Square's [OkHttp](https://square.github.io/okhttp/features/interceptors/) library.

## Installation

Install the package using:

```shell
go get github.com/brain-hol/http-interceptors-go
```

Then, import it into your Go code:

```go
import "github.com/brain-hol/http-interceptors-go"
```

## Usage

Below are examples showing how to set up and use the `Pipeline` and `Interceptor` functions.

### Basic Example

```go
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"github.com/brain-hol/http-interceptors-go"
)

func main() {
	// Create a new Pipeline
	pipeline := &interceptor.Pipeline{}

	// Add a BaseURL interceptor to modify requests to use a base URL
	base, _ := url.Parse("https://api.example.com")
	pipeline.Use(interceptor.BaseURL(*base))

	// Add a Header interceptor to include a custom header
	pipeline.Use(interceptor.Header("Authorization", "Bearer my-token"))

	// Create an HTTP client that uses the Pipeline as the transport
	client := &http.Client{
		Transport: pipeline,
	}

	// Send a request through the pipeline
	req, _ := http.NewRequest("GET", "/resource", nil)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Response Status:", resp.Status)
	}
}
```

In this example:

- The `BaseURL` interceptor ensures that all requests are made relative to `https://api.example.com`.
- The `Header` interceptor automatically adds an authorization header to every request.

### Adding Custom Interceptors

You can also define custom interceptors by implementing a function that wraps an `http.RoundTripper`:

```go
func LoggingInterceptor(next http.RoundTripper) http.RoundTripper {
	return interceptor.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		fmt.Println("Sending request to:", req.URL.String())
		return next.RoundTrip(req)
	})
}

func main() {
	// Create a new Pipeline and add the logging interceptor
	pipeline := &interceptor.Pipeline{}
	pipeline.Use(LoggingInterceptor)

	// Add a client with the pipeline transport
	client := &http.Client{
		Transport: pipeline,
	}

	// Send a request and see logging in action
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	client.Do(req)
}
```

## API Documentation

### `Pipeline`

The `Pipeline` struct is the main component of the package, responsible for managing the chain of interceptors and executing them on each HTTP request.

- **`Use(interceptors ...Interceptor)`**: Adds one or more interceptors to the pipeline. Each interceptor will wrap the `http.RoundTripper` and be invoked on each request.
- **`RoundTrip(req *http.Request)`**: Implements the `http.RoundTripper` interface and processes the request through the chain of interceptors.

### `Interceptor`

An `Interceptor` is a function that takes an `http.RoundTripper` and returns a wrapped `http.RoundTripper`, allowing custom logic to be inserted into the request lifecycle.

```go
type Interceptor func(http.RoundTripper) http.RoundTripper
```

### Built-in Interceptors

- **`BaseURL(baseURL url.URL)`**: Ensures all outgoing requests use the provided `baseURL` if no scheme is present in the request URL.
  
- **`Header(key string, value string)`**: Adds or overrides a header with the specified key and value on every request.

### `RoundTripperFunc`

An adapter to allow ordinary functions to satisfy the `http.RoundTripper` interface.

```go
type RoundTripperFunc func(*http.Request) (*http.Response, error)
```
