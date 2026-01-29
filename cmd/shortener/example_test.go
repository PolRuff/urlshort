package main

import "fmt"

// Example demonstrates how to interact with the URL shortener service using curl.
//
// To shorten a URL:
//
//	$ curl -X POST -d 'https://example.com' http://localhost:8080/
//	http://localhost:8080/abc123
//
// To follow a shortened URL:
//
//	$ curl -v http://localhost:8080/abc123
//	< HTTP/1.1 307 Temporary Redirect
//	< Location: https://example.com
func Example() {
	// This example shows how to use the service via command line.
	// It is not meant to be executed by 'go test'.
	fmt.Println("See godoc for usage examples with curl.")
	// Output: See godoc for usage examples with curl.
}
