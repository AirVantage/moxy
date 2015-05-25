// A moxy authentication plugin which always says yeah!
package main

import (
	"encoding/gob"
	"fmt"
	"github.com/airvantage/moxy"
	"github.com/airvantage/moxy/plugin/auth"
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
	var authRequest auth.AuthCall

	defer c.Close()

	dec := gob.NewDecoder(c)

	err := dec.Decode(&authRequest)
	if err != nil {
		log.Fatal(err)
	}

	res := moxy.AuthResult{true, "", "iot.eclipse.org", 1883}

	enc := gob.NewEncoder(c)
	err = enc.Encode(res)
	if err != nil {
		log.Fatal(err)
	}

}
