package bmc

import (
	"context"
	"net/http"

	"github.com/openchami/inventory/pkg/policies"
)

// DefaultBMCPolicy provides a default implementation for BMC authorization
type DefaultBMCPolicy struct{}

func (p *DefaultBMCPolicy) CanList(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.PolicyDecision {
	// Default: allow authenticated users to list BMCs
	if auth == nil {
		return policies.Deny("authentication required")
	}
	return policies.Allow()
}

func (p *DefaultBMCPolicy) CanGet(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceUID string) policies.PolicyDecision {
	// Default: allow authenticated users to get BMCs
	if auth == nil {
		return policies.Deny("authentication required")
	}
	return policies.Allow()
}

func (p *DefaultBMCPolicy) CanCreate(ctx context.Context, auth *policies.AuthContext, req *http.Request, resource interface{}) policies.PolicyDecision {
	// Default: allow authenticated users to create BMCs
	if auth == nil {
		return policies.Deny("authentication required")
	}

	// You can add custom logic here, for example:
	// - Check if user has "bmc-admin" role
	// - Validate tenant ownership
	// - Check resource quotas

	return policies.Allow()
}

func (p *DefaultBMCPolicy) CanUpdate(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceUID string, resource interface{}) policies.PolicyDecision {
	// Default: allow authenticated users to update BMCs
	if auth == nil {
		return policies.Deny("authentication required")
	}
	return policies.Allow()
}

func (p *DefaultBMCPolicy) CanDelete(ctx context.Context, auth *policies.AuthContext, req *http.Request, resourceUID string) policies.PolicyDecision {
	// Default: require admin role for deletion
	if auth == nil {
		return policies.Deny("authentication required")
	}

	// Example: only admins can delete BMCs
	if policies.HasRole(auth, "admin") || policies.HasRole(auth, "bmc-admin") {
		return policies.Allow()
	}

	return policies.Deny("admin role required for BMC deletion")
}

// NewDefaultBMCPolicy creates a new default BMC policy
func NewDefaultBMCPolicy() policies.ResourcePolicy {
	return &DefaultBMCPolicy{}
}
