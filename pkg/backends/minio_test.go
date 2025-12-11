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

func TestNewMinIOBackend(t *testing.T) {
	config := &BackendConfig{
		ID:       "test-minio",
		Name:     "Test MinIO",
		Type:     BackendTypeMinIO,
		Endpoint: "http://localhost:9000",
		Region:   "us-east-1",
		Enabled:  true,
	}
	
	backend, err := NewMinIOBackend(config)
	if err != nil {
		t.Fatalf("Failed to create MinIO backend: %v", err)
	}
	
	if backend == nil {
		t.Fatal("Backend is nil")
	}
	
	if backend.GetBackendType() != BackendTypeMinIO {
		t.Errorf("Expected type MinIO, got %s", backend.GetBackendType())
	}
	
	if backend.GetBackendID() != "test-minio" {
		t.Errorf("Expected ID 'test-minio', got '%s'", backend.GetBackendID())
	}
}

func TestNewMinIOBackendHTTPS(t *testing.T) {
	config := &BackendConfig{
		ID:       "test-minio-https",
		Name:     "Test MinIO HTTPS",
		Type:     BackendTypeMinIO,
		Endpoint: "https://minio.example.com:9000",
		Region:   "us-west-1",
		Enabled:  true,
	}
	
	backend, err := NewMinIOBackend(config)
	if err != nil {
		t.Fatalf("Failed to create MinIO backend: %v", err)
	}
	
	if backend == nil {
		t.Fatal("Backend is nil")
	}
}

func TestNewMinIOBackendWithCredentials(t *testing.T) {
	config := &BackendConfig{
		ID:        "test-minio-creds",
		Name:      "Test MinIO with Creds",
		Type:      BackendTypeMinIO,
		Endpoint:  "http://localhost:9000",
		Region:    "us-east-1",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Enabled:   true,
	}
	
	backend, err := NewMinIOBackend(config)
	if err != nil {
		t.Fatalf("Failed to create MinIO backend: %v", err)
	}
	
	if backend == nil {
		t.Fatal("Backend is nil")
	}
}

func TestNewMinIOBackendInvalidEndpoint(t *testing.T) {
	config := &BackendConfig{
		ID:       "test-minio-invalid",
		Name:     "Test MinIO Invalid",
		Type:     BackendTypeMinIO,
		Endpoint: "not-a-valid-url",
		Region:   "us-east-1",
		Enabled:  true,
	}
	
	_, err := NewMinIOBackend(config)
	if err == nil {
		t.Error("Expected error for invalid endpoint, got nil")
	}
}
