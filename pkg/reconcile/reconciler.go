// Package reconcile provides inventory-specific reconciliation utilities.
//
// This package re-exports the fabrica reconciliation framework and provides
// inventory-specific reconciler implementations.
package reconcile

import (
	"github.com/alexlovelltroy/fabrica/pkg/reconcile"
)

// Re-export fabrica reconcile types for backwards compatibility
type (
	Reconciler        = reconcile.Reconciler
	Result            = reconcile.Result
	ClientInterface   = reconcile.ClientInterface
	BaseReconciler    = reconcile.BaseReconciler
	Logger            = reconcile.Logger
	Controller        = reconcile.Controller
	ReconcileRequest  = reconcile.ReconcileRequest
	WorkQueue         = reconcile.WorkQueue
	RateLimiter       = reconcile.RateLimiter
)

// Re-export fabrica reconcile functions
var (
	NewDefaultLogger                = reconcile.NewDefaultLogger
	NewController                   = reconcile.NewController
	NewWorkQueue                    = reconcile.NewWorkQueue
	NewRateLimitedWorkQueue         = reconcile.NewRateLimitedWorkQueue
	NewExponentialBackoffRateLimiter = reconcile.NewExponentialBackoffRateLimiter
)
