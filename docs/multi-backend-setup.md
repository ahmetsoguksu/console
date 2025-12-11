# Multi-Backend S3 Support

This document describes how to configure and use the multi-backend S3 support feature in Console.

## Overview

Console now supports managing multiple S3-compatible backends simultaneously, including:
- MinIO
- Garage (with cluster support)
- Versity Gateway
- Generic S3-compatible services (AWS S3, Wasabi, DigitalOcean Spaces, etc.)

## Features

### 1. Multi-Backend Management
- Configure and manage multiple S3 backends from a single console
- Switch between backends seamlessly
- Health monitoring for all backends
- Cluster support for Garage backend

### 2. Centralized User Management
- **Admin Users**: Manage the console and backends
- **S3 Users**: Access object storage across multiple backends
- User-to-backend mapping
- Granular permissions

### 3. Backend Types

#### MinIO
Standard MinIO backend with full feature support.

#### Garage
Lightweight S3-compatible storage system with cluster support.
- Round-robin load balancing across cluster nodes
- Automatic failover
- Read distribution across nodes

#### Versity Gateway
S3-compatible interface to various storage backends.

#### Generic S3
Support for any S3-compatible service (AWS S3, Wasabi, etc.).

## Configuration

### Environment Variables

#### Legacy Single Backend (Backward Compatible)
```bash
export CONSOLE_MINIO_SERVER=http://localhost:9000
export CONSOLE_MINIO_REGION=us-east-1
```

#### Multi-Backend Configuration
```bash
export CONSOLE_ENABLE_MULTI_BACKEND=true
export CONSOLE_BACKENDS='[
  {
    "id": "minio-primary",
    "name": "Primary MinIO",
    "type": "minio",
    "endpoint": "http://minio1.example.com:9000",
    "region": "us-east-1",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin",
    "enabled": true
  },
  {
    "id": "garage-cluster",
    "name": "Garage Cluster",
    "type": "garage",
    "endpoint": "http://garage1.example.com:3900",
    "region": "garage",
    "accessKey": "GK...",
    "secretKey": "...",
    "clusterMode": true,
    "clusterEndpoints": [
      "http://garage1.example.com:3900",
      "http://garage2.example.com:3900",
      "http://garage3.example.com:3900"
    ],
    "enabled": true
  },
  {
    "id": "versity-backend",
    "name": "Versity Gateway",
    "type": "versity",
    "endpoint": "http://versity.example.com:8080",
    "region": "us-west-1",
    "accessKey": "...",
    "secretKey": "...",
    "enabled": true
  }
]'
export CONSOLE_DEFAULT_BACKEND=minio-primary
```

#### Admin User Configuration
```bash
export CONSOLE_ADMIN_USERNAME=admin
export CONSOLE_ADMIN_PASSWORD=securePassword123
```

### Backend Configuration Schema

```json
{
  "id": "unique-backend-id",           // Unique identifier
  "name": "Display Name",              // Human-readable name
  "type": "minio|garage|versity|s3",   // Backend type
  "endpoint": "http://host:port",      // Backend endpoint
  "region": "us-east-1",               // Region
  "accessKey": "access-key",           // Access key (optional for IAM)
  "secretKey": "secret-key",           // Secret key (optional for IAM)
  "secure": true,                      // Use HTTPS (auto-detected from endpoint)
  "clusterMode": false,                // Enable cluster mode (Garage)
  "clusterEndpoints": [],              // Cluster endpoints (Garage)
  "enabled": true,                     // Enable/disable backend
  "metadata": {                        // Optional metadata
    "location": "datacenter-1"
  }
}
```

## User Management

### Admin Users

Admin users have console-level permissions:
- `manage:backends` - Add, remove, configure backends
- `manage:users` - Create, update, delete users
- `manage:s3users` - Manage S3 users across backends
- `view:backends` - View backend information
- `view:users` - View user information
- `access:backend` - Access specific backends

### S3 Users

S3 users have object storage access:
- Can be mapped to multiple backends
- Each backend mapping includes S3 credentials
- Policies and groups per backend

### Creating Admin User at Startup

The console automatically creates an admin user on first startup if configured:

```bash
export CONSOLE_ADMIN_USERNAME=admin
export CONSOLE_ADMIN_PASSWORD=YourSecurePassword
```

## API Endpoints

### Backend Management

- `GET /api/v1/backends` - List all backends
- `GET /api/v1/backends/{id}` - Get backend details
- `POST /api/v1/backends` - Add new backend
- `DELETE /api/v1/backends/{id}` - Remove backend
- `GET /api/v1/backends/health` - Check backend health
- `PUT /api/v1/backends/{id}/default` - Set as default

### User Management

- `POST /api/v1/users/admin` - Create admin user
- `POST /api/v1/users/s3` - Create S3 user
- `GET /api/v1/users` - List users
- `GET /api/v1/users/{id}` - Get user details
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

### S3 User Configuration

- `POST /api/v1/users/{id}/s3-config` - Create S3 config for backend
- `GET /api/v1/users/{id}/s3-config` - List S3 configs
- `GET /api/v1/users/{id}/s3-config/{backendId}` - Get specific config
- `PUT /api/v1/users/{id}/s3-config/{backendId}` - Update config
- `DELETE /api/v1/users/{id}/s3-config/{backendId}` - Delete config

## Usage Examples

### 1. Configure Multiple Backends

```bash
# Set environment variables
export CONSOLE_BACKENDS='[...]'  # JSON configuration
export CONSOLE_DEFAULT_BACKEND=minio-primary

# Start console
./console server
```

### 2. Create Admin User via API

```bash
curl -X POST http://localhost:9090/api/v1/users/admin \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "SecurePass123!",
    "permissions": ["manage:backends", "manage:users"]
  }'
```

### 3. Add Backend via API

```bash
curl -X POST http://localhost:9090/api/v1/backends \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "New MinIO Server",
    "type": "minio",
    "endpoint": "http://minio2.example.com:9000",
    "region": "us-east-1",
    "accessKey": "admin",
    "secretKey": "password"
  }'
```

### 4. Configure Garage Cluster

```json
{
  "id": "garage-prod",
  "name": "Production Garage Cluster",
  "type": "garage",
  "endpoint": "http://garage-lb.example.com:3900",
  "region": "garage",
  "accessKey": "GK...",
  "secretKey": "...",
  "clusterMode": true,
  "clusterEndpoints": [
    "http://garage1.example.com:3900",
    "http://garage2.example.com:3900",
    "http://garage3.example.com:3900"
  ],
  "enabled": true,
  "metadata": {
    "environment": "production",
    "location": "us-west"
  }
}
```

## Migration from Single Backend

If you're currently using a single MinIO backend, Console maintains backward compatibility:

1. Continue using `CONSOLE_MINIO_SERVER` and `CONSOLE_MINIO_REGION`
2. When ready, migrate to multi-backend:
   - Set `CONSOLE_ENABLE_MULTI_BACKEND=true`
   - Configure `CONSOLE_BACKENDS` with your existing backend
   - Add new backends as needed

## Security Considerations

1. **Credentials Storage**: Backend credentials are stored in memory and never exposed via API responses
2. **Admin Authentication**: Use strong passwords for admin users
3. **Backend Access**: Use granular permissions to control backend access
4. **TLS**: Always use HTTPS endpoints in production
5. **Secrets Management**: Consider using environment variables or secret management systems

## Troubleshooting

### Backend Not Connecting
1. Check endpoint URL format (http:// or https://)
2. Verify network connectivity
3. Validate credentials
4. Check backend health: `GET /api/v1/backends/health`

### Cluster Mode Issues (Garage)
1. Ensure all cluster endpoints are reachable
2. Check that cluster nodes are properly configured
3. Verify consistent credentials across cluster
4. Monitor health status of individual nodes

### User Authentication Fails
1. Verify user exists and is active
2. Check user type (admin vs S3)
3. Ensure backend access permissions
4. Validate S3 user config exists for target backend

## Performance Considerations

### Garage Cluster Mode
- Uses round-robin load balancing for write operations
- Random selection for read operations
- Health checks run periodically
- Failed nodes are bypassed automatically

### Connection Pooling
- Each backend maintains its own connection pool
- Connections are reused across requests
- Idle connections are cleaned up automatically

## Future Enhancements

- [ ] Persistent storage for user data (database)
- [ ] User synchronization across backends
- [ ] Backend failover and redundancy
- [ ] Advanced backend routing (region-based, performance-based)
- [ ] Audit logging for admin operations
- [ ] UI for backend and user management
