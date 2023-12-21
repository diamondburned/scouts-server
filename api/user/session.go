package user

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
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
	return hex.EncodeToString(t[:])[:8]
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

// SessionTTL is the time-to-live of a session.
// Storages don't need to enforce this TTL, but they should be able to handle
// sessions that have exceeded this TTL.
const SessionTTL = 7 * 24 * time.Hour

// SessionStorage is in charge of persisting and retrieving session data from a
// database. A session may or may not be associated with a user.
type SessionStorage interface {
	// CreateSession creates a session with the given session ID.
	CreateSession() (SessionToken, error)
	// ChangeSession changes the user that the session is associated with.
	// This is useful when a user logs in or logs out.
	ChangeSession(SessionToken, *UserID) error
	// QuerySession queries the session with the given session ID.
	QuerySession(SessionToken) (*UserID, error)
}

// CachedSessionStorage is a session storage that caches session data.
// Sessions wrapped by this storage are cached for about 5 minutes.
type CachedSessionStorage struct {
	storage SessionStorage
	cache   xsync.MapOf[SessionToken, cachedSession]
}

var _ SessionStorage = (*CachedSessionStorage)(nil)

// SessionCacheTTL is the time-to-live of a session cache entry.
// This should be less than SessionTTL.
const SessionCacheTTL = 5 * time.Minute

type cachedSession struct {
	userID *UserID
	expiry time.Time
}

// NewCachedSessionStorage creates a new cached session storage.
func NewCachedSessionStorage(storage SessionStorage) *CachedSessionStorage {
	if cached, ok := storage.(*CachedSessionStorage); ok {
		return cached
	}
	return &CachedSessionStorage{
		storage: storage,
		cache:   *xsync.NewMapOf[SessionToken, cachedSession](),
	}
}

// CreateSession creates a new anonymous session.
func (s *CachedSessionStorage) CreateSession() (SessionToken, error) {
	token, err := s.storage.CreateSession()
	if err != nil {
		return token, err
	}
	s.cache.Store(token, cachedSession{
		userID: nil,
		expiry: time.Now().Add(SessionCacheTTL),
	})
	return token, nil
}

// ChangeSession changes the user that the session is associated with.
func (s *CachedSessionStorage) ChangeSession(token SessionToken, userID *UserID) error {
	err := s.storage.ChangeSession(token, userID)
	if err != nil {
		return err
	}
	s.cache.Store(token, cachedSession{
		userID: userID,
		expiry: time.Now().Add(SessionCacheTTL),
	})
	return nil
}

// QuerySession queries the session with the given session ID.
func (s *CachedSessionStorage) QuerySession(token SessionToken) (*UserID, error) {
	if session, ok := s.cache.Load(token); ok {
		if session.expiry.After(time.Now()) {
			return session.userID, nil
		}
		s.cache.Delete(token)
	}
	userID, err := s.storage.QuerySession(token)
	if err != nil {
		return userID, err
	}
	s.cache.Store(token, cachedSession{
		userID: userID,
		expiry: time.Now().Add(SessionCacheTTL),
	})
	return userID, nil
}
