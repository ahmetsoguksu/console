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
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when a user already exists
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrS3ConfigNotFound is returned when S3 config is not found
	ErrS3ConfigNotFound = errors.New("s3 user config not found")
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Storage defines the interface for user storage
type Storage interface {
	// User operations
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, userType *UserType, page, pageSize int) (*UserListResponse, error)
	AuthenticateUser(ctx context.Context, username, password string) (*User, error)

	// S3 User Config operations
	CreateS3UserConfig(ctx context.Context, req *CreateS3UserConfigRequest) (*S3UserConfig, error)
	GetS3UserConfig(ctx context.Context, userID, backendID string) (*S3UserConfig, error)
	ListS3UserConfigs(ctx context.Context, userID string) ([]*S3UserConfig, error)
	UpdateS3UserConfig(ctx context.Context, userID, backendID string, config *S3UserConfig) error
	DeleteS3UserConfig(ctx context.Context, userID, backendID string) error
}

// MemoryStorage implements Storage interface with in-memory storage
type MemoryStorage struct {
	users          map[string]*User
	usersByName    map[string]*User
	s3Configs      map[string]map[string]*S3UserConfig // userID -> backendID -> config
	mu             sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:       make(map[string]*User),
		usersByName: make(map[string]*User),
		s3Configs:   make(map[string]map[string]*S3UserConfig),
	}
}

// CreateUser creates a new user
func (s *MemoryStorage) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.usersByName[req.Username]; exists {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		ID:          uuid.New().String(),
		Username:    req.Username,
		Email:       req.Email,
		Type:        req.Type,
		Status:      UserStatusActive,
		Password:    string(hashedPassword),
		Permissions: req.Permissions,
		BackendIDs:  req.BackendIDs,
		Metadata:    req.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.users[user.ID] = user
	s.usersByName[user.Username] = user

	return user, nil
}

// GetUser retrieves a user by ID
func (s *MemoryStorage) GetUser(ctx context.Context, id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *MemoryStorage) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByName[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateUser updates a user
func (s *MemoryStorage) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Update fields
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Permissions != nil {
		user.Permissions = req.Permissions
	}
	if req.BackendIDs != nil {
		user.BackendIDs = req.BackendIDs
	}
	if req.Metadata != nil {
		user.Metadata = req.Metadata
	}

	user.UpdatedAt = time.Now()

	return user, nil
}

// DeleteUser deletes a user
func (s *MemoryStorage) DeleteUser(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return ErrUserNotFound
	}

	delete(s.users, id)
	delete(s.usersByName, user.Username)
	delete(s.s3Configs, id)

	return nil
}

// ListUsers lists users with pagination
func (s *MemoryStorage) ListUsers(ctx context.Context, userType *UserType, page, pageSize int) (*UserListResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var allUsers []*User
	for _, user := range s.users {
		if userType == nil || user.Type == *userType {
			allUsers = append(allUsers, user)
		}
	}

	total := len(allUsers)
	totalPages := (total + pageSize - 1) / pageSize

	start := (page - 1) * pageSize
	if start >= total {
		return &UserListResponse{
			Users:      []*User{},
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		}, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return &UserListResponse{
		Users:      allUsers[start:end],
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// AuthenticateUser authenticates a user
func (s *MemoryStorage) AuthenticateUser(ctx context.Context, username, password string) (*User, error) {
	// Use read lock for user lookup
	s.mu.RLock()
	user, exists := s.usersByName[username]
	s.mu.RUnlock()
	
	if !exists {
		return nil, ErrInvalidCredentials
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login with write lock
	s.mu.Lock()
	now := time.Now()
	user.LastLogin = &now
	s.mu.Unlock()

	return user, nil
}

// CreateS3UserConfig creates S3 user configuration
func (s *MemoryStorage) CreateS3UserConfig(ctx context.Context, req *CreateS3UserConfigRequest) (*S3UserConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[req.UserID]; !exists {
		return nil, ErrUserNotFound
	}

	// Initialize map for user if not exists
	if s.s3Configs[req.UserID] == nil {
		s.s3Configs[req.UserID] = make(map[string]*S3UserConfig)
	}

	now := time.Now()
	config := &S3UserConfig{
		UserID:      req.UserID,
		BackendID:   req.BackendID,
		AccessKey:   req.AccessKey,
		SecretKey:   req.SecretKey,
		Policies:    req.Policies,
		Groups:      req.Groups,
		Enabled:     true,
		SyncEnabled: req.SyncEnabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.s3Configs[req.UserID][req.BackendID] = config

	return config, nil
}

// GetS3UserConfig retrieves S3 user configuration
func (s *MemoryStorage) GetS3UserConfig(ctx context.Context, userID, backendID string) (*S3UserConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userConfigs, exists := s.s3Configs[userID]
	if !exists {
		return nil, ErrS3ConfigNotFound
	}

	config, exists := userConfigs[backendID]
	if !exists {
		return nil, ErrS3ConfigNotFound
	}

	return config, nil
}

// ListS3UserConfigs lists all S3 configurations for a user
func (s *MemoryStorage) ListS3UserConfigs(ctx context.Context, userID string) ([]*S3UserConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userConfigs, exists := s.s3Configs[userID]
	if !exists {
		return []*S3UserConfig{}, nil
	}

	configs := make([]*S3UserConfig, 0, len(userConfigs))
	for _, config := range userConfigs {
		configs = append(configs, config)
	}

	return configs, nil
}

// UpdateS3UserConfig updates S3 user configuration
func (s *MemoryStorage) UpdateS3UserConfig(ctx context.Context, userID, backendID string, config *S3UserConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userConfigs, exists := s.s3Configs[userID]
	if !exists || userConfigs[backendID] == nil {
		return ErrS3ConfigNotFound
	}

	config.UpdatedAt = time.Now()
	s.s3Configs[userID][backendID] = config

	return nil
}

// DeleteS3UserConfig deletes S3 user configuration
func (s *MemoryStorage) DeleteS3UserConfig(ctx context.Context, userID, backendID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userConfigs, exists := s.s3Configs[userID]
	if !exists {
		return ErrS3ConfigNotFound
	}

	delete(userConfigs, backendID)

	return nil
}

// GlobalUserStorage is the global user storage instance
var GlobalUserStorage Storage

func init() {
	GlobalUserStorage = NewMemoryStorage()
}
