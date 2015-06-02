// Main server binary for the moxy MQTT reverse proxy.
package main

import (
	"flag"
	"fmt"
	"github.com/airvantage/moxy"
	"github.com/airvantage/moxy/plugin/auth"
	"github.com/airvantage/moxy/plugin/filter"
	"os"
	"strconv"
	"strings"
)

var listen string
var debug bool
var trace bool
var authplug string
var filterplugs string

func setupFlags() {
	flag.StringVar(&listen, "listen", "0.0.0.0:1883", "the MQTT address and port to listen")
	flag.BoolVar(&debug, "v", false, "verbose debug information")
	flag.BoolVar(&trace, "t", false, "very verbose trace of every communication")
	flag.StringVar(&authplug, "auth", "moxy-dummyauth", "the plugin in charge of authentication")
	flag.StringVar(&filterplugs, "filters", "", "the plugin list in charge of filtering seperated by ':'")
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

	filters := []moxy.MqttFilter{}
	if len(strings.TrimSpace(filterplugs)) > 0 {
		filterstrs := strings.Split(filterplugs, ":")
		filters = make([]moxy.MqttFilter, len(filterstrs))

		for i, v := range filterstrs {
			filters[i] = filter.NewFilterPlugin(v, "F"+strconv.Itoa(i))
		}
	}

	s := moxy.NewServer(debug, trace, listen, authPlugin, filters)

	if err := s.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}

	fmt.Println("bye")
}
