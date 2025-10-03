package node

import (
	"context"
	"net/http"

	"github.com/alexlovelltroy/fabrica/pkg/policy"
)

// DefaultNodePolicy provides a default implementation for Node authorization
type DefaultNodePolicy struct{}

func (p *DefaultNodePolicy) CanList(ctx context.Context, auth *policy.AuthContext, req *http.Request) policy.PolicyDecision {
	// Default: allow authenticated users to list Nodes
	if auth == nil {
		return policy.Deny("authentication required")
	}
	return policy.Allow()
}

func (p *DefaultNodePolicy) CanGet(ctx context.Context, auth *policy.AuthContext, req *http.Request, resourceUID string) policy.PolicyDecision {
	// Default: allow authenticated users to get Nodes
	if auth == nil {
		return policy.Deny("authentication required")
	}
	return policy.Allow()
}

func (p *DefaultNodePolicy) CanCreate(ctx context.Context, auth *policy.AuthContext, req *http.Request, resource interface{}) policy.PolicyDecision {
	// Default: allow authenticated users to create Nodes
	if auth == nil {
		return policy.Deny("authentication required")
	}

	// Example custom logic:
	// Check if user can create nodes in their tenant
	if tenantID, ok := policy.GetStringClaim(auth, "tenant_id"); ok {
		// Validate tenant quota, ownership, etc.
		_ = tenantID // Use tenantID for custom validation
	}

	return policy.Allow()
}

func (p *DefaultNodePolicy) CanUpdate(ctx context.Context, auth *policy.AuthContext, req *http.Request, resourceUID string, resource interface{}) policy.PolicyDecision {
	// Default: allow authenticated users to update Nodes
	if auth == nil {
		return policy.Deny("authentication required")
	}
	return policy.Allow()
}

func (p *DefaultNodePolicy) CanDelete(ctx context.Context, auth *policy.AuthContext, req *http.Request, resourceUID string) policy.PolicyDecision {
	// Default: require node-admin role for deletion
	if auth == nil {
		return policy.Deny("authentication required")
	}

	// Example: only node admins can delete nodes
	if policy.HasRole(auth, "admin") || policy.HasRole(auth, "node-admin") {
		return policy.Allow()
	}

	return policy.Deny("admin role required for Node deletion")
}

// NewDefaultNodePolicy creates a new default Node policy
func NewDefaultNodePolicy() policy.ResourcePolicy {
	return &DefaultNodePolicy{}
}
