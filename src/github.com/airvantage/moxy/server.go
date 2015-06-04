package moxy

import (
	"bufio"
	"bytes"
	"encoding/hex"
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
	trace       bool
	listenStr   string
	listener    net.Listener
	auth        Authenticator
	filters     []MqttFilter
	filtersDown []MqttFilter
}

// NewServer create a new MQTT proxy server, provide the wanted authenticator
func NewServer(dbg, trace bool, listen string, auth Authenticator, filters []MqttFilter) *Server {
	s := new(Server)
	debug = dbg
	s.trace = trace
	s.listenStr = listen
	s.auth = auth
	s.filters = filters

	// create a second array for filtering in the downstream way
	s.filtersDown = make([]MqttFilter, len(filters))
	for i, v := range filters {
		s.filtersDown[len(filters)-i-1] = v
	}
	return s
}

// Serve bind the socket and serve as a proxy. Will block until end of the world.
func (s *Server) Serve() error {

	if debug {
		fmt.Println("starting the proxy server")
		fmt.Println("with ", len(s.filters), "filters")
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
		// send connack error and close
		conAck := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
		conAck.TopicNameCompression = 0
		conAck.ReturnCode = 4 // bad user/password
		conAck.Write(conn)
		return
	}

	// try to connect & proxy
	if debug {
		fmt.Println("starting proxy to", authRes.Host, authRes.Port)
	}

	err = s.proxy(conn, authRes.Host, authRes.Port, connect, authRes.Metadata)

	if err != nil {
		panic(err)
	}
}

func (s *Server) proxy(con net.Conn, host string, port int, origConnect *packets.ConnectPacket, metadata map[string]interface{}) error {

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
		fmt.Println("connecting to", host, port)
	}
	err = connect.Write(c)
	if err != nil {
		return err
	}
	if debug {
		fmt.Println("waiting conack for", host, port)
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
		fmt.Println("Connack => ", conack)
	}

	// write the connack back to the client
	if err = conack.Write(con); err != nil {
		return err
	}

	if conack.ReturnCode != 0 {
		log.Printf("client connection refused by the upstream broker")
		c.Close()
		return nil
	}

	fmt.Println("MQTT connect success for", host, port)

	// downstream proxify
	go s.proxifyStream(c, con, false, metadata)

	// upstream proxify
	s.proxifyStream(con, c, true, metadata)
	return nil
}

func (s *Server) proxifyStream(reader io.Reader, writer io.Writer, upstream bool, metadata map[string]interface{}) {

	if debug {
		if upstream {
			log.Println("proxify upstream")
		} else {
			log.Println("proxify downstream")
		}
	}
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
		if debug {
			fmt.Println("reading a PDU")
		}
		buff := new(bytes.Buffer)
		header, err := r.ReadByte()

		if debug {
			fmt.Println("got a header byte")
		}
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

		if debug {
			fmt.Println("pumping the PDU")
		}
		_, err = io.CopyN(buff, r, int64(length))
		if eofOrPanic(err) {
			break
		}

		// TODO debug: print and anaylize the PDU

		// filter
		if debug {
			log.Println("filtering the received PDU", hex.Dump(buff.Bytes()))
		}
		var fs []MqttFilter
		if upstream {
			fs = s.filters
		} else {
			fs = s.filtersDown
		}

		bin := walkFilters(buff.Bytes(), fs, upstream, metadata)

		// now push the PDU to the remote connection
		_, err = w.Write(bin)
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

func walkFilters(in []byte, filters []MqttFilter, upstream bool, metadata map[string]interface{}) []byte {
	for _, v := range filters {
		in, _ = v.Filter(in, upstream, metadata)
	}
	return in
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
