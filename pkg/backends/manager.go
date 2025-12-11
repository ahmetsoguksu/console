// This file is part of MinIO Console Server
// Copyright (c) 2021 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package backends

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/minio/pkg/v3/env"
)

var (
	// ErrBackendNotFound is returned when a backend is not found
	ErrBackendNotFound = errors.New("backend not found")
	// ErrBackendAlreadyExists is returned when trying to add a duplicate backend
	ErrBackendAlreadyExists = errors.New("backend already exists")
	// ErrNoBackendsConfigured is returned when no backends are configured
	ErrNoBackendsConfigured = errors.New("no backends configured")
)

// Manager manages multiple S3 backends
type Manager struct {
	backends map[string]BackendClient
	configs  map[string]*BackendConfig
	mu       sync.RWMutex
	default  string // ID of default backend
}

// GlobalBackendManager is the global instance of backend manager
var GlobalBackendManager *Manager

func init() {
	GlobalBackendManager = NewManager()
}

// NewManager creates a new backend manager
func NewManager() *Manager {
	return &Manager{
		backends: make(map[string]BackendClient),
		configs:  make(map[string]*BackendConfig),
	}
}

// AddBackend adds a new backend to the manager
func (m *Manager) AddBackend(config *BackendConfig, client BackendClient) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.backends[config.ID]; exists {
		return ErrBackendAlreadyExists
	}

	m.backends[config.ID] = client
	m.configs[config.ID] = config

	// Set as default if it's the first backend
	if m.default == "" {
		m.default = config.ID
	}

	return nil
}

// RemoveBackend removes a backend from the manager
func (m *Manager) RemoveBackend(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.backends[id]; !exists {
		return ErrBackendNotFound
	}

	delete(m.backends, id)
	delete(m.configs, id)

	// If removing default, set new default
	if m.default == id && len(m.backends) > 0 {
		for backendID := range m.backends {
			m.default = backendID
			break
		}
	}

	return nil
}

// GetBackend retrieves a backend by ID
func (m *Manager) GetBackend(id string) (BackendClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	backend, exists := m.backends[id]
	if !exists {
		return nil, ErrBackendNotFound
	}

	return backend, nil
}

// GetDefaultBackend returns the default backend
func (m *Manager) GetDefaultBackend() (BackendClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.default == "" {
		return nil, ErrNoBackendsConfigured
	}

	backend, exists := m.backends[m.default]
	if !exists {
		return nil, ErrBackendNotFound
	}

	return backend, nil
}

// SetDefaultBackend sets the default backend
func (m *Manager) SetDefaultBackend(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.backends[id]; !exists {
		return ErrBackendNotFound
	}

	m.default = id
	return nil
}

// ListBackends returns all backend configurations
func (m *Manager) ListBackends() []*BackendConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configs := make([]*BackendConfig, 0, len(m.configs))
	for _, config := range m.configs {
		// Create a copy without sensitive data
		safeCopy := &BackendConfig{
			ID:               config.ID,
			Name:             config.Name,
			Type:             config.Type,
			Endpoint:         config.Endpoint,
			Region:           config.Region,
			Secure:           config.Secure,
			ClusterMode:      config.ClusterMode,
			ClusterEndpoints: config.ClusterEndpoints,
			Enabled:          config.Enabled,
			Metadata:         config.Metadata,
		}
		configs = append(configs, safeCopy)
	}

	return configs
}

// GetBackendConfig returns the configuration for a specific backend
func (m *Manager) GetBackendConfig(id string) (*BackendConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[id]
	if !exists {
		return nil, ErrBackendNotFound
	}

	// Return a copy without sensitive data
	safeCopy := &BackendConfig{
		ID:               config.ID,
		Name:             config.Name,
		Type:             config.Type,
		Endpoint:         config.Endpoint,
		Region:           config.Region,
		Secure:           config.Secure,
		ClusterMode:      config.ClusterMode,
		ClusterEndpoints: config.ClusterEndpoints,
		Enabled:          config.Enabled,
		Metadata:         config.Metadata,
	}

	return safeCopy, nil
}

// HealthCheck checks the health of all backends
func (m *Manager) HealthCheck(ctx context.Context) map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[string]bool)
	for id, backend := range m.backends {
		health[id] = backend.IsHealthy(ctx)
	}

	return health
}

// LoadFromEnv loads backend configurations from environment variables
// Environment variable format: CONSOLE_BACKENDS='[{"id":"backend1","type":"minio",...}]'
func (m *Manager) LoadFromEnv() error {
	backendsJSON := env.Get("CONSOLE_BACKENDS", "")
	if backendsJSON == "" {
		// For backward compatibility, try to load single backend from legacy env vars
		return m.loadLegacyConfig()
	}

	var configs []BackendConfig
	if err := json.Unmarshal([]byte(backendsJSON), &configs); err != nil {
		return fmt.Errorf("failed to parse CONSOLE_BACKENDS: %w", err)
	}

	for _, config := range configs {
		if err := m.initBackend(&config); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize backend %s: %v\n", config.ID, err)
			continue
		}
	}

	if len(m.backends) == 0 {
		return ErrNoBackendsConfigured
	}

	return nil
}

// loadLegacyConfig loads configuration from legacy environment variables (CONSOLE_MINIO_SERVER, etc.)
func (m *Manager) loadLegacyConfig() error {
	endpoint := env.Get("CONSOLE_MINIO_SERVER", "")
	if endpoint == "" {
		return ErrNoBackendsConfigured
	}

	// Create a default backend from legacy config
	config := &BackendConfig{
		ID:       "default",
		Name:     "Default MinIO",
		Type:     BackendTypeMinIO,
		Endpoint: endpoint,
		Region:   env.Get("CONSOLE_MINIO_REGION", ""),
		Secure:   true, // Will be determined from endpoint
		Enabled:  true,
	}

	return m.initBackend(config)
}

// initBackend initializes a backend based on its configuration
func (m *Manager) initBackend(config *BackendConfig) error {
	var client BackendClient
	var err error

	switch config.Type {
	case BackendTypeMinIO:
		client, err = NewMinIOBackend(config)
	case BackendTypeGarage:
		client, err = NewGarageBackend(config)
	case BackendTypeVersity:
		client, err = NewVersityBackend(config)
	case BackendTypeS3:
		client, err = NewS3Backend(config)
	default:
		return fmt.Errorf("unsupported backend type: %s", config.Type)
	}

	if err != nil {
		return err
	}

	return m.AddBackend(config, client)
}
