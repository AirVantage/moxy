// A moxy authentication plugin which always says yeah!
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
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
