// Main server binary for the moxy MQTT reverse proxy.
package main

import (
	"flag"
	"fmt"
	"github.com/airvantage/moxy"
	"github.com/airvantage/moxy/plugin/auth"
	"os"
)

var listen string
var debug bool
var trace bool
var authplug string

func setupFlags() {
	flag.StringVar(&listen, "listen", "0.0.0.0:1883", "the MQTT address and port to listen")
	flag.BoolVar(&debug, "v", false, "verbose debug information")
	flag.BoolVar(&trace, "t", false, "very verbose trace of every communication")
	flag.StringVar(&authplug, "auth", "moxy-dummyauth", "the plugin in charge of authnetication")
	flag.Parse()
}

func main() {
	fmt.Println("MQTT proxy")
	setupFlags()

	if debug {
		fmt.Println("verbose mode enabled")
	}

	if trace {
		fmt.Println("low level trace of communications enabled")
	}

	authPlugin := auth.NewAuthPlugin(authplug)

	s := moxy.NewServer(debug, trace, listen, authPlugin)

	if err := s.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}

	fmt.Println("bye")
}
