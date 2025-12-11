#!/bin/bash
# Example configuration for multi-backend Console setup

# Enable multi-backend mode
export CONSOLE_ENABLE_MULTI_BACKEND=true

# Configure multiple backends
export CONSOLE_BACKENDS='[
  {
    "id": "minio-primary",
    "name": "Primary MinIO Server",
    "type": "minio",
    "endpoint": "http://localhost:9000",
    "region": "us-east-1",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin",
    "enabled": true,
    "metadata": {
      "environment": "production",
      "location": "datacenter-1"
    }
  },
  {
    "id": "garage-cluster",
    "name": "Garage Storage Cluster",
    "type": "garage",
    "endpoint": "http://localhost:3900",
    "region": "garage",
    "accessKey": "GK...",
    "secretKey": "...",
    "clusterMode": true,
    "clusterEndpoints": [
      "http://garage-node-1:3900",
      "http://garage-node-2:3900",
      "http://garage-node-3:3900"
    ],
    "enabled": true,
    "metadata": {
      "environment": "production",
      "location": "datacenter-2"
    }
  },
  {
    "id": "versity-gateway",
    "name": "Versity ScoutFS Gateway",
    "type": "versity",
    "endpoint": "http://localhost:8080",
    "region": "us-west-1",
    "accessKey": "versity-access",
    "secretKey": "versity-secret",
    "enabled": true,
    "metadata": {
      "environment": "testing"
    }
  },
  {
    "id": "aws-s3",
    "name": "AWS S3 Bucket",
    "type": "s3",
    "endpoint": "https://s3.amazonaws.com",
    "region": "us-east-1",
    "accessKey": "AKIAIOSFODNN7EXAMPLE",
    "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "enabled": false,
    "metadata": {
      "environment": "cloud",
      "provider": "aws"
    }
  }
]'

# Set default backend
export CONSOLE_DEFAULT_BACKEND=minio-primary

# Configure admin user (will be created on first startup)
export CONSOLE_ADMIN_USERNAME=admin
export CONSOLE_ADMIN_PASSWORD=SecureAdminPassword123!

# Console server configuration
export CONSOLE_HOSTNAME=0.0.0.0
export CONSOLE_PORT=9090
export CONSOLE_TLS_PORT=9443

# Optional: Configure TLS
# export CONSOLE_SECURE_TLS_REDIRECT=on

# Start the console
echo "Starting Console with multi-backend configuration..."
echo "Admin user: $CONSOLE_ADMIN_USERNAME"
echo "Console URL: http://localhost:$CONSOLE_PORT"
echo ""
echo "Available backends:"
echo "  - minio-primary (MinIO)"
echo "  - garage-cluster (Garage with 3 nodes)"
echo "  - versity-gateway (Versity)"
echo "  - aws-s3 (AWS S3 - disabled)"
echo ""

# Uncomment to start
# ./console server
