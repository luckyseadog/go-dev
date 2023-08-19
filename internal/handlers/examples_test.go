package handlers

import (
	"bytes"
	"net/http"
)

// ExampleHandlerDefault demonstrates how to make a simple HTTP GET to HandlerDefault.
func ExampleHandlerDefault() {
	response, err := http.Get("http://example.com")
	_ = response // Ignoring the response for this example
	_ = err      // Ignoring the error for this example
}

// ExampleHandlerGet demonstrates how to make a simple HTTP GET to HandlerGet.
// It sends GET requests to example URLs for both gauge and counter metric types.
func ExampleHandlerGet() {
	response, err := http.Get("http://example.com/value/gauge/gaugeMetric")
	_ = response
	_ = err

	response, err = http.Get("http://example.com/value/counter/counterMetric")
	_ = response
	_ = err
}

// ExampleHandlerPing demonstrates how to send an HTTP GET request to check the server's status.
func ExampleHandlerPing() {
	response, err := http.Get("http://example.com/ping")
	_ = response
	_ = err
}

// ExampleHandlerUpdate demonstrates how to send an HTTP POST request to update a metric value.
func ExampleHandlerUpdate() {
	response, err := http.Post("http://example.com/update/gauge/gaugeMetric/10.0", "text/plain", nil)
	_ = response
	_ = err
}

// ExampleHandlerUpdateJSON demonstrates how to send HTTP POST requests with JSON data to update metric value.
func ExampleHandlerUpdateJSON() {
	updateData := []byte(`{"id":"gaugeMetric", "type": "gauge", "value": 10.0}`)
	response, err := http.Post("http://example.com/update", "application/json", bytes.NewBuffer(updateData))
	_ = response // Ignoring the response for this example
	_ = err      // Ignoring the error for this example

	updateData = []byte(`{"id":"gaugeMetric", "type": "counter", "delta": 2}`)
	response, err = http.Post("http://example.com/update", "application/json", bytes.NewBuffer(updateData))
	_ = response // Ignoring the response for this example
	_ = err      // Ignoring the error for this example
}

// ExampleHandlerUpdatesJSON demonstrates how to send an HTTP POST request with an array of JSON updates.
func ExampleHandlerUpdatesJSON() {
	updateData := []byte(`[{"id":"gaugeMetric", "type": "gauge", "value": 10.0}, 
                           {"id":"gaugeMetric", "type": "counter", "delta": 2}]`)
	response, err := http.Post("http://example.com/updates", "application/json", bytes.NewBuffer(updateData))
	_ = response // Ignoring the response for this example
	_ = err      // Ignoring the error for this example
}

// ExampleHandlerValueJSON demonstrates how to send an HTTP POST request to retrieve a metric value using JSON data.
func ExampleHandlerValueJSON() {
	valueData := []byte(`{"id":"gaugeMetric", "type": "gauge"}`)
	response, err := http.Post("http://example.com/value", "application/json", bytes.NewBuffer(valueData))
	_ = response
	_ = err
}

