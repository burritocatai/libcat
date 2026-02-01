package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/libcat/libcat/internal/models"
)

const (
	appDirName  = ".libcat"
	libraryFile = "library.json"
)

// Storage handles persistence of the book library
type Storage struct {
	path string
}

// NewStorage creates a new storage instance
func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	appDir := filepath.Join(homeDir, appDirName)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create app directory: %w", err)
	}

	return &Storage{
		path: filepath.Join(appDir, libraryFile),
	}, nil
}

// Load loads the library from disk
func (s *Storage) Load() (*models.Library, error) {
	// If file doesn't exist, return empty library
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return &models.Library{
			Books:   []models.Book{},
			Updated: time.Now(),
		}, nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read library file: %w", err)
	}

	var library models.Library
	if err := json.Unmarshal(data, &library); err != nil {
		return nil, fmt.Errorf("failed to parse library file: %w", err)
	}

	return &library, nil
}

// Save saves the library to disk
func (s *Storage) Save(library *models.Library) error {
	library.Updated = time.Now()

	data, err := json.MarshalIndent(library, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize library: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write library file: %w", err)
	}

	return nil
}

// GetPath returns the storage file path
func (s *Storage) GetPath() string {
	return s.path
}
