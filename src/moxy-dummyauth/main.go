// A moxy authentication plugin which always says yeah!
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: moxy-auth [unix domain socket file]")
		os.Exit(-1)
	}

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

	for {
		fd, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go server(fd)
	}

}

func server(c net.Conn) {
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
	}

	res.Success = true
	res.ErrorMessage = ""
	res.Host = "iot.eclipse.org"
	res.Port = 1883

	enc := gob.NewEncoder(c)
	err = enc.Encode(res)
	if err != nil {
		log.Fatal(err)
	}

}
