package moxy

import (
	"bufio"
	"bytes"
	"fmt"
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git/packets"
	"io"
	"log"
	"net"
	"strconv"
)

var debug bool = false

// A MQTT proxy server
type Server struct {
	trace     bool
	listenStr string
	listener  net.Listener
	auth      Authenticator
}

// NewServer create a new MQTT proxy server, provide the wanted authenticator
func NewServer(dbg, trace bool, listen string, auth Authenticator) *Server {
	s := new(Server)
	debug = dbg
	s.trace = trace
	s.listenStr = listen
	s.auth = auth
	return s
}

// Serve bind the socket and serve as a proxy. Will block until end of the world.
func (s *Server) Serve() error {

	if debug {
		fmt.Println("starting the proxy server")
	}

	var err error
	s.listener, err = net.Listen("tcp", s.listenStr)
	if err != nil {
		return err
	}

	// endless accept loop
	for {
		if debug {
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
	if debug {
		fmt.Printf("new connection: %v\n", conn.RemoteAddr())
	}
	defer conn.Close()

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

	if debug {
		fmt.Println("connect", connect)
		fmt.Println("calling auth plugin")
	}

	authRes, err := s.auth.AuthUser(conn.RemoteAddr().String(), connect.Username, string(connect.Password))

	if err != nil {
		panic(err)
	}

	if debug {
		fmt.Println("authentication result", authRes)
	}
	if !authRes.Success {
		conAck := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
		conAck.TopicNameCompression = 0
		conAck.ReturnCode = 4 // bad user/password
		conAck.Write(conn)
		return
	}
	conAck := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
	conAck.TopicNameCompression = 0
	conAck.ReturnCode = 0 // connection accepted!
	conAck.Write(conn)

	// try to connect & proxy
	if debug {
		fmt.Println("starting proxy to", authRes.Host, authRes.Port)
	}

	err = proxy(conn, authRes.Host, authRes.Port, connect)

	if err != nil {
		panic(err)
	}
}

func proxy(con net.Conn, host string, port int, origConnect *packets.ConnectPacket) error {

	c, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	if debug {
		fmt.Println("connected")
	}
	// write a connect and wait a connack

	connect := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)

	connect.ClientIdentifier = origConnect.ClientIdentifier
	connect.WillQos = origConnect.WillQos
	connect.WillFlag = origConnect.WillFlag
	connect.WillTopic = origConnect.WillTopic
	connect.UsernameFlag = false
	connect.Dup = origConnect.Dup
	connect.ProtocolVersion = origConnect.ProtocolVersion
	connect.Retain = origConnect.Retain
	connect.CleanSession = origConnect.CleanSession
	connect.ProtocolName = origConnect.ProtocolName
	connect.Qos = origConnect.Qos

	if debug {
		fmt.Println("sending connect")
	}
	err = connect.Write(c)
	if err != nil {
		return err
	}
	if debug {
		fmt.Println("waiting conack")
	}

	// now response!
	cp, err := packets.ReadPacket(c)
	if err != nil {
		return err
	}

	conack, ok := cp.(*packets.ConnackPacket)
	if !ok {
		log.Printf("we want a Conack packet %v", cp)
		c.Close()
		return nil
	}
	if debug {
		fmt.Println("Connack", conack)
	}

	if conack.ReturnCode != 0 {
		log.Printf("client connection refused by the broker")
		c.Close()
		return nil
	}
	// connection received !
	if debug {
		fmt.Println("Conack", conack)
	}

	go proxifyStream(c, con)
	proxifyStream(con, c)
	return nil
}

func proxifyStream(reader io.Reader, writer io.Writer) {

	defer func() {
		if r := recover(); r != nil {
			if debug {
				fmt.Println("Recovered in f", r)
			}
		}
	}()

	r := bufio.NewReader(reader)
	w := bufio.NewWriter(writer)
	for {
		// read a whole MQTT PDU
		buff := new(bytes.Buffer)
		header, err := r.ReadByte()

		if eofOrPanic(err) {
			break
		}

		buff.WriteByte(header)

		// read variable length header
		multiplier := 1
		length := 0

		for {
			b, err := r.ReadByte()
			if eofOrPanic(err) {
				break
			}

			buff.WriteByte(b)

			length += (int(b) & 127) * multiplier
			multiplier *= 128
			if b&128 == 0 {
				break
			}
		}

		// now consume remaining length bytes
		_, err = io.CopyN(buff, r, int64(length))
		if eofOrPanic(err) {
			break
		}

		// TODO debug: print and anaylize the PDU

		// now push the PDU to the remote connection

		_, err = buff.WriteTo(w)
		if eofOrPanic(err) {
			break
		}

		if err != nil {
			panic(err)
		}

		err = w.Flush()
		if eofOrPanic(err) {
			break
		}
	}

	if debug {
		fmt.Println("EoF")
	}
}

func eofOrPanic(err error) bool {
	if err == nil {
		return false
	}

	if err == io.EOF {
		return true
	}

	panic(err)
}
