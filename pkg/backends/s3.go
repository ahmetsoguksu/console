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

// S3Backend implements BackendClient for generic S3-compatible services
// This can be used for AWS S3, Wasabi, DigitalOcean Spaces, etc.
type S3Backend struct {
	*MinIOBackend // S3 is compatible, so we can reuse MinIO implementation
}

// NewS3Backend creates a new generic S3 backend client
func NewS3Backend(config *BackendConfig) (*S3Backend, error) {
	// Parse endpoint to determine if secure
	_, err := xnet.ParseHTTPURL(config.Endpoint)
	if err != nil {
		return nil, err
	}

	// Validate credentials
	if config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("s3 backend requires access key and secret key")
	}

	// Create underlying MinIO backend (since S3 is compatible)
	minioBackend, err := NewMinIOBackend(config)
	if err != nil {
		return nil, err
	}

	return &S3Backend{
		MinIOBackend: minioBackend,
	}, nil
}

// Override GetBackendType to return S3 type
func (s *S3Backend) GetBackendType() BackendType {
	return BackendTypeS3
}

// Override IsHealthy with S3-specific health check if needed
func (s *S3Backend) IsHealthy(ctx context.Context) bool {
	// Try to list buckets as a health check
	_, err := s.client.ListBuckets(ctx)
	return err == nil || !strings.Contains(err.Error(), "connection")
}
