package moxy

// Authenticator authenticate a client and return the broker to contact
type Authenticator interface {
	AuthUser(connection, user, password string) (AuthResult, error)
}

type AuthResult struct {
	Success      bool
	ErrorMessage string
	Host         string
	Port         int
	// freestyle metadata which will be passed to other plugins (like filters)
	Metadata map[string]interface{}
	// list of topic to be subscribed (with the given QoS level)
	Topics map[string]uint
}
