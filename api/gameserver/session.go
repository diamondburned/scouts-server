package gameserver

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"libdb.so/hrt"
)

// ErrSessionNotFound is an error that indicates that a session was not found.
var ErrSessionNotFound = hrt.NewHTTPError(401, "session not found")

// SessionToken is a type that represents a session token.
type SessionToken [24]byte

// GenerateSessionToken generates a new session token.
func GenerateSessionToken() SessionToken {
	var token SessionToken
	_, err := rand.Read(token[:])
	if err != nil {
		panic(err)
	}
	return token
}

// String returns the string representation of the session token.
func (t SessionToken) String() string {
	b, _ := t.MarshalText()
	return string(b[:8]) + "..."
}

// MarshalText marshals the session token into text.
func (t SessionToken) MarshalText() ([]byte, error) {
	return []byte(base64.RawStdEncoding.EncodeToString(t[:])), nil
}

// UnmarshalText unmarshals the session token from text.
func (t *SessionToken) UnmarshalText(text []byte) error {
	data, err := base64.RawStdEncoding.DecodeString(string(text))
	if err != nil {
		return err
	}
	copy(t[:], data)
	return nil
}

// SessionMetadata is a struct that contains metadata about a session.
type SessionMetadata struct {
	// UserID is the ID of the user that initiated the session.
	// If this is nil, then the session was initiated by an anonymous user.
	UserID *UserID `json:"user_id,omitempty"`
}

// SessionTTL is the time-to-live of a session.
// Storages don't need to enforce this TTL, but they should be able to handle
// sessions that have exceeded this TTL.
const SessionTTL = 7 * 24 * time.Hour

// SessionStorage is in charge of persisting and retrieving session data from a
// database. A session may or may not be associated with a user.
type SessionStorage interface {
	// CreateSession creates a session with the given session ID and session
	// metadata.
	CreateSession(SessionMetadata) (SessionToken, error)
	// UpdateSession updates the session with the given session ID.
	UpdateSession(SessionToken, SessionMetadata) error
	// QuerySession queries the session with the given session ID.
	QuerySession(SessionToken) (SessionMetadata, error)
}
