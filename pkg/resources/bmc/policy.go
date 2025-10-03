package bmc

import (
	"context"
	"net/http"

	"github.com/alexlovelltroy/fabrica/pkg/policy"
)

// DefaultBMCPolicy provides a default implementation for BMC authorization
type DefaultBMCPolicy struct{}

func (p *DefaultBMCPolicy) CanList(ctx context.Context, auth *policy.AuthContext, req *http.Request) policy.PolicyDecision {
	// Default: allow authenticated users to list BMCs
	if auth == nil {
		return policy.Deny("authentication required")
	}
	return policy.Allow()
}

func (p *DefaultBMCPolicy) CanGet(ctx context.Context, auth *policy.AuthContext, req *http.Request, resourceUID string) policy.PolicyDecision {
	// Default: allow authenticated users to get BMCs
	if auth == nil {
		return policy.Deny("authentication required")
	}
	return policy.Allow()
}

func (p *DefaultBMCPolicy) CanCreate(ctx context.Context, auth *policy.AuthContext, req *http.Request, resource interface{}) policy.PolicyDecision {
	// Default: allow authenticated users to create BMCs
	if auth == nil {
		return policy.Deny("authentication required")
	}

	// You can add custom logic here, for example:
	// - Check if user has "bmc-admin" role
	// - Validate tenant ownership
	// - Check resource quotas

	return policy.Allow()
}

func (p *DefaultBMCPolicy) CanUpdate(ctx context.Context, auth *policy.AuthContext, req *http.Request, resourceUID string, resource interface{}) policy.PolicyDecision {
	// Default: allow authenticated users to update BMCs
	if auth == nil {
		return policy.Deny("authentication required")
	}
	return policy.Allow()
}

func (p *DefaultBMCPolicy) CanDelete(ctx context.Context, auth *policy.AuthContext, req *http.Request, resourceUID string) policy.PolicyDecision {
	// Default: require admin role for deletion
	if auth == nil {
		return policy.Deny("authentication required")
	}

	// Example: only admins can delete BMCs
	if policy.HasRole(auth, "admin") || policy.HasRole(auth, "bmc-admin") {
		return policy.Allow()
	}

	return policy.Deny("admin role required for BMC deletion")
}

// NewDefaultBMCPolicy creates a new default BMC policy
func NewDefaultBMCPolicy() policy.ResourcePolicy {
	return &DefaultBMCPolicy{}
}
