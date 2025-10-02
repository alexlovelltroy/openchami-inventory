# Authentication and Authorization

Guide to authentication, authorization, and policy management in OpenCHAMI Inventory.

## Table of Contents

- [Overview](#overview)
- [Authentication Modes](#authentication-modes)
- [Policy System](#policy-system)
- [Testing Mode](#testing-mode)
- [Production Setup](#production-setup)
- [Custom Policies](#custom-policies)
- [BMC Authentication Methods](#bmc-authentication-methods)

## Overview

OpenCHAMI Inventory supports flexible authentication and authorization through:

1. **Authentication**: Verifying user identity (JWT tokens)
2. **Authorization**: Controlling access to resources (policies)
3. **Testing Mode**: Permissive policy for development (`--disable-auth`)

### Architecture

```
Request → JWT Middleware → Policy Check → Handler → Storage
          (AuthContext)      (Allow/Deny)
```

## Authentication Modes

### Production Mode (Default)

Server requires JWT authentication for all requests.

**Start server:**
```bash
./bin/server --port 9999
```

**Make authenticated request:**
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:9999/bmcs
```

### Testing Mode

Server uses permissive policy that allows all operations without authentication.

**⚠️ WARNING: Never use in production!**

**Start server:**
```bash
./bin/server --port 9999 --disable-auth
```

**Make unauthenticated request:**
```bash
curl http://localhost:9999/bmcs
```

**Environment variable:**
```bash
export INVENTORY_DISABLE_AUTH=true
./bin/server --port 9999
```

## Policy System

### Policy Interface

All resources implement the `ResourcePolicy` interface:

```go
type ResourcePolicy interface {
    CanList(ctx context.Context, auth *AuthContext, req *http.Request) PolicyDecision
    CanGet(ctx context.Context, auth *AuthContext, req *http.Request, resourceID string) PolicyDecision
    CanCreate(ctx context.Context, auth *AuthContext, req *http.Request) PolicyDecision
    CanUpdate(ctx context.Context, auth *AuthContext, req *http.Request, resourceID string) PolicyDecision
    CanDelete(ctx context.Context, auth *AuthContext, req *http.Request, resourceID string) PolicyDecision
}
```

### Policy Decisions

Policies return one of:
- `Allow()` - Grant access
- `Deny(reason)` - Deny access with reason

### AuthContext

Authentication information passed to policies:

```go
type AuthContext struct {
    UserID   string
    Username string
    Email    string
    Roles    []string
    Groups   []string
    Claims   map[string]interface{}
}
```

## Testing Mode

### Permissive Policy

When `--disable-auth` is set, the server uses `PermissivePolicy`:

**Location:** `pkg/policies/permissive.go`

```go
type PermissivePolicy struct{}

func (p *PermissivePolicy) CanList(ctx context.Context, auth *AuthContext, req *http.Request) PolicyDecision {
    return Allow()
}

func (p *PermissivePolicy) CanGet(ctx context.Context, auth *AuthContext, req *http.Request, resourceID string) PolicyDecision {
    return Allow()
}

// ... all methods return Allow()
```

### How It Works

**In `cmd/server/main.go`:**

```go
if disableAuth {
    log.Warn("Authentication disabled - using permissive policy (TESTING ONLY)")
    
    // Register permissive policy for all resources
    policyRegistry.RegisterPolicy("BMC", policies.NewPermissivePolicy())
    policyRegistry.RegisterPolicy("Node", policies.NewPermissivePolicy())
    policyRegistry.RegisterPolicy("FRU", policies.NewPermissivePolicy())
    policyRegistry.RegisterPolicy("BootConfiguration", policies.NewPermissivePolicy())
} else {
    // Use default policies (require authentication)
    policyRegistry.RegisterPolicy("BMC", bmc.NewDefaultBMCPolicy())
    policyRegistry.RegisterPolicy("Node", node.NewDefaultNodePolicy())
    policyRegistry.RegisterPolicy("FRU", fru.NewDefaultFRUPolicy())
    policyRegistry.RegisterPolicy("BootConfiguration", boot.NewDefaultBootPolicy())
}
```

### When to Use Testing Mode

✅ **Appropriate:**
- Local development
- Integration testing
- Demo scripts
- CI/CD test environments
- Debugging API issues

❌ **Never use:**
- Production deployments
- Public-facing instances
- Multi-user environments
- Any system with real data

## Production Setup

### 1. Configure JWT Middleware

**In `cmd/server/main.go`:**

```go
// Add tokensmith middleware
import "github.com/openchami/tokensmith/middleware"

server := fuego.NewServer(
    fuego.WithAddr(fmt.Sprintf("%s:%d", host, port)),
    fuego.WithMiddleware(middleware.TokensmithMiddleware(tokensmithConfig)),
)
```

### 2. Create Default Policies

**Example: `pkg/resources/bmc/policy.go`:**

```go
package bmc

import (
    "context"
    "net/http"
    "github.com/openchami/inventory/pkg/policies"
)

type DefaultBMCPolicy struct{}

func NewDefaultBMCPolicy() *DefaultBMCPolicy {
    return &DefaultBMCPolicy{}
}

func (p *DefaultBMCPolicy) CanList(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.PolicyDecision {
    // Require authentication
    if auth == nil {
        return policies.Deny("authentication required")
    }
    
    // Allow all authenticated users to list BMCs
    return policies.Allow()
}

func (p *DefaultBMCPolicy) CanGet(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceID string) policies.PolicyDecision {
    if auth == nil {
        return policies.Deny("authentication required")
    }
    
    // Allow all authenticated users to get BMCs
    return policies.Allow()
}

func (p *DefaultBMCPolicy) CanCreate(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.PolicyDecision {
    if auth == nil {
        return policies.Deny("authentication required")
    }
    
    // Only admins can create BMCs
    if policies.HasRole(auth, "admin") {
        return policies.Allow()
    }
    
    return policies.Deny("admin role required to create BMCs")
}

func (p *DefaultBMCPolicy) CanUpdate(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceID string) policies.PolicyDecision {
    if auth == nil {
        return policies.Deny("authentication required")
    }
    
    // Admins and operators can update BMCs
    if policies.HasRole(auth, "admin") || policies.HasRole(auth, "operator") {
        return policies.Allow()
    }
    
    return policies.Deny("admin or operator role required")
}

func (p *DefaultBMCPolicy) CanDelete(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceID string) policies.PolicyDecision {
    if auth == nil {
        return policies.Deny("authentication required")
    }
    
    // Only admins can delete BMCs
    if policies.HasRole(auth, "admin") {
        return policies.Allow()
    }
    
    return policies.Deny("admin role required to delete BMCs")
}
```

### 3. Register Policies

**In `cmd/server/main.go`:**

```go
// Production mode - use default policies
policyRegistry.RegisterPolicy("BMC", bmc.NewDefaultBMCPolicy())
policyRegistry.RegisterPolicy("Node", node.NewDefaultNodePolicy())
policyRegistry.RegisterPolicy("FRU", fru.NewDefaultFRUPolicy())
policyRegistry.RegisterPolicy("BootConfiguration", boot.NewDefaultBootPolicy())
```

### 4. Test Authentication

```bash
# Without token (should fail)
curl http://localhost:9999/bmcs
# {"error": "authentication required"}

# With valid token
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:9999/bmcs
# [{"name": "bmc-001", ...}]
```

## Custom Policies

### Role-Based Access Control

```go
type RBACPolicy struct {
    allowedRoles map[string][]string // operation -> roles
}

func NewRBACPolicy() *RBACPolicy {
    return &RBACPolicy{
        allowedRoles: map[string][]string{
            "list":   {"viewer", "operator", "admin"},
            "get":    {"viewer", "operator", "admin"},
            "create": {"operator", "admin"},
            "update": {"operator", "admin"},
            "delete": {"admin"},
        },
    }
}

func (p *RBACPolicy) CanList(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.PolicyDecision {
    return p.checkRoles(auth, "list")
}

func (p *RBACPolicy) checkRoles(auth *policies.AuthContext, operation string) policies.PolicyDecision {
    if auth == nil {
        return policies.Deny("authentication required")
    }
    
    allowedRoles := p.allowedRoles[operation]
    for _, role := range auth.Roles {
        for _, allowed := range allowedRoles {
            if role == allowed {
                return policies.Allow()
            }
        }
    }
    
    return policies.Deny(fmt.Sprintf("requires one of: %v", allowedRoles))
}
```

### Attribute-Based Access Control

```go
type ABACPolicy struct{}

func (p *ABACPolicy) CanGet(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceID string) policies.PolicyDecision {
    if auth == nil {
        return policies.Deny("authentication required")
    }
    
    // Load resource to check attributes
    resource, err := storage.LoadBMC(ctx, resourceID)
    if err != nil {
        return policies.Deny("resource not found")
    }
    
    // Allow if user's datacenter matches resource datacenter
    userDC := auth.Claims["datacenter"].(string)
    resourceDC := resource.Metadata.Labels["datacenter"]
    
    if userDC == resourceDC {
        return policies.Allow()
    }
    
    // Or if user is admin
    if policies.HasRole(auth, "admin") {
        return policies.Allow()
    }
    
    return policies.Deny("access restricted to same datacenter")
}
```

### Time-Based Access Control

```go
func (p *TimeBasedPolicy) CanCreate(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.PolicyDecision {
    if auth == nil {
        return policies.Deny("authentication required")
    }
    
    // Only allow creates during maintenance window
    now := time.Now()
    hour := now.Hour()
    
    // Maintenance window: 2 AM - 4 AM
    if hour >= 2 && hour < 4 {
        return policies.Allow()
    }
    
    // Unless user is admin
    if policies.HasRole(auth, "admin") {
        return policies.Allow()
    }
    
    return policies.Deny("creates only allowed during maintenance window (2-4 AM)")
}
```

## BMC Authentication Methods

The Inventory system tracks BMC-level authentication (how to connect to BMCs), separate from API-level authentication (who can access the inventory).

### v1 BMC Authentication

Simple username/password:

```json
{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}
```

### v2beta1 BMC Authentication

Structured authentication with multiple methods:

#### Basic Authentication

```json
{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "type": "Redfish",
  "authentication": {
    "method": "basic",
    "basic": {
      "username": "admin",
      "password": "changeme"
    }
  }
}
```

#### Client Certificate (mTLS)

```json
{
  "name": "node002-bmc",
  "address": "https://10.1.1.101",
  "type": "Redfish",
  "authentication": {
    "method": "client-cert",
    "clientCert": {
      "certificateRef": "secret://inventory/bmc-cert",
      "keyRef": "secret://inventory/bmc-key",
      "caBundle": "-----BEGIN CERTIFICATE-----\n..."
    }
  }
}
```

#### OIDC/OAuth2

```json
{
  "name": "node003-bmc",
  "address": "https://10.1.1.102",
  "type": "Redfish",
  "authentication": {
    "method": "oidc",
    "oidc": {
      "issuerUrl": "https://auth.example.com",
      "clientId": "bmc-inventory",
      "clientSecret": "secret123",
      "scopes": ["redfish"]
    }
  }
}
```

## Policy Helpers

**Location:** `pkg/policies/helpers.go`

```go
// Check if user has role
func HasRole(auth *AuthContext, role string) bool {
    for _, r := range auth.Roles {
        if r == role {
            return true
        }
    }
    return false
}

// Check if user is in group
func HasGroup(auth *AuthContext, group string) bool {
    for _, g := range auth.Groups {
        if g == group {
            return true
        }
    }
    return false
}

// Check if user has any of the roles
func HasAnyRole(auth *AuthContext, roles ...string) bool {
    for _, role := range roles {
        if HasRole(auth, role) {
            return true
        }
    }
    return false
}

// Extract claim from auth context
func GetClaim(auth *AuthContext, key string) (interface{}, bool) {
    val, ok := auth.Claims[key]
    return val, ok
}
```

## Security Best Practices

### 1. Never Use --disable-auth in Production

```bash
# ❌ NEVER DO THIS IN PRODUCTION
./bin/server --disable-auth

# ✅ Production
./bin/server
```

### 2. Use HTTPS in Production

```bash
# Terminate TLS at load balancer or reverse proxy
# Example nginx config:
server {
    listen 443 ssl;
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:9999;
        proxy_set_header Authorization $http_authorization;
    }
}
```

### 3. Implement Least Privilege

```go
// Default: deny
func (p *MyPolicy) CanDelete(...) PolicyDecision {
    if auth == nil {
        return Deny("authentication required")
    }
    
    // Only allow specific role
    if HasRole(auth, "admin") {
        return Allow()
    }
    
    // Deny by default
    return Deny("insufficient permissions")
}
```

### 4. Audit Access

```go
func (p *AuditPolicy) CanDelete(ctx context.Context, auth *AuthContext, req *http.Request, resourceID string) PolicyDecision {
    decision := p.basePolicy.CanDelete(ctx, auth, req, resourceID)
    
    // Log all delete attempts
    log.WithFields(log.Fields{
        "user":       auth.Username,
        "resource":   resourceID,
        "operation":  "delete",
        "decision":   decision.Allowed,
        "reason":     decision.Reason,
        "ip":         req.RemoteAddr,
    }).Info("access attempt")
    
    return decision
}
```

### 5. Rotate Secrets

- Rotate BMC passwords regularly
- Use short-lived JWT tokens
- Store secrets in secret management system (Vault, etc.)

## Troubleshooting

See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md#authentication-errors) for common authentication issues.

## See Also

- [User Guide](./USER-GUIDE.md) - Usage guide
- [Troubleshooting](./TROUBLESHOOTING.md) - Common issues
- [Development Guide](../developer/DEVELOPMENT.md) - Policy development
