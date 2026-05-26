// === FILE: main.go ===
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create and run relay
	relay, err := NewRelay(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating relay: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	if err := relay.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Relay error: %v\n", err)
		os.Exit(1)
	}
}
