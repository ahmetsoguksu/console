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
	"fmt"
	"strings"

	xnet "github.com/minio/pkg/v3/net"
)

// VersityBackend implements BackendClient for Versity ScoutFS Gateway
// Versity Gateway provides S3-compatible interface to various storage backends
type VersityBackend struct {
	*MinIOBackend // Versity is S3-compatible, so we can reuse MinIO implementation
}

// NewVersityBackend creates a new Versity backend client
func NewVersityBackend(config *BackendConfig) (*VersityBackend, error) {
	// Parse endpoint to determine if secure
	_, err := xnet.ParseHTTPURL(config.Endpoint)
	if err != nil {
		return nil, err
	}

	// Validate credentials
	if config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("versity backend requires access key and secret key")
	}

	// Create underlying MinIO backend (since Versity is S3-compatible)
	minioBackend, err := NewMinIOBackend(config)
	if err != nil {
		return nil, err
	}

	return &VersityBackend{
		MinIOBackend: minioBackend,
	}, nil
}

// Override GetBackendType to return Versity type
func (v *VersityBackend) GetBackendType() BackendType {
	return BackendTypeVersity
}

// Override IsHealthy with Versity-specific health check if needed
func (v *VersityBackend) IsHealthy(ctx context.Context) bool {
	// Try to list buckets as a health check
	_, err := v.client.ListBuckets(ctx)
	return err == nil || !strings.Contains(err.Error(), "connection")
}
