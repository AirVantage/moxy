// Package for implementing a encoding/gob unix socket based authentication plugin
package auth

import (
	"encoding/gob"
	"github.com/airvantage/moxy"
	"net"
	"os/exec"
)

// A plugin for authentication, implements the Authenticator interface
type AuthPlugin struct {
	cmd *exec.Cmd
}

// group the argument for the plugin RPC call
type AuthCall struct {
	UserName string
	Password string
}

// NewAuthPlugin create a new authentication plugin using the provided system command
func NewAuthPlugin(name string) *AuthPlugin {
	res := new(AuthPlugin)

	res.cmd = exec.Command(name, "/tmp/auth.sock")

	if err := res.cmd.Start(); err != nil {
		panic(err)
	}

	// connect to the socket
	return res
}

// implements the Authenticator interface by communication with the plugin command
func (ap *AuthPlugin) AuthUser(connection, user, password string) (moxy.AuthResult, error) {
	var call AuthCall

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

	var result moxy.AuthResult
	err = dec.Decode(&result)
	if err != nil {
		return moxy.AuthResult{}, err
	}
	return result, nil
}
