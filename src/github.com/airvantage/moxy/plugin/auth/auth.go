// Package for implementing a encoding/gob unix socket based authentication plugin
package auth

import (
	"encoding/gob"
	"github.com/airvantage/moxy"
	"github.com/airvantage/moxy/plugin"
)

// A plugin for authentication, implements the Authenticator interface
type AuthPlugin struct {
	*plugin.Plugin
}

// NewAuthPlugin create a new authentication plugin using the provided system command
func NewAuthPlugin(name string) *AuthPlugin {

	var res AuthPlugin
	res.Plugin = plugin.NewPlugin(name, "auth.sock")
	return &res
}

// implements the Authenticator interface by communication with the plugin command
func (ap *AuthPlugin) AuthUser(connection, user, password string) (moxy.AuthResult, error) {

	var call struct {
		Password string
		UserName string
	}

	call.Password = password
	call.UserName = user

	c, err := ap.Dial()

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
		Metadata     map[string]interface{}
	}
	err = dec.Decode(&result)
	if err != nil {
		panic(err)
	}
	return result, nil
}
