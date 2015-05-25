package moxy

import (
	"fmt"
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git/packets"
	"log"
	"net"
)

// A MQTT proxy server
type Server struct {
	debug     bool
	trace     bool
	listenStr string
	listener  net.Listener
	auth      Authenticator
}

// NewServer create a new MQTT proxy server, provide the wanted authenticator
func NewServer(debug, trace bool, listen string, auth Authenticator) *Server {
	s := new(Server)
	s.debug = debug
	s.trace = trace
	s.listenStr = listen
	s.auth = auth
	return s
}

// Serve bind the socket and serve as a proxy. Will block until end of the world.
func (s *Server) Serve() error {

	if s.debug {
		fmt.Println("starting the proxy server")
	}

	var err error
	s.listener, err = net.Listen("tcp", s.listenStr)
	if err != nil {
		return err
	}

	// endless accept loop
	for {
		if s.debug {
			fmt.Println("accept")
		}
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		// start serving goroutine
		go s.serve(conn)
	}
}

func (s *Server) serve(conn net.Conn) {
	if s.debug {
		fmt.Printf("new connection: %v\n", conn.RemoteAddr())
	}
	cp, err := packets.ReadPacket(conn)
	if err != nil {
		panic(err)
	}
	connect, ok := cp.(*packets.ConnectPacket)
	if !ok {
		log.Printf("we want a Connect packet %v", cp)
		conn.Close()
		return
	}

	if s.debug {
		fmt.Println("connect", connect)
		fmt.Println("calling auth plugin")
	}

	authRes, err := s.auth.AuthUser(conn.RemoteAddr().String(), connect.Username, string(connect.Password))

	if err != nil {
		panic(err)
	}

	if s.debug {
		fmt.Println("authentication result", authRes)
	}
	// now we need to call the authentication plugin
	conn.Close()
}
