# Framework Extraction & Refactoring Plan

## Executive Summary

This plan outlines the extraction of generic inventory management components from `openchami-inventory` into a reusable framework (`fabrica`), followed by refactoring the HPC-specific implementation to use the new framework.

**Goals:**
- Create a reusable, domain-agnostic inventory management framework
- Maintain backward compatibility with existing HPC inventory functionality
- Enable other projects to build inventory systems using the same foundation
- Improve maintainability through clear separation of concerns

---

## Phase 1: Create Framework Repository

### 1.1 Repository Setup

**Repository Name:** `fabrica` (or `go-inventory-framework`)
**Module Path:** `github.com/alovelltroy/fabrica`
**License:** MIT (same as openchami-inventory)

#### Tasks:
- [ ] Create new GitHub repository
- [ ] Initialize Go module: `go mod init github.com/alexlovelltroy/fabrica`
- [ ] Set up basic repository structure
- [ ] Configure CI/CD (GitHub Actions)
- [ ] Add code of conduct, contributing guidelines
- [ ] Create initial README with framework overview

**Deliverable:** Empty repository with proper structure and governance

---

### 1.2 Core Framework Package Structure

```
fabrica/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── release.yml
├── pkg/
│   ├── resource/          # Core resource model
│   ├── versioning/        # Multi-version schema system
│   ├── codegen/           # Code generation engine
│   ├── storage/           # Storage abstraction
│   ├── policy/            # Authorization framework
│   └── server/            # Generic server utilities
├── templates/             # Code generation templates
│   ├── handlers.go.tmpl
│   ├── storage.go.tmpl
│   ├── client.go.tmpl
│   ├── client-cmd.go.tmpl
│   ├── models.go.tmpl
│   ├── routes.go.tmpl
│   └── policies.go.tmpl
├── examples/              # Example implementations
│   ├── simple-inventory/  # Minimal example
│   └── blog-cms/          # Different domain example
├── docs/                  # Framework documentation
│   ├── README.md
│   ├── getting-started.md
│   ├── architecture.md
│   ├── resource-model.md
│   ├── versioning.md
│   ├── code-generation.md
│   ├── storage.md
│   ├── authorization.md
│   └── api-reference.md
└── tests/
    └── integration/
```

---

### 1.3 Extract Core Components

#### 1.3.1 Resource Model Package (`pkg/resource/`)

**Source Files from openchami-inventory:**
- `pkg/resources/resource.go` → `pkg/resource/resource.go`
- `pkg/resources/metadata.go` → `pkg/resource/metadata.go`
- `pkg/resources/conditions.go` → `pkg/resource/conditions.go`

**Changes Required:**
- Update package name from `resources` to `resource`
- Make all components fully generic (no HPC-specific references)
- Add comprehensive godoc documentation
- Include usage examples in documentation

**Key Interfaces:**
```go
// Core resource interface
type Resource interface {
    GetUID() string
    SetUID(string)
    GetName() string
    SetName(string)
    GetLabels() map[string]string
    SetLabel(key, value string)
    // ... other metadata methods
}

// Base resource struct that implementations can embed
type BaseResource struct {
    APIVersion    string
    Kind          string
    SchemaVersion string
    Metadata      Metadata
    Spec          interface{}
    Status        interface{}
}
```

#### 1.3.2 Versioning Package (`pkg/versioning/`)

**Source Files from openchami-inventory:**
- `pkg/versioning/version_registry.go` → `pkg/versioning/registry.go`
- `pkg/versioning/version_middleware.go` → `pkg/versioning/middleware.go`
- `pkg/versioning/version_test.go` → `pkg/versioning/registry_test.go`
- `pkg/versioning/middleware_test.go` → `pkg/versioning/middleware_test.go`

**Changes Required:**
- Remove any HPC-specific version examples from tests
- Add generic examples to tests
- Enhance documentation with migration guides
- Add version compatibility matrix documentation

**Key Features:**
- Schema version registry
- Automatic version conversion
- HTTP middleware for version negotiation
- Support for alpha/beta/stable stability levels

#### 1.3.3 Storage Package (`pkg/storage/`)

**Source Files from openchami-inventory:**
- `internal/storage/interfaces.go` → `pkg/storage/interfaces.go`
- `internal/storage/file_backend.go` → `pkg/storage/file.go`

**Changes Required:**
- Move from `internal` to `pkg` (make public)
- Add interface documentation with implementation guide
- Add pluggable backend examples
- Remove HPC-specific resource type references from convenience functions

**Key Interfaces:**
```go
// Main storage backend interface
type Backend interface {
    LoadAll(ctx context.Context, resourceType string) ([]json.RawMessage, error)
    Load(ctx context.Context, resourceType, uid string) (json.RawMessage, error)
    Save(ctx context.Context, resourceType, uid string, data json.RawMessage) error
    Delete(ctx context.Context, resourceType, uid string) error
    // ... version-aware methods
}

// Type-safe storage wrapper
type ResourceStorage[T Resource] interface {
    LoadAll(ctx context.Context) ([]T, error)
    Load(ctx context.Context, uid string) (T, error)
    Save(ctx context.Context, resource T) error
    Delete(ctx context.Context, uid string) error
}
```

#### 1.3.4 Policy Package (`pkg/policy/`)

**Source Files from openchami-inventory:**
- `pkg/policies/registry.go` → `pkg/policy/registry.go`
- `pkg/policies/permissive.go` → `pkg/policy/permissive.go`

**Changes Required:**
- Rename package from `policies` to `policy` (singular)
- Remove resource-specific policy implementations (move to examples)
- Add more authorization patterns (RBAC, ABAC examples)
- Document policy implementation guide

**Key Interfaces:**
```go
// Core authorization interface
type ResourcePolicy interface {
    CanList(ctx context.Context, auth *AuthContext, req *http.Request) Decision
    CanGet(ctx context.Context, auth *AuthContext, req *http.Request, uid string) Decision
    CanCreate(ctx context.Context, auth *AuthContext, req *http.Request, resource interface{}) Decision
    CanUpdate(ctx context.Context, auth *AuthContext, req *http.Request, uid string, resource interface{}) Decision
    CanDelete(ctx context.Context, auth *AuthContext, req *http.Request, uid string) Decision
}

// Policy registry for multi-resource systems
type Registry struct {
    // ...
}
```

#### 1.3.5 Code Generation Package (`pkg/codegen/`)

**Source Files from openchami-inventory:**
- `pkg/codegen/generator.go` → `pkg/codegen/generator.go`
- `pkg/codegen/templates/*.tmpl` → `templates/*.tmpl`

**Changes Required:**
- Make templates fully parameterized (remove hardcoded resource types)
- Add template customization guide
- Support custom template directories
- Add CLI tool for running codegen: `cmd/inventory-codegen/`

**Key Features:**
- Template-based code generation
- Generate: handlers, storage, client, CLI, models, routes
- Pluggable template system
- Resource metadata extraction

#### 1.3.6 Server Utilities Package (`pkg/server/`)

**New Package - Extracted patterns from openchami-inventory:**

**Purpose:** Common server setup utilities

**Components:**
- Server configuration structure
- Middleware helpers
- Health check endpoints
- Version info endpoints
- CORS configuration helpers
- Logging setup

**Example:**
```go
package server

type Config struct {
    Host         string
    Port         int
    CORSEnabled  bool
    CORSOrigins  []string
    StoragePath  string
    DisableAuth  bool
}

func NewServer(cfg Config) *fuego.Server {
    // Common server setup
}

func AddHealthCheck(s *fuego.Server) {
    // Standard health endpoint
}
```

---

### 1.4 Code Generation Templates

#### Move and Enhance Templates:

All templates from `pkg/codegen/templates/` → `templates/`

**Templates to migrate:**
- `handlers.go.tmpl` - REST API CRUD handlers
- `storage.go.tmpl` - Storage layer operations
- `client.go.tmpl` - HTTP client library
- `client-cmd.go.tmpl` - CLI application
- `models.go.tmpl` - Request/response models
- `routes.go.tmpl` - Route registration
- `policies.go.tmpl` - Policy scaffolding

**Enhancements:**
- Remove hardcoded import paths (use template variables)
- Add template documentation headers
- Create template customization guide
- Add example custom templates

---

### 1.5 Framework Documentation

#### Core Documentation Files:

**`README.md`** (Repository root)
```markdown
# Inventory Framework

A production-ready Go framework for building REST API-based inventory management systems with multi-version schema support, automatic code generation, and pluggable backends.

## Features
- Kubernetes-style resource model
- Multi-version schema support with automatic conversion
- Template-based code generation
- Type-safe storage abstraction
- Flexible authorization framework
- REST API, HTTP client, and CLI generation

## Quick Start
[30-second example]

## Use Cases
- Hardware inventory (HPC, datacenter equipment)
- Asset management
- Cloud resource tracking
- IoT device management
- Any domain requiring versioned resource tracking
```

**`docs/getting-started.md`**
- Installation
- First resource definition
- Generate code
- Run server
- Use client/CLI

**`docs/architecture.md`**
- Framework design principles
- Component overview
- Data flow diagrams
- Extension points

**`docs/resource-model.md`**
- Resource structure (APIVersion, Kind, Metadata, Spec, Status)
- UID generation strategies
- Labels and annotations
- Resource lifecycle

**`docs/versioning.md`**
- Multi-version schema design
- Version registration
- Conversion patterns
- Migration strategies
- Version negotiation

**`docs/code-generation.md`**
- How code generation works
- Template system
- Customizing templates
- Integration into build process

**`docs/storage.md`**
- Storage backend interface
- Implementing custom backends
- File backend
- Database backend patterns
- Caching strategies

**`docs/authorization.md`**
- Policy framework
- Implementing custom policies
- RBAC patterns
- ABAC patterns
- JWT integration

**`docs/api-reference.md`**
- Complete API documentation (godoc)
- Package references
- Interface specifications

---

### 1.6 Example Implementations

#### 1.6.1 Simple Inventory Example (`examples/simple-inventory/`)

**Purpose:** Minimal working example to demonstrate framework usage

**Resource:** Simple "Device" resource
```go
type Device struct {
    resource.BaseResource
    Spec   DeviceSpec
    Status DeviceStatus
}

type DeviceSpec struct {
    Name     string
    Type     string
    Location string
}
```

**Complete working system:**
- Resource definition
- Code generation
- REST API server
- CLI client
- Storage setup

#### 1.6.2 Different Domain Example (`examples/blog-cms/`)

**Purpose:** Show framework versatility in non-hardware domain

**Resources:**
- Posts
- Authors
- Comments

**Demonstrates:**
- Multiple resource types
- Cross-resource relationships
- Custom policies
- Version migration

---

### 1.7 Testing & Quality Assurance

#### Unit Tests
- All packages have >80% coverage
- Interface compliance tests
- Version conversion tests
- Storage backend tests

#### Integration Tests
- End-to-end code generation
- Server startup and routing
- Version negotiation flows
- Storage operations
- Authorization flows

#### Benchmarks
- Code generation performance
- Storage backend performance
- Version conversion overhead
- Serialization performance

#### CI/CD
```yaml
# .github/workflows/test.yml
- Run tests on Go 1.24+
- Run integration tests
- Check test coverage
- Lint code (golangci-lint)
- Verify examples compile
- Test template generation
```

---

### 1.8 Dependencies Management

**Minimal Dependencies:**
- `github.com/go-fuego/fuego` - HTTP framework
- `github.com/spf13/cobra` - CLI framework (examples)
- `github.com/spf13/viper` - Configuration (examples)
- `golang.org/x/text` - Text processing

**Consider:**
- Making web framework pluggable (support alternative to fuego)
- Optional dependencies for examples only

---

### 1.9 Release Strategy

**Initial Release: v0.1.0**
- Core framework functionality
- Basic documentation
- Simple example
- Alpha quality - breaking changes expected

**v0.5.0 Goal:**
- Complete documentation
- Multiple examples
- Beta quality - API stabilizing

**v1.0.0 Goal:**
- API stable
- Production ready
- Complete test coverage
- Comprehensive documentation

---

## Phase 2: Refactor HPC Inventory Implementation

### 2.1 Update openchami-inventory Dependencies

#### Update go.mod
```go
module github.com/openchami/inventory

go 1.24.7

require (
    github.com/openchami/inventory-framework v0.1.0
    github.com/stmcginnis/gofish v0.20.0  // HPC-specific
    // ... other HPC-specific dependencies
)
```

#### Remove Duplicated Code
- Delete `pkg/resources/resource.go` (use framework)
- Delete `pkg/resources/metadata.go` (use framework)
- Delete `pkg/resources/conditions.go` (use framework)
- Delete `pkg/versioning/` (use framework)
- Delete `pkg/codegen/` (use framework)
- Delete `internal/storage/interfaces.go` (use framework)
- Delete `pkg/policies/registry.go` (use framework)

---

### 2.2 Restructure HPC-Specific Code

**New Structure:**
```
openchami-inventory/
├── go.mod
├── pkg/
│   ├── resources/           # HPC resource definitions only
│   │   ├── bmc/
│   │   │   ├── bmc.go      # Uses framework.BaseResource
│   │   │   ├── policy.go
│   │   │   └── v2beta1/
│   │   ├── node/
│   │   ├── fru/
│   │   └── boot/
│   ├── crawler/             # HPC-specific hardware discovery
│   └── policies/            # HPC-specific policy implementations
├── cmd/
│   ├── server/              # HPC inventory server
│   ├── inventory-cli/       # HPC inventory CLI
│   ├── crawler/             # HPC crawler tool
│   └── populate-bmcs/       # HPC testing tool
└── docs/                    # HPC-specific documentation
```

---

### 2.3 Update Resource Definitions

#### Example: BMC Resource Update

**Before (openchami-inventory):**
```go
package bmc

import "github.com/openchami/inventory/pkg/resources"

type BMC struct {
    resources.Resource
    Spec   BMCSpec
    Status BMCStatus
}
```

**After (using framework):**
```go
package bmc

import "github.com/openchami/inventory-framework/pkg/resource"

type BMC struct {
    resource.BaseResource
    Spec   BMCSpec
    Status BMCStatus
}
```

#### Update All Resources:
- [ ] Update `pkg/resources/bmc/bmc.go`
- [ ] Update `pkg/resources/node/node.go`
- [ ] Update `pkg/resources/fru/fru.go`
- [ ] Update `pkg/resources/boot/boot.go`
- [ ] Update version-specific resources (v2beta1)

---

### 2.4 Update Import Paths

**Global Search & Replace:**

| Old Import | New Import |
|------------|------------|
| `github.com/openchami/inventory/pkg/resources` | `github.com/openchami/inventory-framework/pkg/resource` |
| `github.com/openchami/inventory/pkg/versioning` | `github.com/openchami/inventory-framework/pkg/versioning` |
| `github.com/openchami/inventory/pkg/codegen` | `github.com/openchami/inventory-framework/pkg/codegen` |
| `github.com/openchami/inventory/internal/storage` | `github.com/openchami/inventory-framework/pkg/storage` |
| `github.com/openchami/inventory/pkg/policies` | `github.com/openchami/inventory-framework/pkg/policy` |

**Files to update:**
- All `.go` files in `pkg/resources/`
- All files in `cmd/`
- Code generation configurations
- Test files

---

### 2.5 Update Code Generation

#### Update cmd/codegen/main.go

**Before:**
```go
import "github.com/openchami/inventory/pkg/codegen"
```

**After:**
```go
import "github.com/openchami/inventory-framework/pkg/codegen"
```

#### Update Template References
- Remove local templates (use framework templates)
- Configure custom template path if needed
- Update Makefile code generation targets

---

### 2.6 Update Server Implementation

#### Update cmd/server/main.go

**Before:**
```go
import (
    "github.com/openchami/inventory/pkg/versioning"
    "github.com/openchami/inventory/pkg/policies"
)
```

**After:**
```go
import (
    "github.com/openchami/inventory-framework/pkg/versioning"
    "github.com/openchami/inventory-framework/pkg/policy"
    "github.com/openchami/inventory-framework/pkg/server"
)
```

**Simplifications possible:**
- Use `server.NewServer()` from framework
- Use `server.AddHealthCheck()` from framework
- Use standard middleware from framework

---

### 2.7 Update Storage Layer

#### Update storage initialization

**Before:**
```go
import "github.com/openchami/inventory/internal/storage"

backend := storage.NewFileBackend(storagePath)
bmcStorage := storage.GetBMCStorage(backend)
```

**After:**
```go
import "github.com/openchami/inventory-framework/pkg/storage"

backend := storage.NewFileBackend(storagePath)
bmcStorage := storage.NewResourceStorage[*bmc.BMC](backend, "BMC")
```

**Changes:**
- Use framework storage interfaces
- Update convenience functions
- Keep file backend from framework

---

### 2.8 Update Policy Implementations

#### Update policy implementations

**Before:**
```go
import "github.com/openchami/inventory/pkg/policies"

type BMCPolicy struct {
    // ...
}

func (p *BMCPolicy) CanList(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.PolicyDecision {
    // ...
}
```

**After:**
```go
import "github.com/openchami/inventory-framework/pkg/policy"

type BMCPolicy struct {
    // ...
}

func (p *BMCPolicy) CanList(ctx context.Context, auth *policy.AuthContext, req *http.Request) policy.Decision {
    // ...
}
```

**Update:**
- [ ] `pkg/resources/bmc/policy.go`
- [ ] `pkg/resources/node/policy.go`
- [ ] Any other custom policies

---

### 2.9 Update Documentation

#### Create Migration Guide (`docs/MIGRATION.md`)
```markdown
# Migration to Framework-Based Architecture

## Overview
OpenCHAMI Inventory now uses the `inventory-framework` for core functionality.

## What Changed
- Core resource model moved to framework
- Versioning system moved to framework
- Storage interfaces moved to framework
- Policy framework moved to framework

## For Users
- No changes to REST API
- No changes to CLI commands
- No changes to data formats

## For Developers
- Update import paths (see table)
- Use framework packages for core functionality
- HPC-specific logic remains in this repository

## Import Path Changes
[Table of old → new imports]
```

#### Update Existing Documentation

**Files to update:**
- `README.md` - Add framework reference
- `docs/developer/DEVELOPMENT.md` - Update architecture section
- `docs/developer/CODE-GENERATION.md` - Reference framework docs
- `docs/user/USER-GUIDE.md` - No changes needed (user-facing unchanged)

**Add new sections:**
- Framework integration guide
- Custom resource development guide
- Contributing to framework vs. inventory

---

### 2.10 Testing & Validation

#### Test Coverage
- [ ] All existing tests pass with framework imports
- [ ] Integration tests validate full system
- [ ] Version negotiation still works
- [ ] Storage operations function correctly
- [ ] Authorization flows unchanged

#### Regression Testing
- [ ] Run full test suite
- [ ] Test version conversion (v1 ↔ v2beta1)
- [ ] Test all CLI commands
- [ ] Test all REST API endpoints
- [ ] Verify data migration (if needed)

#### Performance Testing
- [ ] No significant performance regression
- [ ] Code generation speed maintained
- [ ] Storage operations performance
- [ ] API response times

---

### 2.11 Update Build & Deployment

#### Update Makefile

**Changes:**
- Update code generation to use framework codegen
- Update dependency management
- Add framework update targets

```makefile
# Update code generation
.PHONY: generate
generate:
	go run github.com/openchami/inventory-framework/cmd/inventory-codegen@latest \
		--resources pkg/resources \
		--output cmd/server \
		--package main

# Update framework
.PHONY: update-framework
update-framework:
	go get -u github.com/openchami/inventory-framework@latest
	go mod tidy
```

#### Update CI/CD

**Changes to `.github/workflows/`:**
- Test against multiple framework versions
- Ensure framework compatibility
- Update dependency caching

---

### 2.12 Release Strategy

#### Version Bump Strategy

**Current:** `github.com/openchami/inventory` v2.x
**After refactor:** v3.0.0 (breaking change due to framework extraction)

#### Release Notes Template
```markdown
# v3.0.0 - Framework-Based Architecture

## Major Changes
- Extracted core components to `inventory-framework`
- Cleaner separation of HPC-specific and generic code
- Improved maintainability and extensibility

## Breaking Changes for Developers
- Import paths changed (see migration guide)
- Internal packages reorganized

## No Breaking Changes for Users
- REST API unchanged
- CLI unchanged
- Data formats unchanged
- Configuration unchanged

## Migration Guide
See docs/MIGRATION.md for detailed migration instructions.
```

---

## Implementation Timeline

### Phase 1: Framework Creation (4-6 weeks)

**Week 1-2: Repository Setup & Core Extraction**
- Create repository and structure
- Extract resource model package
- Extract versioning package
- Set up CI/CD

**Week 3-4: Storage, Policy, Codegen**
- Extract storage package
- Extract policy package
- Extract code generation
- Migrate templates

**Week 5-6: Documentation & Examples**
- Write framework documentation
- Create simple example
- Create blog CMS example
- Write API reference

### Phase 2: HPC Inventory Refactoring (2-3 weeks)

**Week 1: Update Dependencies & Imports**
- Update go.mod
- Update all import paths
- Remove duplicated code
- Update code generation

**Week 2: Testing & Validation**
- Fix any breaking issues
- Run full test suite
- Performance testing
- Documentation updates

**Week 3: Release & Migration**
- Create migration guide
- Release framework v0.1.0
- Release inventory v3.0.0
- Support community migration

---

## Success Criteria

### Phase 1 Success Criteria
- ✅ Framework repository created and functional
- ✅ All core packages extracted and documented
- ✅ At least 2 working examples
- ✅ Test coverage >80%
- ✅ CI/CD passing
- ✅ Documentation complete

### Phase 2 Success Criteria
- ✅ All tests passing with framework
- ✅ No performance regression
- ✅ API backward compatible
- ✅ CLI backward compatible
- ✅ Data migration successful (if needed)
- ✅ Documentation updated
- ✅ Release published

---

## Risk Mitigation

### Risk 1: Breaking Changes
**Mitigation:**
- Maintain API compatibility
- Provide migration scripts
- Version bump to v3.0.0
- Comprehensive migration guide

### Risk 2: Performance Regression
**Mitigation:**
- Benchmark before/after
- Profile critical paths
- Optimize hot paths
- Load testing

### Risk 3: Community Confusion
**Mitigation:**
- Clear communication plan
- Detailed migration docs
- Support channels active
- FAQ document

### Risk 4: Dependency Management
**Mitigation:**
- Pin framework version initially
- Semantic versioning
- Compatibility matrix
- Automated dependency updates

---

## Long-Term Vision

### Framework Evolution
- Database backend implementation
- GraphQL API generation option
- gRPC service generation
- Event sourcing support
- Multi-tenancy patterns

### Community Building
- Encourage framework adoption
- Accept community contributions
- Build plugin ecosystem
- Create framework working group

### HPC Inventory Evolution
- Focus on HPC-specific features
- Hardware discovery improvements
- Redfish/IPMI enhancements
- Integration with other OpenCHAMI components

---

## Appendices

### Appendix A: Import Path Reference

| Component | Old Path | New Path |
|-----------|----------|----------|
| Resource Model | `github.com/openchami/inventory/pkg/resources` | `github.com/openchami/inventory-framework/pkg/resource` |
| Versioning | `github.com/openchami/inventory/pkg/versioning` | `github.com/openchami/inventory-framework/pkg/versioning` |
| Storage | `github.com/openchami/inventory/internal/storage` | `github.com/openchami/inventory-framework/pkg/storage` |
| Policy | `github.com/openchami/inventory/pkg/policies` | `github.com/openchami/inventory-framework/pkg/policy` |
| Codegen | `github.com/openchami/inventory/pkg/codegen` | `github.com/openchami/inventory-framework/pkg/codegen` |

### Appendix B: File Migration Checklist

**Files to move to framework:**
- [x] `pkg/resources/resource.go`
- [x] `pkg/resources/metadata.go`
- [x] `pkg/resources/conditions.go`
- [x] `pkg/versioning/*`
- [x] `pkg/codegen/*`
- [x] `internal/storage/interfaces.go`
- [x] `internal/storage/file_backend.go`
- [x] `pkg/policies/registry.go`
- [x] `pkg/policies/permissive.go`
- [x] `pkg/codegen/templates/*`

**Files to remain in HPC inventory:**
- [x] `pkg/resources/bmc/*`
- [x] `pkg/resources/node/*`
- [x] `pkg/resources/fru/*`
- [x] `pkg/resources/boot/*`
- [x] `pkg/crawler/*`
- [x] `cmd/server/*`
- [x] `cmd/inventory-cli/*`
- [x] `cmd/crawler/*`
- [x] `cmd/populate-bmcs/*`

### Appendix C: Communication Plan

**Announcement Channels:**
- GitHub Discussions
- OpenCHAMI mailing list
- Slack/Discord channels
- Blog post on openchami.org

**Key Messages:**
1. Why: Better separation, reusability, maintainability
2. What: Core framework extraction
3. When: Phased approach, stable migration path
4. How: Migration guide, support available

**Support Plan:**
- Migration office hours
- Dedicated support channel
- FAQ document
- Example migration PRs

---

## Conclusion

This refactoring plan provides a comprehensive roadmap for extracting the generic inventory management framework while maintaining the HPC-specific functionality of openchami-inventory. The phased approach minimizes risk and ensures a smooth transition for users and developers.

The resulting architecture will:
- ✅ Enable framework reuse across domains
- ✅ Improve code maintainability
- ✅ Preserve all existing functionality
- ✅ Provide clear upgrade path
- ✅ Foster broader community adoption

**Next Steps:**
1. Review and approve this plan
2. Create framework repository
3. Begin Phase 1 implementation
4. Regular progress updates
5. Community feedback integration
