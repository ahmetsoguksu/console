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

	"github.com/minio/console/pkg/usermgmt"
)

// CreateAdminUser creates a new admin user
func CreateAdminUser(ctx context.Context, req *usermgmt.CreateUserRequest) (*usermgmt.User, error) {
	// Ensure this is an admin user
	req.Type = usermgmt.UserTypeAdmin
	
	user, err := usermgmt.GlobalUserStorage.CreateUser(ctx, req)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	return user, nil
}

// CreateS3User creates a new S3 user
func CreateS3User(ctx context.Context, req *usermgmt.CreateUserRequest) (*usermgmt.User, error) {
	// Ensure this is an S3 user
	req.Type = usermgmt.UserTypeS3
	
	user, err := usermgmt.GlobalUserStorage.CreateUser(ctx, req)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	return user, nil
}

// GetUser retrieves a user by ID
func GetUser(ctx context.Context, userID string) (*usermgmt.User, error) {
	user, err := usermgmt.GlobalUserStorage.GetUser(ctx, userID)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	return user, nil
}

// UpdateUser updates a user
func UpdateUser(ctx context.Context, userID string, req *usermgmt.UpdateUserRequest) (*usermgmt.User, error) {
	user, err := usermgmt.GlobalUserStorage.UpdateUser(ctx, userID, req)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	return user, nil
}

// DeleteUser deletes a user
func DeleteUser(ctx context.Context, userID string) error {
	if err := usermgmt.GlobalUserStorage.DeleteUser(ctx, userID); err != nil {
		return ErrorWithContext(ctx, err)
	}
	
	return nil
}

// ListUsers lists users with optional filtering
func ListUsers(ctx context.Context, userType *usermgmt.UserType, page, pageSize int) (*usermgmt.UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	users, err := usermgmt.GlobalUserStorage.ListUsers(ctx, userType, page, pageSize)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	return users, nil
}

// AuthenticateAdminUser authenticates an admin user
func AuthenticateAdminUser(ctx context.Context, username, password string) (*usermgmt.User, error) {
	user, err := usermgmt.GlobalUserStorage.AuthenticateUser(ctx, username, password)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	// Ensure this is an admin user
	if user.Type != usermgmt.UserTypeAdmin {
		return nil, ErrorWithContext(ctx, usermgmt.ErrInvalidCredentials)
	}
	
	return user, nil
}

// CreateS3UserConfig creates S3 user configuration for a backend
func CreateS3UserConfig(ctx context.Context, req *usermgmt.CreateS3UserConfigRequest) (*usermgmt.S3UserConfig, error) {
	config, err := usermgmt.GlobalUserStorage.CreateS3UserConfig(ctx, req)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	return config, nil
}

// GetS3UserConfig retrieves S3 user configuration
func GetS3UserConfig(ctx context.Context, userID, backendID string) (*usermgmt.S3UserConfig, error) {
	config, err := usermgmt.GlobalUserStorage.GetS3UserConfig(ctx, userID, backendID)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	return config, nil
}

// ListS3UserConfigs lists all S3 configurations for a user
func ListS3UserConfigs(ctx context.Context, userID string) ([]*usermgmt.S3UserConfig, error) {
	configs, err := usermgmt.GlobalUserStorage.ListS3UserConfigs(ctx, userID)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	
	return configs, nil
}

// UpdateS3UserConfig updates S3 user configuration
func UpdateS3UserConfig(ctx context.Context, userID, backendID string, config *usermgmt.S3UserConfig) error {
	if err := usermgmt.GlobalUserStorage.UpdateS3UserConfig(ctx, userID, backendID, config); err != nil {
		return ErrorWithContext(ctx, err)
	}
	
	return nil
}

// DeleteS3UserConfig deletes S3 user configuration
func DeleteS3UserConfig(ctx context.Context, userID, backendID string) error {
	if err := usermgmt.GlobalUserStorage.DeleteS3UserConfig(ctx, userID, backendID); err != nil {
		return ErrorWithContext(ctx, err)
	}
	
	return nil
}

// HasPermission checks if a user has a specific permission
func HasPermission(user *usermgmt.User, permission string) bool {
	if user == nil {
		return false
	}
	
	for _, perm := range user.Permissions {
		if perm == permission || perm == "*" {
			return true
		}
	}
	
	return false
}

// HasBackendAccess checks if a user has access to a specific backend
func HasBackendAccess(user *usermgmt.User, backendID string) bool {
	if user == nil {
		return false
	}
	
	// Admin users with manage:backends permission have access to all backends
	if user.Type == usermgmt.UserTypeAdmin && HasPermission(user, string(usermgmt.PermissionManageBackends)) {
		return true
	}
	
	// Check specific backend access
	for _, id := range user.BackendIDs {
		if id == backendID || id == "*" {
			return true
		}
	}
	
	return false
}

// UserResponse wraps user information for API responses
type UserResponse struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Type        usermgmt.UserType      `json:"type"`
	Status      usermgmt.UserStatus    `json:"status"`
	Permissions []string               `json:"permissions"`
	BackendIDs  []string               `json:"backendIds"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
	LastLogin   *string                `json:"lastLogin,omitempty"`
}

// ConvertUserToResponse converts a user to API response
func ConvertUserToResponse(user *usermgmt.User) *UserResponse {
	resp := &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Type:        user.Type,
		Status:      user.Status,
		Permissions: user.Permissions,
		BackendIDs:  user.BackendIDs,
		Metadata:    user.Metadata,
		CreatedAt:   user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	if user.LastLogin != nil {
		lastLogin := user.LastLogin.Format("2006-01-02T15:04:05Z07:00")
		resp.LastLogin = &lastLogin
	}
	
	return resp
}

// S3UserConfigResponse wraps S3 user config for API responses
type S3UserConfigResponse struct {
	UserID      string   `json:"userId"`
	BackendID   string   `json:"backendId"`
	AccessKey   string   `json:"accessKey"`
	Policies    []string `json:"policies"`
	Groups      []string `json:"groups"`
	Enabled     bool     `json:"enabled"`
	SyncEnabled bool     `json:"syncEnabled"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// ConvertS3ConfigToResponse converts S3 config to API response
func ConvertS3ConfigToResponse(config *usermgmt.S3UserConfig) *S3UserConfigResponse {
	return &S3UserConfigResponse{
		UserID:      config.UserID,
		BackendID:   config.BackendID,
		AccessKey:   config.AccessKey,
		Policies:    config.Policies,
		Groups:      config.Groups,
		Enabled:     config.Enabled,
		SyncEnabled: config.SyncEnabled,
		CreatedAt:   config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
