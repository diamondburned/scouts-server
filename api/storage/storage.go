// Package storage implements local filesystem storage for various interfaces in
// the gameserver package.
package storage

import (
	"os"
	"path/filepath"
)

// StorageManager allows for the creation of storage services.
type StorageManager struct {
	path string
}

// NewStorageManager creates a new storage manager.
func NewStorageManager(path string) *StorageManager {
	return &StorageManager{path: filepath.Clean(path)}
}

func (m *StorageManager) pathFor(name string) (string, error) {
	path := filepath.Join(m.path, "v1", name)
	return path, os.MkdirAll(path, 0700)
}

// OpenSessionStorage opens a session storage service.
func (m *StorageManager) OpenSessionStorage() (*SessionStorage, error) {
	return newSessionStorage(m)
}
