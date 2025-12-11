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
	"fmt"
	"os"

	"github.com/minio/pkg/v3/env"
)

// InitializeBackends initializes backends from environment configuration
func InitializeBackends() error {
	// Check if multi-backend mode is enabled
	enableMultiBackend := env.Get("CONSOLE_ENABLE_MULTI_BACKEND", "false")
	
	if enableMultiBackend == "true" {
		fmt.Println("Initializing multi-backend mode...")
		if err := GlobalBackendManager.LoadFromEnv(); err != nil {
			return fmt.Errorf("failed to load backends: %w", err)
		}
		
		backends := GlobalBackendManager.ListBackends()
		fmt.Printf("Loaded %d backend(s)\n", len(backends))
		for _, backend := range backends {
			fmt.Printf("  - %s (%s) at %s\n", backend.Name, backend.Type, backend.Endpoint)
		}
	} else {
		// Legacy single backend mode
		fmt.Println("Initializing legacy single backend mode...")
		if err := GlobalBackendManager.LoadFromEnv(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not initialize backend: %v\n", err)
			fmt.Fprintf(os.Stderr, "Console will start but backend operations may fail.\n")
			// Don't return error to maintain backward compatibility
		} else {
			fmt.Println("Backend initialized successfully")
		}
	}
	
	return nil
}
