// Package storage implements local filesystem storage for various interfaces in
// the gameserver package.
package storage

import "path/filepath"

// StorageManager allows for the creation of storage services.
type StorageManager struct {
	path string
}

// NewStorageManager creates a new storage manager.
func NewStorageManager(path string) *StorageManager {
	return &StorageManager{path: filepath.Clean(path)}
}

func (m *StorageManager) pathFor(name string) string {
	return filepath.Join(m.path, "v1", name)
}
