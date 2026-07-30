package main

import "os"

type ClientID string

//go:structinit
type Config struct { // want Config:"HasFieldOrder\\[ClientID ServerURL ServerGRPCPort ServerRestPort\\]"
	// ClientID to be used to register in R-Event.
	ClientID ClientID
	// ServerURL R-Event server url.
	ServerURL string
	// ServerGRPCPort R-Event gRPC port.
	ServerGRPCPort int
	// ServerRestPort R-Event REST port.
	ServerRestPort int
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "local"
	}

	return Config{ // want "fields are not initialized in declared order"
		ServerURL:      "http://localhost",
		ServerGRPCPort: 10000,
		ServerRestPort: 10001,
		ClientID:       ClientID(hostname),
	}
}
