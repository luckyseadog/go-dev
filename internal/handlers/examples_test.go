package handlers

import (
	"bytes"
	"net/http"
)

// ExampleHandlerDefault demonstrates how to make a simple HTTP GET to HandlerDefault.
func ExampleHandlerDefault() {
	response, err := http.Get("http://example.com")
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()
}

// ExampleHandlerGet demonstrates how to make a simple HTTP GET to HandlerGet.
// It sends GET requests to example URLs for both gauge and counter metric types.
func ExampleHandlerGet() {
	response, err := http.Get("http://example.com/value/gauge/gaugeMetric")
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()

	response, err = http.Get("http://example.com/value/counter/counterMetric")
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()
}

// ExampleHandlerPing demonstrates how to send an HTTP GET request to check the server's status.
func ExampleHandlerPing() {
	response, err := http.Get("http://example.com/ping")
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()
}

// ExampleHandlerUpdate demonstrates how to send an HTTP POST request to update a metric value.
func ExampleHandlerUpdate() {
	response, err := http.Post("http://example.com/update/gauge/gaugeMetric/10.0", "text/plain", nil)
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()
}

// ExampleHandlerUpdateJSON demonstrates how to send HTTP POST requests with JSON data to update metric value.
func ExampleHandlerUpdateJSON() {
	updateData := []byte(`{"id":"gaugeMetric", "type": "gauge", "value": 10.0}`)
	response, err := http.Post("http://example.com/update", "application/json", bytes.NewBuffer(updateData))
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()

	updateData = []byte(`{"id":"gaugeMetric", "type": "counter", "delta": 2}`)
	response, err = http.Post("http://example.com/update", "application/json", bytes.NewBuffer(updateData))
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()
}

// ExampleHandlerUpdatesJSON demonstrates how to send an HTTP POST request with an array of JSON updates.
func ExampleHandlerUpdatesJSON() {
	updateData := []byte(`[{"id":"gaugeMetric", "type": "gauge", "value": 10.0}, 
                           {"id":"gaugeMetric", "type": "counter", "delta": 2}]`)
	response, err := http.Post("http://example.com/updates", "application/json", bytes.NewBuffer(updateData))
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()
}

// ExampleHandlerValueJSON demonstrates how to send an HTTP POST request to retrieve a metric value using JSON data.
func ExampleHandlerValueJSON() {
	valueData := []byte(`{"id":"gaugeMetric", "type": "gauge"}`)
	response, err := http.Post("http://example.com/value", "application/json", bytes.NewBuffer(valueData))
	if err != nil {
		// skip error handling.
	}
	response.Body.Close()
}
