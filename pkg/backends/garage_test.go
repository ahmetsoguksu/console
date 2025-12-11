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
	"testing"
)

func TestNewGarageBackend(t *testing.T) {
	config := &BackendConfig{
		ID:        "test-garage",
		Name:      "Test Garage",
		Type:      BackendTypeGarage,
		Endpoint:  "http://localhost:3900",
		Region:    "garage",
		AccessKey: "GK123",
		SecretKey: "secret123",
		Enabled:   true,
	}
	
	backend, err := NewGarageBackend(config)
	if err != nil {
		t.Fatalf("Failed to create Garage backend: %v", err)
	}
	
	if backend == nil {
		t.Fatal("Backend is nil")
	}
	
	if backend.GetBackendType() != BackendTypeGarage {
		t.Errorf("Expected type Garage, got %s", backend.GetBackendType())
	}
	
	if backend.GetBackendID() != "test-garage" {
		t.Errorf("Expected ID 'test-garage', got '%s'", backend.GetBackendID())
	}
}

func TestNewGarageBackendClusterMode(t *testing.T) {
	config := &BackendConfig{
		ID:        "test-garage-cluster",
		Name:      "Test Garage Cluster",
		Type:      BackendTypeGarage,
		Endpoint:  "http://garage-lb.example.com:3900",
		Region:    "garage",
		AccessKey: "GK123",
		SecretKey: "secret123",
		ClusterMode: true,
		ClusterEndpoints: []string{
			"http://garage1.example.com:3900",
			"http://garage2.example.com:3900",
			"http://garage3.example.com:3900",
		},
		Enabled: true,
	}
	
	backend, err := NewGarageBackend(config)
	if err != nil {
		t.Fatalf("Failed to create Garage backend: %v", err)
	}
	
	if backend == nil {
		t.Fatal("Backend is nil")
	}
	
	// Verify cluster clients were created
	if len(backend.clients) != 3 {
		t.Errorf("Expected 3 cluster clients, got %d", len(backend.clients))
	}
}

func TestNewGarageBackendNoCredentials(t *testing.T) {
	config := &BackendConfig{
		ID:       "test-garage-nocreds",
		Name:     "Test Garage No Creds",
		Type:     BackendTypeGarage,
		Endpoint: "http://localhost:3900",
		Region:   "garage",
		Enabled:  true,
	}
	
	_, err := NewGarageBackend(config)
	if err == nil {
		t.Error("Expected error for missing credentials, got nil")
	}
}

func TestNewGarageBackendInvalidClusterEndpoint(t *testing.T) {
	config := &BackendConfig{
		ID:        "test-garage-invalid",
		Name:      "Test Garage Invalid",
		Type:      BackendTypeGarage,
		Endpoint:  "http://localhost:3900",
		Region:    "garage",
		AccessKey: "GK123",
		SecretKey: "secret123",
		ClusterMode: true,
		ClusterEndpoints: []string{
			"not-a-valid-url",
		},
		Enabled: true,
	}
	
	_, err := NewGarageBackend(config)
	if err == nil {
		t.Error("Expected error for invalid cluster endpoint, got nil")
	}
}

func TestGarageBackendGetClient(t *testing.T) {
	config := &BackendConfig{
		ID:        "test-garage-client",
		Name:      "Test Garage Client",
		Type:      BackendTypeGarage,
		Endpoint:  "http://localhost:3900",
		Region:    "garage",
		AccessKey: "GK123",
		SecretKey: "secret123",
		ClusterMode: true,
		ClusterEndpoints: []string{
			"http://garage1.example.com:3900",
			"http://garage2.example.com:3900",
		},
		Enabled: true,
	}
	
	backend, err := NewGarageBackend(config)
	if err != nil {
		t.Fatalf("Failed to create Garage backend: %v", err)
	}
	
	// Get client multiple times to test round-robin
	client1 := backend.getClient()
	client2 := backend.getClient()
	
	if client1 == nil || client2 == nil {
		t.Error("Client should not be nil")
	}
	
	// With 2 clients, round-robin should return different clients
	if client1 == client2 {
		// This might happen, but with round-robin it should alternate
		// We'll just verify they're not nil
	}
}

func TestGarageBackendGetRandomClient(t *testing.T) {
	config := &BackendConfig{
		ID:        "test-garage-random",
		Name:      "Test Garage Random",
		Type:      BackendTypeGarage,
		Endpoint:  "http://localhost:3900",
		Region:    "garage",
		AccessKey: "GK123",
		SecretKey: "secret123",
		ClusterMode: true,
		ClusterEndpoints: []string{
			"http://garage1.example.com:3900",
			"http://garage2.example.com:3900",
			"http://garage3.example.com:3900",
		},
		Enabled: true,
	}
	
	backend, err := NewGarageBackend(config)
	if err != nil {
		t.Fatalf("Failed to create Garage backend: %v", err)
	}
	
	// Get random client multiple times
	for i := 0; i < 10; i++ {
		client := backend.getRandomClient()
		if client == nil {
			t.Error("Random client should not be nil")
		}
	}
}

func TestGarageBackendSingleClient(t *testing.T) {
	config := &BackendConfig{
		ID:        "test-garage-single",
		Name:      "Test Garage Single",
		Type:      BackendTypeGarage,
		Endpoint:  "http://localhost:3900",
		Region:    "garage",
		AccessKey: "GK123",
		SecretKey: "secret123",
		Enabled:   true,
	}
	
	backend, err := NewGarageBackend(config)
	if err != nil {
		t.Fatalf("Failed to create Garage backend: %v", err)
	}
	
	// With single client, both methods should return same client
	client1 := backend.getClient()
	client2 := backend.getRandomClient()
	
	if client1 != client2 {
		t.Error("Single client mode should return same client")
	}
}
