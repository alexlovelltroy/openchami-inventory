# OpenCHAMI Inventory Documentation

Complete documentation for the OpenCHAMI Inventory system.

## 📖 Documentation Map

### For Users

Start here if you want to **use** the OpenCHAMI Inventory system:

- **[User Guide](user/USER-GUIDE.md)** ⭐ **Start here!**
  - Installation and quick start
  - Server configuration
  - CLI and API usage
  - Common workflows

- **[CLI Reference](user/CLI-REFERENCE.md)**
  - Complete command reference
  - Configuration options
  - Examples and tips

- **[API Reference](user/API-REFERENCE.md)**
  - REST API endpoints
  - Request/response formats
  - Error codes
  - Examples

- **[Version Negotiation](user/VERSION-NEGOTIATION.md)**
  - Multi-version schema support
  - Version conversion
  - Best practices

- **[Authentication](user/AUTHENTICATION.md)**
  - Authentication modes
  - Policy system
  - Testing vs production
  - Custom policies

- **[Troubleshooting](user/TROUBLESHOOTING.md)**
  - Common issues and solutions
  - Error messages
  - Debug mode

### For Developers

Start here if you want to **modify or extend** the OpenCHAMI Inventory system:

- **[Development Guide](developer/DEVELOPMENT.md)** ⭐ **Start here!**
  - Architecture overview
  - Development workflow
  - Adding new resources
  - Contributing guidelines

- **[Code Generation](developer/CODE-GENERATION.md)**
  - Template system
  - Modifying templates
  - Adding features
  - Best practices

- **[Testing Guide](developer/TESTING.md)**
  - Running tests
  - Development mode
  - Integration testing
  - Demo scripts

## 🎯 Quick Links by Task

### I want to...

#### Install and run the system
→ [User Guide - Installation](user/USER-GUIDE.md#installation)
→ [User Guide - Quick Start](user/USER-GUIDE.md#quick-start)

#### Use the CLI tool
→ [User Guide - Using the CLI](user/USER-GUIDE.md#using-the-cli)
→ [CLI Reference](user/CLI-REFERENCE.md)

#### Use the REST API
→ [User Guide - Using the REST API](user/USER-GUIDE.md#using-the-rest-api)
→ [API Reference](user/API-REFERENCE.md)

#### Work with multiple API versions
→ [Version Negotiation](user/VERSION-NEGOTIATION.md)
→ [User Guide - Version Negotiation](user/USER-GUIDE.md#version-negotiation)

#### Configure authentication
→ [Authentication](user/AUTHENTICATION.md)
→ [User Guide - Authentication](user/USER-GUIDE.md#authentication)

#### Fix an error
→ [Troubleshooting](user/TROUBLESHOOTING.md)
→ Search by error message in troubleshooting guide

#### Develop a new feature
→ [Development Guide](developer/DEVELOPMENT.md)
→ [Code Generation](developer/CODE-GENERATION.md)

#### Add a new resource type
→ [Development Guide - How to Add a New Resource](developer/DEVELOPMENT.md#how-to-add-a-new-resource-type)

#### Modify code generation templates
→ [Code Generation - Modifying Templates](developer/CODE-GENERATION.md#modifying-templates)

#### Run tests
→ [Testing Guide](developer/TESTING.md)

## 📚 Component Documentation

Specific documentation for system components:

- **[Storage System](../internal/storage/README.md)** - Storage backends and configuration
- **[BMC v2beta1](../pkg/resources/bmc/v2beta1/README.md)** - Enhanced BMC authentication
- **[Templates](../pkg/codegen/templates/README.md)** - Code generation template reference
- **[Scripts](../scripts/README.md)** - Utility and demo scripts

## 🔍 Documentation by Topic

### Authentication & Security
- [Authentication Guide](user/AUTHENTICATION.md)
- [Testing Mode (--disable-auth)](user/AUTHENTICATION.md#testing-mode)
- [Production Setup](user/AUTHENTICATION.md#production-setup)
- [Custom Policies](user/AUTHENTICATION.md#custom-policies)

### API Versioning
- [Version Negotiation](user/VERSION-NEGOTIATION.md)
- [Version Conversion](user/VERSION-NEGOTIATION.md#version-conversion)
- [BMC v2beta1 Features](../pkg/resources/bmc/v2beta1/README.md)

### Configuration
- [Server Configuration](user/USER-GUIDE.md#server-configuration)
- [CLI Configuration](user/CLI-REFERENCE.md#configuration)
- [Environment Variables](user/USER-GUIDE.md#environment-variables)

### Development
- [Architecture Overview](developer/DEVELOPMENT.md#architecture-overview)
- [Code Generation System](developer/CODE-GENERATION.md)
- [Adding Resources](developer/DEVELOPMENT.md#how-to-add-a-new-resource-type)
- [Testing & Development Mode](developer/TESTING.md)

## 📖 Learning Paths

### Path 1: New User

1. [User Guide - Quick Start](user/USER-GUIDE.md#quick-start) (5 min)
2. [User Guide - Using the CLI](user/USER-GUIDE.md#using-the-cli) (10 min)
3. [Common Workflows](user/USER-GUIDE.md#common-workflows) (15 min)
4. [CLI Reference](user/CLI-REFERENCE.md) (reference)

**Time: ~30 minutes + reference material**

### Path 2: API Consumer

1. [User Guide - Quick Start](user/USER-GUIDE.md#quick-start) (5 min)
2. [API Reference](user/API-REFERENCE.md) (20 min)
3. [Version Negotiation](user/VERSION-NEGOTIATION.md) (15 min)
4. [Authentication](user/AUTHENTICATION.md) (10 min)

**Time: ~50 minutes**

### Path 3: New Developer

1. [Development Guide - Architecture](developer/DEVELOPMENT.md#architecture-overview) (15 min)
2. [Code Generation Overview](developer/CODE-GENERATION.md#overview) (10 min)
3. [Adding a Resource](developer/DEVELOPMENT.md#how-to-add-a-new-resource-type) (30 min)
4. [Testing Guide](developer/TESTING.md) (10 min)

**Time: ~65 minutes**

### Path 4: Template Developer

1. [Code Generation - Architecture](developer/CODE-GENERATION.md#architecture) (10 min)
2. [Templates Overview](developer/CODE-GENERATION.md#templates) (15 min)
3. [Modifying Templates](developer/CODE-GENERATION.md#modifying-templates) (30 min)
4. [Template Reference](../pkg/codegen/templates/README.md) (reference)

**Time: ~55 minutes + reference material**

## 🆘 Getting Help

### Documentation Issues

- **Can't find what you need?** [Open an issue](https://github.com/openchami/inventory/issues/new)
- **Documentation unclear?** [Suggest improvements](https://github.com/openchami/inventory/issues/new)
- **Found an error?** [Submit a PR](https://github.com/openchami/inventory/pulls)

### Technical Support

- **Bug reports**: [GitHub Issues](https://github.com/openchami/inventory/issues)
- **Feature requests**: [GitHub Discussions](https://github.com/openchami/inventory/discussions)
- **Questions**: [GitHub Discussions Q&A](https://github.com/openchami/inventory/discussions/categories/q-a)

## 📝 Contributing to Documentation

We welcome documentation improvements! To contribute:

1. **For typos/small fixes**: Submit a PR directly
2. **For new sections**: Open an issue first to discuss
3. **For reorganization**: Propose changes in an issue

See [CONTRIBUTING.md](../CONTRIBUTING.md) for details.

## 🔖 Version Information

- **Documentation Version**: Corresponds to main branch
- **Last Updated**: October 2025
- **API Version**: inventory/v2
- **Schema Versions**: v1 (stable), v2beta1 (beta)

## 📋 Document Status

| Document | Status | Last Updated |
|----------|--------|--------------|
| User Guide | ✅ Complete | Oct 2025 |
| CLI Reference | ✅ Complete | Oct 2025 |
| API Reference | ✅ Complete | Oct 2025 |
| Version Negotiation | ✅ Complete | Oct 2025 |
| Authentication | ✅ Complete | Oct 2025 |
| Troubleshooting | ✅ Complete | Oct 2025 |
| Development Guide | ✅ Complete | Oct 2025 |
| Code Generation | ✅ Complete | Oct 2025 |
| Testing Guide | ✅ Complete | Oct 2025 |

## 📧 Documentation Feedback

Have feedback on the documentation? We'd love to hear it!

- **What's working well?**
- **What's confusing?**
- **What's missing?**

[Share your feedback](https://github.com/openchami/inventory/discussions/new?category=feedback)
