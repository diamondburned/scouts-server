package storage

import (
	"time"

	"libdb.so/persist"
	"libdb.so/persist/driver/badgerdb"
	"libdb.so/scouts-server/api/user"
)

type sessionMetadata struct {
	userID user.UserID
	expiry int64
}

// SessionStorage is the session storage service.
type SessionStorage struct {
	m persist.Map[user.SessionToken, sessionMetadata]
}

var _ user.SessionStorage = (*SessionStorage)(nil)

// NewSessionStorage creates a new session storage service.
func NewSessionStorage(manager *StorageManager) (*SessionStorage, error) {
	m, err := persist.NewMap[user.SessionToken, sessionMetadata](
		badgerdb.Open, manager.pathFor("sessions"))
	if err != nil {
		return nil, err
	}
	return &SessionStorage{m: m}, nil
}

func (s *SessionStorage) CreateSession(userID user.UserID) (user.SessionToken, error) {
	for {
		token := user.GenerateSessionToken()
		_, exists, err := s.m.LoadOrStore(token, sessionMetadata{
			userID: userID,
			expiry: time.Now().Add(user.SessionTTL).Unix(),
		})
		if err != nil {
			return user.SessionToken{}, err
		}
		if exists {
			continue
		}
		return token, nil
	}
}

func (s *SessionStorage) QuerySession(token user.SessionToken) (user.UserID, error) {
	value, ok, err := s.m.Load(token)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, user.ErrSessionNotFound
	}
	if value.expiry < time.Now().Unix() {
		return 0, user.ErrSessionNotFound
	}
	return value.userID, nil
}
