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
	"fmt"

	"github.com/minio/pkg/v3/env"
)

// InitializeAdminUser creates the default admin user if configured
func InitializeAdminUser() error {
	adminUsername := env.Get("CONSOLE_ADMIN_USERNAME", "")
	adminPassword := env.Get("CONSOLE_ADMIN_PASSWORD", "")
	
	if adminUsername == "" || adminPassword == "" {
		fmt.Println("No admin user configured (CONSOLE_ADMIN_USERNAME and CONSOLE_ADMIN_PASSWORD not set)")
		return nil
	}
	
	// Check if admin user already exists
	ctx := context.Background()
	_, err := GlobalUserStorage.GetUserByUsername(ctx, adminUsername)
	if err == nil {
		fmt.Printf("Admin user '%s' already exists\n", adminUsername)
		return nil
	}
	
	if err != ErrUserNotFound {
		return fmt.Errorf("error checking for existing admin user: %w", err)
	}
	
	// Get admin email from environment or use default
	adminEmail := env.Get("CONSOLE_ADMIN_EMAIL", fmt.Sprintf("%s@localhost", adminUsername))
	
	// Create admin user
	req := &CreateUserRequest{
		Username: adminUsername,
		Email:    adminEmail,
		Password: adminPassword,
		Type:     UserTypeAdmin,
		Permissions: []string{
			string(PermissionManageBackends),
			string(PermissionManageUsers),
			string(PermissionManageS3Users),
			string(PermissionViewBackends),
			string(PermissionViewUsers),
			string(PermissionViewS3Users),
			string(PermissionAccessBackend),
			string(PermissionManageBuckets),
			string(PermissionManageObjects),
		},
		BackendIDs: []string{"*"}, // Access to all backends
	}
	
	user, err := GlobalUserStorage.CreateUser(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}
	
	fmt.Printf("Admin user '%s' created successfully with ID: %s\n", user.Username, user.ID)
	return nil
}
