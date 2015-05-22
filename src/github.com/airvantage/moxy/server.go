package moxy

import (
	"fmt"
	"net"
)

type Server struct {
	debug     bool
	trace     bool
	listenStr string
	listener  net.Listener
}

func NewServer(debug, trace bool, listen string) *Server {
	s := new(Server)
	s.debug = debug
	s.trace = trace
	s.listenStr = listen
	return s
}

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
	conn.Close()
}
