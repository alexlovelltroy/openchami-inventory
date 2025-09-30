package node

import (
	"context"
	"net/http"

	"github.com/openchami/inventory/pkg/policies"
)

// DefaultNodePolicy provides a default implementation for Node authorization
type DefaultNodePolicy struct{}

func (p *DefaultNodePolicy) CanList(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.PolicyDecision {
	// Default: allow authenticated users to list Nodes
	if auth == nil {
		return policies.Deny("authentication required")
	}
	return policies.Allow()
}

func (p *DefaultNodePolicy) CanGet(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceUID string) policies.PolicyDecision {
	// Default: allow authenticated users to get Nodes
	if auth == nil {
		return policies.Deny("authentication required")
	}
	return policies.Allow()
}

func (p *DefaultNodePolicy) CanCreate(ctx context.Context, auth *policies.AuthContext, req *http.Request, resource interface{}) policies.PolicyDecision {
	// Default: allow authenticated users to create Nodes
	if auth == nil {
		return policies.Deny("authentication required")
	}

	// Example custom logic:
	// Check if user can create nodes in their tenant
	if tenantID, ok := policies.GetStringClaim(auth, "tenant_id"); ok {
		// Validate tenant quota, ownership, etc.
		_ = tenantID // Use tenantID for custom validation
	}

	return policies.Allow()
}

func (p *DefaultNodePolicy) CanUpdate(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceUID string, resource interface{}) policies.PolicyDecision {
	// Default: allow authenticated users to update Nodes
	if auth == nil {
		return policies.Deny("authentication required")
	}
	return policies.Allow()
}

func (p *DefaultNodePolicy) CanDelete(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceUID string) policies.PolicyDecision {
	// Default: require node-admin role for deletion
	if auth == nil {
		return policies.Deny("authentication required")
	}

	// Example: only node admins can delete nodes
	if policies.HasRole(auth, "admin") || policies.HasRole(auth, "node-admin") {
		return policies.Allow()
	}

	return policies.Deny("admin role required for Node deletion")
}

// NewDefaultNodePolicy creates a new default Node policy
func NewDefaultNodePolicy() policies.ResourcePolicy {
	return &DefaultNodePolicy{}
}
