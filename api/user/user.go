package user

import (
	"fmt"
	"strconv"
	"time"

	"github.com/godruoyi/go-snowflake"
	"github.com/oklog/ulid/v2"
)

// UserID is a type that represents a user UserID.
type UserID uint64

// GenerateUserID generates a new user ID.
func GenerateUserID() UserID {
	return UserID(snowflake.ID())
}

// CreatedAt returns the time that the user ID was generated.
func (u UserID) CreatedAt() time.Time {
	return ulid.Time(snowflake.ParseID(uint64(u)).Timestamp)
}

// String returns the string representation of the user ID.
func (u UserID) String() string {
	return strconv.FormatUint(uint64(u), 16)
}

// MarshalText marshals the user ID into text.
func (u UserID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText unmarshals the user ID from text.
func (u *UserID) UnmarshalText(text []byte) error {
	id, err := strconv.ParseUint(string(text), 16, 64)
	if err != nil {
		return err
	}
	*u = UserID(id)
	return nil
}

// Authorized is a struct that contains the session token and the user ID
// for an authorized user.
type Authorized struct {
	// User is the ID of the user.
	// If this is nil, then the user is anonymous.
	User    *UserID `json:"user_id,omitempty"`
	session SessionToken
}

// NewAuthorized creates a new authorized user.
func NewAuthorized(session SessionToken, user UserID) Authorized {
	return Authorized{
		User:    &user,
		session: session,
	}
}

// NewAnonymous creates a new anonymous authorized user.
func NewAnonymous(session SessionToken) Authorized {
	return Authorized{
		session: session,
	}
}

// OptionalAuthorizedUserString returns the string representation of the
// authorized user or "<nil>" if the authorized user is nil.
func OptionalAuthorizedUserString(user *Authorized) string {
	if user == nil {
		return "<nil>"
	}
	return user.String()
}

// Session returns the session token.
func (u Authorized) Session() SessionToken {
	return u.session
}

// String returns the string representation of the authorized user.
// The token is truncated to 8 characters.
func (u Authorized) String() string {
	str := u.session.String()
	if u.User != nil {
		str += fmt.Sprintf("[%s]", *u.User)
	} else {
		str += "[?]"
	}
	return str
}

// Eq returns true if the authorized user is equal to the other authorized
// user. It only compares the token.
func (u Authorized) Eq(other Authorized) bool {
	return u.session == other.session
}
