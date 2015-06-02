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
	Metadata     map[string]interface{}
}
