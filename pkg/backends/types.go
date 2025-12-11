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
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/minio/minio-go/v7/pkg/notification"
	"github.com/minio/minio-go/v7/pkg/replication"
	"github.com/minio/minio-go/v7/pkg/sse"
	"github.com/minio/minio-go/v7/pkg/tags"
)

// BackendType represents the type of S3-compatible backend
type BackendType string

const (
	// BackendTypeMinIO represents MinIO backend
	BackendTypeMinIO BackendType = "minio"
	// BackendTypeGarage represents Garage backend
	BackendTypeGarage BackendType = "garage"
	// BackendTypeVersity represents Versity Gateway backend
	BackendTypeVersity BackendType = "versity"
	// BackendTypeS3 represents generic S3-compatible backend
	BackendTypeS3 BackendType = "s3"
)

// BackendConfig represents configuration for a single S3 backend
type BackendConfig struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        BackendType `json:"type"`
	Endpoint    string      `json:"endpoint"`
	Region      string      `json:"region"`
	Secure      bool        `json:"secure"`
	AccessKey   string      `json:"-"` // Never serialize
	SecretKey   string      `json:"-"` // Never serialize
	ClusterMode bool        `json:"clusterMode"`
	// ClusterEndpoints is used for backends like Garage that support cluster mode
	ClusterEndpoints []string          `json:"clusterEndpoints,omitempty"`
	Enabled          bool              `json:"enabled"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// BackendClient is the interface that all S3 backend implementations must satisfy
type BackendClient interface {
	// Bucket operations
	ListBuckets(ctx context.Context) ([]minio.BucketInfo, error)
	MakeBucket(ctx context.Context, bucketName, location string, objectLocking bool) error
	RemoveBucket(ctx context.Context, bucketName string) error
	BucketExists(ctx context.Context, bucketName string) (bool, error)

	// Bucket policy operations
	SetBucketPolicy(ctx context.Context, bucketName, policy string) error
	GetBucketPolicy(ctx context.Context, bucketName string) (string, error)

	// Bucket notification operations
	GetBucketNotification(ctx context.Context, bucketName string) (notification.Configuration, error)
	SetBucketNotification(ctx context.Context, bucketName string, config notification.Configuration) error

	// Object operations
	ListObjects(ctx context.Context, bucket string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	StatObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (minio.ObjectInfo, error)
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
	CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error)

	// Object retention operations
	GetObjectRetention(ctx context.Context, bucketName, objectName, versionID string) (*minio.RetentionMode, *time.Time, error)
	PutObjectRetention(ctx context.Context, bucketName, objectName string, opts minio.PutObjectRetentionOptions) error

	// Object legal hold operations
	GetObjectLegalHold(ctx context.Context, bucketName, objectName string, opts minio.GetObjectLegalHoldOptions) (*minio.LegalHoldStatus, error)
	PutObjectLegalHold(ctx context.Context, bucketName, objectName string, opts minio.PutObjectLegalHoldOptions) error

	// Object tagging operations
	GetObjectTagging(ctx context.Context, bucketName, objectName string, opts minio.GetObjectTaggingOptions) (*tags.Tags, error)
	PutObjectTagging(ctx context.Context, bucketName, objectName string, otags *tags.Tags, opts minio.PutObjectTaggingOptions) error

	// Bucket encryption operations
	SetBucketEncryption(ctx context.Context, bucketName string, config *sse.Configuration) error
	GetBucketEncryption(ctx context.Context, bucketName string) (*sse.Configuration, error)
	RemoveBucketEncryption(ctx context.Context, bucketName string) error

	// Bucket lifecycle operations
	GetBucketLifecycle(ctx context.Context, bucketName string) (*lifecycle.Configuration, error)
	SetBucketLifecycle(ctx context.Context, bucketName string, config *lifecycle.Configuration) error

	// Bucket object lock operations
	GetBucketObjectLockConfig(ctx context.Context, bucketName string) (*minio.RetentionMode, *uint, *minio.ValidityUnit, error)
	SetObjectLockConfig(ctx context.Context, bucketName string, mode *minio.RetentionMode, validity *uint, unit *minio.ValidityUnit) error

	// Bucket tagging operations
	GetBucketTagging(ctx context.Context, bucketName string) (*tags.Tags, error)
	SetBucketTagging(ctx context.Context, bucketName string, tags *tags.Tags) error
	RemoveBucketTagging(ctx context.Context, bucketName string) error

	// Versioning operations
	GetBucketVersioning(ctx context.Context, bucketName string) (minio.BucketVersioningConfiguration, error)
	SetBucketVersioning(ctx context.Context, bucketName string, config minio.BucketVersioningConfiguration) error

	// Replication operations (may not be supported by all backends)
	GetBucketReplication(ctx context.Context, bucketName string) (replication.Config, error)
	SetBucketReplication(ctx context.Context, bucketName string, cfg replication.Config) error
	RemoveBucketReplication(ctx context.Context, bucketName string) error

	// Backend-specific info
	GetBackendType() BackendType
	GetBackendID() string
	IsHealthy(ctx context.Context) bool
}
