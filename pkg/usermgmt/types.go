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

package usermgmt

import (
	"time"
)

// UserType represents the type of user
type UserType string

const (
	// UserTypeAdmin represents admin users who manage the console and backends
	UserTypeAdmin UserType = "admin"
	// UserTypeS3 represents S3 users who access object storage
	UserTypeS3 UserType = "s3"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	// UserStatusActive represents an active user
	UserStatusActive UserStatus = "active"
	// UserStatusInactive represents an inactive user
	UserStatusInactive UserStatus = "inactive"
	// UserStatusLocked represents a locked user
	UserStatusLocked UserStatus = "locked"
)

// User represents a user in the system
type User struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	Type        UserType          `json:"type"`
	Status      UserStatus        `json:"status"`
	Password    string            `json:"-"` // Never serialize
	Permissions []string          `json:"permissions"`
	BackendIDs  []string          `json:"backendIds"` // Backends this user has access to
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	LastLogin   *time.Time        `json:"lastLogin,omitempty"`
}

// S3UserConfig represents S3-specific user configuration for a backend
type S3UserConfig struct {
	UserID      string    `json:"userId"`      // Reference to User.ID
	BackendID   string    `json:"backendId"`   // Backend this config applies to
	AccessKey   string    `json:"accessKey"`   // S3 access key
	SecretKey   string    `json:"-"`           // S3 secret key, never serialize
	Policies    []string  `json:"policies"`    // S3 policies
	Groups      []string  `json:"groups"`      // S3 groups
	Enabled     bool      `json:"enabled"`     // Whether user is enabled on this backend
	SyncEnabled bool      `json:"syncEnabled"` // Whether to sync changes to backend
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AdminPermission represents admin-level permissions
type AdminPermission string

const (
	// PermissionManageBackends allows managing backend configurations
	PermissionManageBackends AdminPermission = "manage:backends"
	// PermissionManageUsers allows managing users
	PermissionManageUsers AdminPermission = "manage:users"
	// PermissionManageS3Users allows managing S3 users across backends
	PermissionManageS3Users AdminPermission = "manage:s3users"
	// PermissionViewBackends allows viewing backend information
	PermissionViewBackends AdminPermission = "view:backends"
	// PermissionViewUsers allows viewing user information
	PermissionViewUsers AdminPermission = "view:users"
	// PermissionViewS3Users allows viewing S3 user information
	PermissionViewS3Users AdminPermission = "view:s3users"
	// PermissionAccessBackend allows accessing a specific backend
	PermissionAccessBackend AdminPermission = "access:backend"
	// PermissionManageBuckets allows managing buckets
	PermissionManageBuckets AdminPermission = "manage:buckets"
	// PermissionManageObjects allows managing objects
	PermissionManageObjects AdminPermission = "manage:objects"
)

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username    string            `json:"username" validate:"required"`
	Email       string            `json:"email" validate:"required,email"`
	Password    string            `json:"password" validate:"required,min=8"`
	Type        UserType          `json:"type" validate:"required"`
	Permissions []string          `json:"permissions"`
	BackendIDs  []string          `json:"backendIds"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email       *string           `json:"email,omitempty"`
	Status      *UserStatus       `json:"status,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	BackendIDs  []string          `json:"backendIds,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateS3UserConfigRequest represents a request to create S3 user configuration
type CreateS3UserConfigRequest struct {
	UserID      string   `json:"userId" validate:"required"`
	BackendID   string   `json:"backendId" validate:"required"`
	AccessKey   string   `json:"accessKey" validate:"required"`
	SecretKey   string   `json:"secretKey" validate:"required"`
	Policies    []string `json:"policies"`
	Groups      []string `json:"groups"`
	SyncEnabled bool     `json:"syncEnabled"`
}

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users      []*User `json:"users"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"pageSize"`
	TotalPages int     `json:"totalPages"`
}
