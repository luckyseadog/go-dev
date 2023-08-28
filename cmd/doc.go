/*
Package provides entry points for starting metric service.
This is achieved by running HTTP server and HTTP client.

To build the agent binary:
	go build ./agent

To build the server binary:
	go build ./server

When starting a binary file, you should include necessary flags.
Flag details can be found in the files:
- ./server/server.go
- ./agent/agent.go

The package also contains a multichecker that checks for common mistakes in code.
To run the multichecker:

Build the multichecker binary:
	go build ./staticlint -o multichecker

Run the multichecker with the root of the project as an argument:
	./multichecker <root of the project>
*/

package main
