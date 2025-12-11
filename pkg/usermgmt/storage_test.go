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
	"testing"
)

func TestCreateUser(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	req := &CreateUserRequest{
		Username:    "testuser",
		Email:       "test@example.com",
		Password:    "password123",
		Type:        UserTypeAdmin,
		Permissions: []string{"manage:backends"},
		BackendIDs:  []string{"backend1"},
	}
	
	user, err := storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
	if user.Type != UserTypeAdmin {
		t.Errorf("Expected type Admin, got %s", user.Type)
	}
	if user.Status != UserStatusActive {
		t.Errorf("Expected status Active, got %s", user.Status)
	}
	if user.ID == "" {
		t.Error("User ID should not be empty")
	}
}

func TestCreateUserDuplicate(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	req := &CreateUserRequest{
		Username: "duplicate",
		Email:    "duplicate@example.com",
		Password: "password123",
		Type:     UserTypeAdmin,
	}
	
	// Create first user
	_, err := storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}
	
	// Try to create duplicate
	_, err = storage.CreateUser(ctx, req)
	if err != ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestGetUser(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	req := &CreateUserRequest{
		Username: "getuser",
		Email:    "getuser@example.com",
		Password: "password123",
		Type:     UserTypeS3,
	}
	
	created, err := storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Get user by ID
	retrieved, err := storage.GetUser(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	
	if retrieved.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, retrieved.ID)
	}
	if retrieved.Username != "getuser" {
		t.Errorf("Expected username 'getuser', got '%s'", retrieved.Username)
	}
}

func TestGetUserByUsername(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	req := &CreateUserRequest{
		Username: "byname",
		Email:    "byname@example.com",
		Password: "password123",
		Type:     UserTypeAdmin,
	}
	
	created, err := storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Get user by username
	retrieved, err := storage.GetUserByUsername(ctx, "byname")
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}
	
	if retrieved.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, retrieved.ID)
	}
}

func TestGetUserNotFound(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	_, err := storage.GetUser(ctx, "non-existent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	// Create user
	req := &CreateUserRequest{
		Username: "updateuser",
		Email:    "old@example.com",
		Password: "password123",
		Type:     UserTypeAdmin,
	}
	
	created, err := storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Update user
	newEmail := "new@example.com"
	newStatus := UserStatusInactive
	updateReq := &UpdateUserRequest{
		Email:  &newEmail,
		Status: &newStatus,
	}
	
	updated, err := storage.UpdateUser(ctx, created.ID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	
	if updated.Email != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got '%s'", updated.Email)
	}
	if updated.Status != UserStatusInactive {
		t.Errorf("Expected status Inactive, got %s", updated.Status)
	}
}

func TestDeleteUser(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	req := &CreateUserRequest{
		Username: "deleteuser",
		Email:    "delete@example.com",
		Password: "password123",
		Type:     UserTypeAdmin,
	}
	
	created, err := storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Delete user
	err = storage.DeleteUser(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	
	// Verify user is deleted
	_, err = storage.GetUser(ctx, created.ID)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestListUsers(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	// Create multiple users
	for i := 0; i < 5; i++ {
		userType := UserTypeAdmin
		if i%2 == 0 {
			userType = UserTypeS3
		}
		
		req := &CreateUserRequest{
			Username: "user" + string(rune('0'+i)),
			Email:    "user" + string(rune('0'+i)) + "@example.com",
			Password: "password123",
			Type:     userType,
		}
		
		_, err := storage.CreateUser(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}
	
	// List all users
	response, err := storage.ListUsers(ctx, nil, 1, 10)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	
	if response.Total != 5 {
		t.Errorf("Expected 5 total users, got %d", response.Total)
	}
	if len(response.Users) != 5 {
		t.Errorf("Expected 5 users in response, got %d", len(response.Users))
	}
}

func TestListUsersByType(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	// Create users of different types
	for i := 0; i < 3; i++ {
		req := &CreateUserRequest{
			Username: "admin" + string(rune('0'+i)),
			Email:    "admin" + string(rune('0'+i)) + "@example.com",
			Password: "password123",
			Type:     UserTypeAdmin,
		}
		storage.CreateUser(ctx, req)
	}
	
	for i := 0; i < 2; i++ {
		req := &CreateUserRequest{
			Username: "s3user" + string(rune('0'+i)),
			Email:    "s3user" + string(rune('0'+i)) + "@example.com",
			Password: "password123",
			Type:     UserTypeS3,
		}
		storage.CreateUser(ctx, req)
	}
	
	// List only admin users
	adminType := UserTypeAdmin
	response, err := storage.ListUsers(ctx, &adminType, 1, 10)
	if err != nil {
		t.Fatalf("Failed to list admin users: %v", err)
	}
	
	if response.Total != 3 {
		t.Errorf("Expected 3 admin users, got %d", response.Total)
	}
	
	// List only S3 users
	s3Type := UserTypeS3
	response, err = storage.ListUsers(ctx, &s3Type, 1, 10)
	if err != nil {
		t.Fatalf("Failed to list S3 users: %v", err)
	}
	
	if response.Total != 2 {
		t.Errorf("Expected 2 S3 users, got %d", response.Total)
	}
}

func TestAuthenticateUser(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	password := "mypassword123"
	req := &CreateUserRequest{
		Username: "authuser",
		Email:    "auth@example.com",
		Password: password,
		Type:     UserTypeAdmin,
	}
	
	_, err := storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Authenticate with correct password
	user, err := storage.AuthenticateUser(ctx, "authuser", password)
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}
	
	if user.Username != "authuser" {
		t.Errorf("Expected username 'authuser', got '%s'", user.Username)
	}
	
	if user.LastLogin == nil {
		t.Error("LastLogin should be set after authentication")
	}
}

func TestAuthenticateUserWrongPassword(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	req := &CreateUserRequest{
		Username: "wrongpass",
		Email:    "wrongpass@example.com",
		Password: "correctpassword",
		Type:     UserTypeAdmin,
	}
	
	storage.CreateUser(ctx, req)
	
	// Try to authenticate with wrong password
	_, err := storage.AuthenticateUser(ctx, "wrongpass", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthenticateUserNotFound(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	_, err := storage.AuthenticateUser(ctx, "nonexistent", "password")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestCreateS3UserConfig(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	// Create a user first
	userReq := &CreateUserRequest{
		Username: "s3configuser",
		Email:    "s3config@example.com",
		Password: "password123",
		Type:     UserTypeS3,
	}
	
	user, err := storage.CreateUser(ctx, userReq)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Create S3 config
	configReq := &CreateS3UserConfigRequest{
		UserID:    user.ID,
		BackendID: "backend1",
		AccessKey: "accesskey123",
		SecretKey: "secretkey123",
		Policies:  []string{"policy1"},
	}
	
	config, err := storage.CreateS3UserConfig(ctx, configReq)
	if err != nil {
		t.Fatalf("Failed to create S3 config: %v", err)
	}
	
	if config.UserID != user.ID {
		t.Errorf("Expected UserID '%s', got '%s'", user.ID, config.UserID)
	}
	if config.BackendID != "backend1" {
		t.Errorf("Expected BackendID 'backend1', got '%s'", config.BackendID)
	}
	if !config.Enabled {
		t.Error("Config should be enabled by default")
	}
}

func TestGetS3UserConfig(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	// Create user and config
	userReq := &CreateUserRequest{
		Username: "getconfig",
		Email:    "getconfig@example.com",
		Password: "password123",
		Type:     UserTypeS3,
	}
	user, _ := storage.CreateUser(ctx, userReq)
	
	configReq := &CreateS3UserConfigRequest{
		UserID:    user.ID,
		BackendID: "backend1",
		AccessKey: "accesskey123",
		SecretKey: "secretkey123",
	}
	storage.CreateS3UserConfig(ctx, configReq)
	
	// Get config
	config, err := storage.GetS3UserConfig(ctx, user.ID, "backend1")
	if err != nil {
		t.Fatalf("Failed to get S3 config: %v", err)
	}
	
	if config.AccessKey != "accesskey123" {
		t.Errorf("Expected AccessKey 'accesskey123', got '%s'", config.AccessKey)
	}
}

func TestListS3UserConfigs(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	// Create user
	userReq := &CreateUserRequest{
		Username: "listconfigs",
		Email:    "listconfigs@example.com",
		Password: "password123",
		Type:     UserTypeS3,
	}
	user, _ := storage.CreateUser(ctx, userReq)
	
	// Create multiple configs
	backends := []string{"backend1", "backend2", "backend3"}
	for _, backend := range backends {
		configReq := &CreateS3UserConfigRequest{
			UserID:    user.ID,
			BackendID: backend,
			AccessKey: "key_" + backend,
			SecretKey: "secret_" + backend,
		}
		storage.CreateS3UserConfig(ctx, configReq)
	}
	
	// List configs
	configs, err := storage.ListS3UserConfigs(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list S3 configs: %v", err)
	}
	
	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}
}

func TestDeleteS3UserConfig(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	
	// Create user and config
	userReq := &CreateUserRequest{
		Username: "deleteconfig",
		Email:    "deleteconfig@example.com",
		Password: "password123",
		Type:     UserTypeS3,
	}
	user, _ := storage.CreateUser(ctx, userReq)
	
	configReq := &CreateS3UserConfigRequest{
		UserID:    user.ID,
		BackendID: "backend1",
		AccessKey: "accesskey123",
		SecretKey: "secretkey123",
	}
	storage.CreateS3UserConfig(ctx, configReq)
	
	// Delete config
	err := storage.DeleteS3UserConfig(ctx, user.ID, "backend1")
	if err != nil {
		t.Fatalf("Failed to delete S3 config: %v", err)
	}
	
	// Verify deletion
	_, err = storage.GetS3UserConfig(ctx, user.ID, "backend1")
	if err != ErrS3ConfigNotFound {
		t.Errorf("Expected ErrS3ConfigNotFound, got: %v", err)
	}
}
