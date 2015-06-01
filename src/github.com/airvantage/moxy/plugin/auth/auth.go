// Package for implementing a encoding/gob unix socket based authentication plugin
package auth

import (
	"bufio"
	"encoding/gob"
	"github.com/airvantage/moxy"
	"io"
	"net"
	"os"
	"os/exec"
)

// A plugin for authentication, implements the Authenticator interface
type AuthPlugin struct {
	cmd *exec.Cmd
}

// NewAuthPlugin create a new authentication plugin using the provided system command
func NewAuthPlugin(name string) *AuthPlugin {
	res := new(AuthPlugin)

	res.cmd = exec.Command(name, "/tmp/auth.sock")

	// redirect stdout & stderr
	stdout, err := res.cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := res.cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	go fwd(stdout, os.Stdout)
	go fwd(stderr, os.Stderr)

	if err := res.cmd.Start(); err != nil {
		panic(err)
	}

	// connect to the socket
	return res
}

func fwd(in io.Reader, out io.Writer) {
	r := bufio.NewReader(in)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				out.Write([]byte("AUTH> ERROR READING " + err.Error()))
			}
			break
		}
		out.Write([]byte("AUTH> " + string(line)))
	}
}

// implements the Authenticator interface by communication with the plugin command
func (ap *AuthPlugin) AuthUser(connection, user, password string) (moxy.AuthResult, error) {

	var call struct {
		Password string
		UserName string
	}

	call.Password = password
	call.UserName = user

	c, err := net.Dial("unix", "/tmp/auth.sock")

	if err != nil {
		panic(err)
	}

	defer c.Close()

	enc := gob.NewEncoder(c)
	enc.Encode(call)

	dec := gob.NewDecoder(c)

	var result struct {
		Success      bool
		ErrorMessage string
		Host         string
		Port         int
	}
	err = dec.Decode(&result)
	if err != nil {
		return moxy.AuthResult{}, err
	}
	return result, nil
}
