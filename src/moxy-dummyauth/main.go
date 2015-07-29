// A moxy authentication plugin which always says yeah!
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type connInfo struct {
	Host string
	Port int
}

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: moxy-dummyauth [unix domain socket file]")
		os.Exit(-1)
	}

	log.Println("starting the dummy auth", os.Args[1])

	l, err := net.Listen("unix", os.Args[1])

	if err != nil {
		panic(err)
	}

	defer l.Close()

	// hook on kill signal  for cleaing the socket file
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		// Wait for a SIGINT or SIGKILL:
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		// Stop listening (and unlink the socket if unix type):
		l.Close()
		// And we're done:
		os.Exit(0)
	}(sigc)

	connInfo := mqttConnectionInfo()

	for {
		fd, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go server(fd, connInfo)
	}

}

func server(c net.Conn, info connInfo) {
	var authRequest struct {
		Password string
		UserName string
	}
	defer c.Close()

	dec := gob.NewDecoder(c)

	err := dec.Decode(&authRequest)
	if err != nil {
		log.Fatal(err)
	}

	var res struct {
		Success      bool
		ErrorMessage string
		Host         string
		Port         int
		Metadata     map[string]interface{}
		Topics       map[string]uint
	}

	res.Success = true
	res.ErrorMessage = ""
	res.Host, res.Port = info.Host, info.Port

	enc := gob.NewEncoder(c)
	err = enc.Encode(res)
	if err != nil {
		log.Fatal(err)
	}

}

// Read connection info from MQTT_URL env variable (eg "localhost:1884").
// Defaults to "iot.eclipse.org:1883" is variable is not set.
func mqttConnectionInfo() connInfo {
	host, port := "iot.eclipse.org", 1883

	mqttURL := os.Getenv("MQTT_URL")
	if mqttURL != "" {

		mqttArgs := strings.Split(mqttURL, ":")

		if len(mqttArgs) != 2 {
			log.Fatal("Bad value for environment variable MQTT_URL, must be populated with the base URL for reaching the backend MQTT broker. ex : \"localhost:1884\"")
		}

		host = mqttArgs[0]

		var err error
		port, err = strconv.Atoi(mqttArgs[1])
		if err != nil {
			log.Fatal("MQTT_URL port must be a number", err)
		}
	}

	return connInfo{Host: host, Port: port}

}
