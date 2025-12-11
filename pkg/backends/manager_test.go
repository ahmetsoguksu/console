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
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.backends == nil {
		t.Error("backends map is nil")
	}
	if manager.configs == nil {
		t.Error("configs map is nil")
	}
}

func TestAddBackend(t *testing.T) {
	manager := NewManager()
	
	config := &BackendConfig{
		ID:       "test-backend",
		Name:     "Test Backend",
		Type:     BackendTypeMinIO,
		Endpoint: "http://localhost:9000",
		Region:   "us-east-1",
		Enabled:  true,
	}
	
	// Create a mock backend
	backend, err := NewMinIOBackend(config)
	if err != nil {
		t.Fatalf("Failed to create MinIO backend: %v", err)
	}
	
	// Add backend
	err = manager.AddBackend(config, backend)
	if err != nil {
		t.Fatalf("Failed to add backend: %v", err)
	}
	
	// Verify backend was added
	retrievedBackend, err := manager.GetBackend("test-backend")
	if err != nil {
		t.Fatalf("Failed to retrieve backend: %v", err)
	}
	if retrievedBackend == nil {
		t.Fatal("Retrieved backend is nil")
	}
	
	// Verify it's the default
	defaultBackend, err := manager.GetDefaultBackend()
	if err != nil {
		t.Fatalf("Failed to get default backend: %v", err)
	}
	if defaultBackend.GetBackendID() != "test-backend" {
		t.Errorf("Expected default backend to be 'test-backend', got '%s'", defaultBackend.GetBackendID())
	}
}

func TestAddBackendDuplicate(t *testing.T) {
	manager := NewManager()
	
	config := &BackendConfig{
		ID:       "duplicate-backend",
		Name:     "Duplicate Backend",
		Type:     BackendTypeMinIO,
		Endpoint: "http://localhost:9000",
		Region:   "us-east-1",
		Enabled:  true,
	}
	
	backend, err := NewMinIOBackend(config)
	if err != nil {
		t.Fatalf("Failed to create MinIO backend: %v", err)
	}
	
	// Add backend first time
	err = manager.AddBackend(config, backend)
	if err != nil {
		t.Fatalf("Failed to add backend first time: %v", err)
	}
	
	// Try to add duplicate
	err = manager.AddBackend(config, backend)
	if err != ErrBackendAlreadyExists {
		t.Errorf("Expected ErrBackendAlreadyExists, got: %v", err)
	}
}

func TestRemoveBackend(t *testing.T) {
	manager := NewManager()
	
	config := &BackendConfig{
		ID:       "remove-backend",
		Name:     "Remove Backend",
		Type:     BackendTypeMinIO,
		Endpoint: "http://localhost:9000",
		Region:   "us-east-1",
		Enabled:  true,
	}
	
	backend, err := NewMinIOBackend(config)
	if err != nil {
		t.Fatalf("Failed to create MinIO backend: %v", err)
	}
	
	// Add backend
	err = manager.AddBackend(config, backend)
	if err != nil {
		t.Fatalf("Failed to add backend: %v", err)
	}
	
	// Remove backend
	err = manager.RemoveBackend("remove-backend")
	if err != nil {
		t.Fatalf("Failed to remove backend: %v", err)
	}
	
	// Verify backend was removed
	_, err = manager.GetBackend("remove-backend")
	if err != ErrBackendNotFound {
		t.Errorf("Expected ErrBackendNotFound, got: %v", err)
	}
}

func TestRemoveNonExistentBackend(t *testing.T) {
	manager := NewManager()
	
	err := manager.RemoveBackend("non-existent")
	if err != ErrBackendNotFound {
		t.Errorf("Expected ErrBackendNotFound, got: %v", err)
	}
}

func TestSetDefaultBackend(t *testing.T) {
	manager := NewManager()
	
	// Add two backends
	config1 := &BackendConfig{
		ID:       "backend1",
		Name:     "Backend 1",
		Type:     BackendTypeMinIO,
		Endpoint: "http://localhost:9001",
		Region:   "us-east-1",
		Enabled:  true,
	}
	
	config2 := &BackendConfig{
		ID:       "backend2",
		Name:     "Backend 2",
		Type:     BackendTypeMinIO,
		Endpoint: "http://localhost:9002",
		Region:   "us-east-1",
		Enabled:  true,
	}
	
	backend1, _ := NewMinIOBackend(config1)
	backend2, _ := NewMinIOBackend(config2)
	
	manager.AddBackend(config1, backend1)
	manager.AddBackend(config2, backend2)
	
	// Set backend2 as default
	err := manager.SetDefaultBackend("backend2")
	if err != nil {
		t.Fatalf("Failed to set default backend: %v", err)
	}
	
	// Verify default backend
	defaultBackend, err := manager.GetDefaultBackend()
	if err != nil {
		t.Fatalf("Failed to get default backend: %v", err)
	}
	if defaultBackend.GetBackendID() != "backend2" {
		t.Errorf("Expected default backend to be 'backend2', got '%s'", defaultBackend.GetBackendID())
	}
}

func TestSetDefaultBackendNonExistent(t *testing.T) {
	manager := NewManager()
	
	err := manager.SetDefaultBackend("non-existent")
	if err != ErrBackendNotFound {
		t.Errorf("Expected ErrBackendNotFound, got: %v", err)
	}
}

func TestListBackends(t *testing.T) {
	manager := NewManager()
	
	// Add multiple backends
	configs := []*BackendConfig{
		{
			ID:       "backend1",
			Name:     "Backend 1",
			Type:     BackendTypeMinIO,
			Endpoint: "http://localhost:9001",
			Region:   "us-east-1",
			Enabled:  true,
		},
		{
			ID:       "backend2",
			Name:     "Backend 2",
			Type:     BackendTypeGarage,
			Endpoint: "http://localhost:3900",
			Region:   "garage",
			Enabled:  true,
		},
	}
	
	for _, config := range configs {
		var backend BackendClient
		var err error
		if config.Type == BackendTypeMinIO {
			backend, err = NewMinIOBackend(config)
		} else {
			config.AccessKey = "test"
			config.SecretKey = "test"
			backend, err = NewGarageBackend(config)
		}
		if err != nil {
			t.Fatalf("Failed to create backend: %v", err)
		}
		manager.AddBackend(config, backend)
	}
	
	// List backends
	list := manager.ListBackends()
	if len(list) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(list))
	}
	
	// Verify sensitive data is not exposed
	for _, config := range list {
		if config.AccessKey != "" {
			t.Error("AccessKey should not be exposed in list")
		}
		if config.SecretKey != "" {
			t.Error("SecretKey should not be exposed in list")
		}
	}
}

func TestGetBackendConfig(t *testing.T) {
	manager := NewManager()
	
	config := &BackendConfig{
		ID:        "config-backend",
		Name:      "Config Backend",
		Type:      BackendTypeMinIO,
		Endpoint:  "http://localhost:9000",
		Region:    "us-east-1",
		AccessKey: "secret-access",
		SecretKey: "secret-key",
		Enabled:   true,
	}
	
	backend, _ := NewMinIOBackend(config)
	manager.AddBackend(config, backend)
	
	// Get config
	retrievedConfig, err := manager.GetBackendConfig("config-backend")
	if err != nil {
		t.Fatalf("Failed to get backend config: %v", err)
	}
	
	// Verify sensitive data is not exposed
	if retrievedConfig.AccessKey != "" {
		t.Error("AccessKey should not be exposed in config")
	}
	if retrievedConfig.SecretKey != "" {
		t.Error("SecretKey should not be exposed in config")
	}
	
	// Verify other fields
	if retrievedConfig.ID != "config-backend" {
		t.Errorf("Expected ID 'config-backend', got '%s'", retrievedConfig.ID)
	}
	if retrievedConfig.Name != "Config Backend" {
		t.Errorf("Expected Name 'Config Backend', got '%s'", retrievedConfig.Name)
	}
}

func TestHealthCheck(t *testing.T) {
	manager := NewManager()
	
	config := &BackendConfig{
		ID:       "health-backend",
		Name:     "Health Backend",
		Type:     BackendTypeMinIO,
		Endpoint: "http://localhost:9000",
		Region:   "us-east-1",
		Enabled:  true,
	}
	
	backend, _ := NewMinIOBackend(config)
	manager.AddBackend(config, backend)
	
	// Perform health check
	ctx := context.Background()
	health := manager.HealthCheck(ctx)
	
	if len(health) != 1 {
		t.Errorf("Expected 1 health status, got %d", len(health))
	}
	
	// Check if health status exists for the backend
	if _, exists := health["health-backend"]; !exists {
		t.Error("Health status not found for 'health-backend'")
	}
}

func TestGetDefaultBackendNoBackends(t *testing.T) {
	manager := NewManager()
	
	_, err := manager.GetDefaultBackend()
	if err != ErrNoBackendsConfigured {
		t.Errorf("Expected ErrNoBackendsConfigured, got: %v", err)
	}
}
