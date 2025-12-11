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
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/minio/minio-go/v7/pkg/notification"
	"github.com/minio/minio-go/v7/pkg/replication"
	"github.com/minio/minio-go/v7/pkg/sse"
	"github.com/minio/minio-go/v7/pkg/tags"
	xnet "github.com/minio/pkg/v3/net"
)

// GarageBackend implements BackendClient for Garage S3
// Garage is a lightweight S3-compatible object storage system designed for self-hosting
type GarageBackend struct {
	clients []*minio.Client // Multiple clients for cluster mode
	config  *BackendConfig
	mu      sync.RWMutex
	current int // Current client index for round-robin
}

// NewGarageBackend creates a new Garage backend client
func NewGarageBackend(config *BackendConfig) (*GarageBackend, error) {
	backend := &GarageBackend{
		config: config,
	}

	// Parse endpoint to determine if secure
	u, err := xnet.ParseHTTPURL(config.Endpoint)
	if err != nil {
		return nil, err
	}
	
	secure := u.Scheme == "https"
	
	// Create credentials
	var creds *credentials.Credentials
	if config.AccessKey != "" && config.SecretKey != "" {
		creds = credentials.NewStaticV4(config.AccessKey, config.SecretKey, "")
	} else {
		return nil, fmt.Errorf("garage backend requires access key and secret key")
	}

	// If cluster mode is enabled, create clients for each endpoint
	if config.ClusterMode && len(config.ClusterEndpoints) > 0 {
		backend.clients = make([]*minio.Client, len(config.ClusterEndpoints))
		for i, endpoint := range config.ClusterEndpoints {
			u, err := xnet.ParseHTTPURL(endpoint)
			if err != nil {
				return nil, fmt.Errorf("invalid cluster endpoint %s: %w", endpoint, err)
			}
			
			client, err := minio.New(u.Host, &minio.Options{
				Creds:  creds,
				Secure: u.Scheme == "https",
				Region: config.Region,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create client for endpoint %s: %w", endpoint, err)
			}
			backend.clients[i] = client
		}
	} else {
		// Single endpoint mode
		client, err := minio.New(u.Host, &minio.Options{
			Creds:  creds,
			Secure: secure,
			Region: config.Region,
		})
		if err != nil {
			return nil, err
		}
		backend.clients = []*minio.Client{client}
	}

	return backend, nil
}

// getClient returns a client using round-robin load balancing for cluster mode
func (g *GarageBackend) getClient() *minio.Client {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if len(g.clients) == 1 {
		return g.clients[0]
	}
	
	// Round-robin selection
	client := g.clients[g.current]
	g.current = (g.current + 1) % len(g.clients)
	return client
}

// getRandomClient returns a random client (for read operations in cluster mode)
func (g *GarageBackend) getRandomClient() *minio.Client {
	if len(g.clients) == 1 {
		return g.clients[0]
	}
	
	return g.clients[rand.Intn(len(g.clients))]
}

// Bucket operations
func (g *GarageBackend) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return g.getClient().ListBuckets(ctx)
}

func (g *GarageBackend) MakeBucket(ctx context.Context, bucketName, location string, objectLocking bool) error {
	// Note: Garage may not support object locking in all versions
	return g.getClient().MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region:        location,
		ObjectLocking: objectLocking,
	})
}

func (g *GarageBackend) RemoveBucket(ctx context.Context, bucketName string) error {
	return g.getClient().RemoveBucket(ctx, bucketName)
}

func (g *GarageBackend) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return g.getRandomClient().BucketExists(ctx, bucketName)
}

// Bucket policy operations
func (g *GarageBackend) SetBucketPolicy(ctx context.Context, bucketName, policy string) error {
	return g.getClient().SetBucketPolicy(ctx, bucketName, policy)
}

func (g *GarageBackend) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	return g.getRandomClient().GetBucketPolicy(ctx, bucketName)
}

// Bucket notification operations
func (g *GarageBackend) GetBucketNotification(ctx context.Context, bucketName string) (notification.Configuration, error) {
	// Note: Garage may have limited notification support
	return g.getRandomClient().GetBucketNotification(ctx, bucketName)
}

func (g *GarageBackend) SetBucketNotification(ctx context.Context, bucketName string, config notification.Configuration) error {
	return g.getClient().SetBucketNotification(ctx, bucketName, config)
}

// Object operations
func (g *GarageBackend) ListObjects(ctx context.Context, bucket string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return g.getRandomClient().ListObjects(ctx, bucket, opts)
}

func (g *GarageBackend) StatObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (minio.ObjectInfo, error) {
	return g.getRandomClient().StatObject(ctx, bucketName, objectName, opts)
}

func (g *GarageBackend) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return g.getClient().PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (g *GarageBackend) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return g.getRandomClient().GetObject(ctx, bucketName, objectName, opts)
}

func (g *GarageBackend) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return g.getClient().RemoveObject(ctx, bucketName, objectName, opts)
}

func (g *GarageBackend) CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error) {
	return g.getClient().CopyObject(ctx, dst, src)
}

// Object retention operations (may not be fully supported by Garage)
func (g *GarageBackend) GetObjectRetention(ctx context.Context, bucketName, objectName, versionID string) (*minio.RetentionMode, *time.Time, error) {
	mode, retainUntilDate, err := g.getRandomClient().GetObjectRetention(ctx, bucketName, objectName, versionID)
	return &mode, &retainUntilDate, err
}

func (g *GarageBackend) PutObjectRetention(ctx context.Context, bucketName, objectName string, opts minio.PutObjectRetentionOptions) error {
	return g.getClient().PutObjectRetention(ctx, bucketName, objectName, opts)
}

// Object legal hold operations (may not be fully supported by Garage)
func (g *GarageBackend) GetObjectLegalHold(ctx context.Context, bucketName, objectName string, opts minio.GetObjectLegalHoldOptions) (*minio.LegalHoldStatus, error) {
	status, err := g.getRandomClient().GetObjectLegalHold(ctx, bucketName, objectName, opts)
	return &status, err
}

func (g *GarageBackend) PutObjectLegalHold(ctx context.Context, bucketName, objectName string, opts minio.PutObjectLegalHoldOptions) error {
	return g.getClient().PutObjectLegalHold(ctx, bucketName, objectName, opts)
}

// Object tagging operations
func (g *GarageBackend) GetObjectTagging(ctx context.Context, bucketName, objectName string, opts minio.GetObjectTaggingOptions) (*tags.Tags, error) {
	return g.getRandomClient().GetObjectTagging(ctx, bucketName, objectName, opts)
}

func (g *GarageBackend) PutObjectTagging(ctx context.Context, bucketName, objectName string, otags *tags.Tags, opts minio.PutObjectTaggingOptions) error {
	return g.getClient().PutObjectTagging(ctx, bucketName, objectName, otags, opts)
}

// Bucket encryption operations (may not be fully supported by Garage)
func (g *GarageBackend) SetBucketEncryption(ctx context.Context, bucketName string, config *sse.Configuration) error {
	return g.getClient().SetBucketEncryption(ctx, bucketName, config)
}

func (g *GarageBackend) GetBucketEncryption(ctx context.Context, bucketName string) (*sse.Configuration, error) {
	return g.getRandomClient().GetBucketEncryption(ctx, bucketName)
}

func (g *GarageBackend) RemoveBucketEncryption(ctx context.Context, bucketName string) error {
	return g.getClient().RemoveBucketEncryption(ctx, bucketName)
}

// Bucket lifecycle operations
func (g *GarageBackend) GetBucketLifecycle(ctx context.Context, bucketName string) (*lifecycle.Configuration, error) {
	return g.getRandomClient().GetBucketLifecycle(ctx, bucketName)
}

func (g *GarageBackend) SetBucketLifecycle(ctx context.Context, bucketName string, config *lifecycle.Configuration) error {
	return g.getClient().SetBucketLifecycle(ctx, bucketName, config)
}

// Bucket object lock operations (may not be fully supported by Garage)
func (g *GarageBackend) GetBucketObjectLockConfig(ctx context.Context, bucketName string) (*minio.RetentionMode, *uint, *minio.ValidityUnit, error) {
	mode, validity, unit, err := g.getRandomClient().GetBucketObjectLockConfig(ctx, bucketName)
	return &mode, &validity, &unit, err
}

func (g *GarageBackend) SetObjectLockConfig(ctx context.Context, bucketName string, mode *minio.RetentionMode, validity *uint, unit *minio.ValidityUnit) error {
	return g.getClient().SetObjectLockConfig(ctx, bucketName, mode, validity, unit)
}

// Bucket tagging operations
func (g *GarageBackend) GetBucketTagging(ctx context.Context, bucketName string) (*tags.Tags, error) {
	return g.getRandomClient().GetBucketTagging(ctx, bucketName)
}

func (g *GarageBackend) SetBucketTagging(ctx context.Context, bucketName string, t *tags.Tags) error {
	return g.getClient().SetBucketTagging(ctx, bucketName, t)
}

func (g *GarageBackend) RemoveBucketTagging(ctx context.Context, bucketName string) error {
	return g.getClient().RemoveBucketTagging(ctx, bucketName)
}

// Versioning operations
func (g *GarageBackend) GetBucketVersioning(ctx context.Context, bucketName string) (minio.BucketVersioningConfiguration, error) {
	return g.getRandomClient().GetBucketVersioning(ctx, bucketName)
}

func (g *GarageBackend) SetBucketVersioning(ctx context.Context, bucketName string, config minio.BucketVersioningConfiguration) error {
	return g.getClient().SetBucketVersioning(ctx, bucketName, config)
}

// Replication operations (likely not supported by Garage)
func (g *GarageBackend) GetBucketReplication(ctx context.Context, bucketName string) (replication.Config, error) {
	return g.getRandomClient().GetBucketReplication(ctx, bucketName)
}

func (g *GarageBackend) SetBucketReplication(ctx context.Context, bucketName string, cfg replication.Config) error {
	return g.getClient().SetBucketReplication(ctx, bucketName, cfg)
}

func (g *GarageBackend) RemoveBucketReplication(ctx context.Context, bucketName string) error {
	return g.getClient().RemoveBucketReplication(ctx, bucketName)
}

// Backend-specific info
func (g *GarageBackend) GetBackendType() BackendType {
	return BackendTypeGarage
}

func (g *GarageBackend) GetBackendID() string {
	return g.config.ID
}

func (g *GarageBackend) IsHealthy(ctx context.Context) bool {
	// Check health of all clients in cluster mode
	healthyCount := 0
	for _, client := range g.clients {
		_, err := client.ListBuckets(ctx)
		if err == nil || !strings.Contains(err.Error(), "connection") {
			healthyCount++
		}
	}
	
	// Consider healthy if at least one client is working
	return healthyCount > 0
}
