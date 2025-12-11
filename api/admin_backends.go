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

package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/minio/console/models"
	"github.com/minio/console/pkg/backends"
)

// ListBackends returns all configured backends
func ListBackends(ctx context.Context) ([]*backends.BackendConfig, error) {
	return backends.GlobalBackendManager.ListBackends(), nil
}

// GetBackendInfo returns information about a specific backend
func GetBackendInfo(ctx context.Context, backendID string) (*backends.BackendConfig, error) {
	config, err := backends.GlobalBackendManager.GetBackendConfig(backendID)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetBackendHealth returns health status of all backends
func GetBackendHealth(ctx context.Context) (map[string]bool, error) {
	return backends.GlobalBackendManager.HealthCheck(ctx), nil
}

// AddBackend adds a new backend configuration
func AddBackend(ctx context.Context, req *backends.BackendConfig) error {
	// Initialize the backend based on configuration
	var client backends.BackendClient
	var err error

	switch req.Type {
	case backends.BackendTypeMinIO:
		client, err = backends.NewMinIOBackend(req)
	case backends.BackendTypeGarage:
		client, err = backends.NewGarageBackend(req)
	case backends.BackendTypeVersity:
		client, err = backends.NewVersityBackend(req)
	case backends.BackendTypeS3:
		client, err = backends.NewS3Backend(req)
	default:
		return ErrBadRequest
	}

	if err != nil {
		return err
	}

	// Add to manager
	if err := backends.GlobalBackendManager.AddBackend(req, client); err != nil {
		return err
	}

	return nil
}

// RemoveBackend removes a backend from the system
func RemoveBackend(ctx context.Context, backendID string) error {
	return backends.GlobalBackendManager.RemoveBackend(backendID)
}

// SetDefaultBackend sets the default backend
func SetDefaultBackend(ctx context.Context, backendID string) error {
	return backends.GlobalBackendManager.SetDefaultBackend(backendID)
}

// BackendResponse wraps backend configuration for API responses
type BackendResponse struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	Type             backends.BackendType      `json:"type"`
	Endpoint         string                    `json:"endpoint"`
	Region           string                    `json:"region"`
	Secure           bool                      `json:"secure"`
	ClusterMode      bool                      `json:"clusterMode"`
	ClusterEndpoints []string                  `json:"clusterEndpoints,omitempty"`
	Enabled          bool                      `json:"enabled"`
	Metadata         map[string]string         `json:"metadata,omitempty"`
	Healthy          bool                      `json:"healthy"`
}

// BackendsListResponse is the response for listing backends
type BackendsListResponse struct {
	Backends []*BackendResponse `json:"backends"`
	Total    int                `json:"total"`
}

// ConvertBackendToResponse converts a backend config to API response
func ConvertBackendToResponse(config *backends.BackendConfig, healthy bool) *BackendResponse {
	return &BackendResponse{
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
		Healthy:          healthy,
	}
}

// getBackendFromSession retrieves backend ID from session or uses default
func getBackendFromSession(session *models.Principal) (string, error) {
	// For now, always use default backend
	// TODO: Add backend preference to session/token
	backend, err := backends.GlobalBackendManager.GetDefaultBackend()
	if err != nil {
		return "", err
	}
	
	return backend.GetBackendID(), nil
}

// getBackendClient retrieves the backend client for the session
func getBackendClient(session *models.Principal, backendID string) (backends.BackendClient, error) {
	if backendID == "" {
		var err error
		backendID, err = getBackendFromSession(session)
		if err != nil {
			return nil, err
		}
	}
	
	return backends.GlobalBackendManager.GetBackend(backendID)
}

// BackendConfigRequest represents a request to configure a backend
type BackendConfigRequest struct {
	Name             string                    `json:"name"`
	Type             backends.BackendType      `json:"type"`
	Endpoint         string                    `json:"endpoint"`
	Region           string                    `json:"region"`
	AccessKey        string                    `json:"accessKey"`
	SecretKey        string                    `json:"secretKey"`
	ClusterMode      bool                      `json:"clusterMode"`
	ClusterEndpoints []string                  `json:"clusterEndpoints,omitempty"`
	Metadata         map[string]string         `json:"metadata,omitempty"`
}

// ValidateBackendConfig validates backend configuration
func ValidateBackendConfig(req *BackendConfigRequest) error {
	if req.Name == "" {
		return ErrBadRequest
	}
	if req.Endpoint == "" {
		return ErrBadRequest
	}
	if req.Type == "" {
		return ErrBadRequest
	}
	
	// Validate backend type
	switch req.Type {
	case backends.BackendTypeMinIO, backends.BackendTypeGarage, 
	     backends.BackendTypeVersity, backends.BackendTypeS3:
		// Valid types
	default:
		return ErrBadRequest
	}
	
	return nil
}

// ParseBackendConfigRequest parses and validates backend config from request
func ParseBackendConfigRequest(r *http.Request) (*BackendConfigRequest, error) {
	var req BackendConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	
	if err := ValidateBackendConfig(&req); err != nil {
		return nil, err
	}
	
	return &req, nil
}
