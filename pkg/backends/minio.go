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
	"io"
	"strings"
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

// MinIOBackend implements BackendClient for MinIO
type MinIOBackend struct {
	client *minio.Client
	config *BackendConfig
}

// NewMinIOBackend creates a new MinIO backend client
func NewMinIOBackend(config *BackendConfig) (*MinIOBackend, error) {
	// Parse endpoint to determine if secure
	u, err := xnet.ParseHTTPURL(config.Endpoint)
	if err != nil {
		return nil, err
	}
	
	secure := u.Scheme == "https"
	endpoint := u.Host

	// Create credentials
	var creds *credentials.Credentials
	if config.AccessKey != "" && config.SecretKey != "" {
		creds = credentials.NewStaticV4(config.AccessKey, config.SecretKey, "")
	} else {
		creds = credentials.NewIAM("") // Use IAM role if no credentials provided
	}

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: secure,
		Region: config.Region,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOBackend{
		client: client,
		config: config,
	}, nil
}

// Bucket operations
func (m *MinIOBackend) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return m.client.ListBuckets(ctx)
}

func (m *MinIOBackend) MakeBucket(ctx context.Context, bucketName, location string, objectLocking bool) error {
	return m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region:        location,
		ObjectLocking: objectLocking,
	})
}

func (m *MinIOBackend) RemoveBucket(ctx context.Context, bucketName string) error {
	return m.client.RemoveBucket(ctx, bucketName)
}

func (m *MinIOBackend) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return m.client.BucketExists(ctx, bucketName)
}

// Bucket policy operations
func (m *MinIOBackend) SetBucketPolicy(ctx context.Context, bucketName, policy string) error {
	return m.client.SetBucketPolicy(ctx, bucketName, policy)
}

func (m *MinIOBackend) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	return m.client.GetBucketPolicy(ctx, bucketName)
}

// Bucket notification operations
func (m *MinIOBackend) GetBucketNotification(ctx context.Context, bucketName string) (notification.Configuration, error) {
	return m.client.GetBucketNotification(ctx, bucketName)
}

func (m *MinIOBackend) SetBucketNotification(ctx context.Context, bucketName string, config notification.Configuration) error {
	return m.client.SetBucketNotification(ctx, bucketName, config)
}

// Object operations
func (m *MinIOBackend) ListObjects(ctx context.Context, bucket string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return m.client.ListObjects(ctx, bucket, opts)
}

func (m *MinIOBackend) StatObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (minio.ObjectInfo, error) {
	return m.client.StatObject(ctx, bucketName, objectName, opts)
}

func (m *MinIOBackend) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return m.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (m *MinIOBackend) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return m.client.GetObject(ctx, bucketName, objectName, opts)
}

func (m *MinIOBackend) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return m.client.RemoveObject(ctx, bucketName, objectName, opts)
}

func (m *MinIOBackend) CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error) {
	return m.client.CopyObject(ctx, dst, src)
}

// Object retention operations
func (m *MinIOBackend) GetObjectRetention(ctx context.Context, bucketName, objectName, versionID string) (*minio.RetentionMode, *time.Time, error) {
	return m.client.GetObjectRetention(ctx, bucketName, objectName, versionID)
}

func (m *MinIOBackend) PutObjectRetention(ctx context.Context, bucketName, objectName string, opts minio.PutObjectRetentionOptions) error {
	return m.client.PutObjectRetention(ctx, bucketName, objectName, opts)
}

// Object legal hold operations
func (m *MinIOBackend) GetObjectLegalHold(ctx context.Context, bucketName, objectName string, opts minio.GetObjectLegalHoldOptions) (*minio.LegalHoldStatus, error) {
	return m.client.GetObjectLegalHold(ctx, bucketName, objectName, opts)
}

func (m *MinIOBackend) PutObjectLegalHold(ctx context.Context, bucketName, objectName string, opts minio.PutObjectLegalHoldOptions) error {
	return m.client.PutObjectLegalHold(ctx, bucketName, objectName, opts)
}

// Object tagging operations
func (m *MinIOBackend) GetObjectTagging(ctx context.Context, bucketName, objectName string, opts minio.GetObjectTaggingOptions) (*tags.Tags, error) {
	return m.client.GetObjectTagging(ctx, bucketName, objectName, opts)
}

func (m *MinIOBackend) PutObjectTagging(ctx context.Context, bucketName, objectName string, otags *tags.Tags, opts minio.PutObjectTaggingOptions) error {
	return m.client.PutObjectTagging(ctx, bucketName, objectName, otags, opts)
}

// Bucket encryption operations
func (m *MinIOBackend) SetBucketEncryption(ctx context.Context, bucketName string, config *sse.Configuration) error {
	return m.client.SetBucketEncryption(ctx, bucketName, config)
}

func (m *MinIOBackend) GetBucketEncryption(ctx context.Context, bucketName string) (*sse.Configuration, error) {
	return m.client.GetBucketEncryption(ctx, bucketName)
}

func (m *MinIOBackend) RemoveBucketEncryption(ctx context.Context, bucketName string) error {
	return m.client.RemoveBucketEncryption(ctx, bucketName)
}

// Bucket lifecycle operations
func (m *MinIOBackend) GetBucketLifecycle(ctx context.Context, bucketName string) (*lifecycle.Configuration, error) {
	return m.client.GetBucketLifecycle(ctx, bucketName)
}

func (m *MinIOBackend) SetBucketLifecycle(ctx context.Context, bucketName string, config *lifecycle.Configuration) error {
	return m.client.SetBucketLifecycle(ctx, bucketName, config)
}

// Bucket object lock operations
func (m *MinIOBackend) GetBucketObjectLockConfig(ctx context.Context, bucketName string) (*minio.RetentionMode, *uint, *minio.ValidityUnit, error) {
	return m.client.GetBucketObjectLockConfig(ctx, bucketName)
}

func (m *MinIOBackend) SetObjectLockConfig(ctx context.Context, bucketName string, mode *minio.RetentionMode, validity *uint, unit *minio.ValidityUnit) error {
	return m.client.SetObjectLockConfig(ctx, bucketName, mode, validity, unit)
}

// Bucket tagging operations
func (m *MinIOBackend) GetBucketTagging(ctx context.Context, bucketName string) (*tags.Tags, error) {
	return m.client.GetBucketTagging(ctx, bucketName)
}

func (m *MinIOBackend) SetBucketTagging(ctx context.Context, bucketName string, t *tags.Tags) error {
	return m.client.SetBucketTagging(ctx, bucketName, t)
}

func (m *MinIOBackend) RemoveBucketTagging(ctx context.Context, bucketName string) error {
	return m.client.RemoveBucketTagging(ctx, bucketName)
}

// Versioning operations
func (m *MinIOBackend) GetBucketVersioning(ctx context.Context, bucketName string) (minio.BucketVersioningConfiguration, error) {
	return m.client.GetBucketVersioning(ctx, bucketName)
}

func (m *MinIOBackend) SetBucketVersioning(ctx context.Context, bucketName string, config minio.BucketVersioningConfiguration) error {
	return m.client.SetBucketVersioning(ctx, bucketName, config)
}

// Replication operations
func (m *MinIOBackend) GetBucketReplication(ctx context.Context, bucketName string) (replication.Config, error) {
	return m.client.GetBucketReplication(ctx, bucketName)
}

func (m *MinIOBackend) SetBucketReplication(ctx context.Context, bucketName string, cfg replication.Config) error {
	return m.client.SetBucketReplication(ctx, bucketName, cfg)
}

func (m *MinIOBackend) RemoveBucketReplication(ctx context.Context, bucketName string) error {
	return m.client.RemoveBucketReplication(ctx, bucketName)
}

// Backend-specific info
func (m *MinIOBackend) GetBackendType() BackendType {
	return BackendTypeMinIO
}

func (m *MinIOBackend) GetBackendID() string {
	return m.config.ID
}

func (m *MinIOBackend) IsHealthy(ctx context.Context) bool {
	// Try to list buckets as a health check
	_, err := m.client.ListBuckets(ctx)
	return err == nil || !strings.Contains(err.Error(), "connection")
}
