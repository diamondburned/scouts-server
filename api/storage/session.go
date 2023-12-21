package storage

import (
	"time"

	"libdb.so/persist"
	"libdb.so/persist/driver/badgerdb"
	"libdb.so/scouts-server/api/user"
)

type sessionMetadata struct {
	userID *user.UserID
	expiry int64
}

// SessionStorage is the session storage service.
type SessionStorage struct {
	m persist.Map[user.SessionToken, sessionMetadata]
}

var _ user.SessionStorage = (*SessionStorage)(nil)

func newSessionStorage(manager *StorageManager) (*SessionStorage, error) {
	path, err := manager.pathFor("sessions")
	if err != nil {
		return nil, err
	}

	m, err := persist.NewMap[user.SessionToken, sessionMetadata](badgerdb.Open, path)
	if err != nil {
		return nil, err
	}

	return &SessionStorage{m: m}, nil
}

func (s *SessionStorage) CreateSession() (user.SessionToken, error) {
	for {
		token := user.GenerateSessionToken()
		_, exists, err := s.m.LoadOrStore(token, sessionMetadata{
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

func (s *SessionStorage) ChangeSession(token user.SessionToken, userID *user.UserID) error {
	value, ok, err := s.m.Load(token)
	if err != nil {
		return err
	}
	if !ok {
		return user.ErrSessionNotFound
	}
	value.userID = userID
	return s.m.Store(token, value)
}

func (s *SessionStorage) QuerySession(token user.SessionToken) (*user.UserID, error) {
	value, ok, err := s.m.Load(token)
	if err != nil {
		return nil, err
	}
	if !ok || value.expiry < time.Now().Unix() {
		return nil, user.ErrSessionNotFound
	}
	return value.userID, nil
}
