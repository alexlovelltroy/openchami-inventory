# OpenCHAMI Inventory

> A modern inventory management system for high-performance computing environments with multi-version schema support and automatic code generation.

## 🎯 Overview

OpenCHAMI Inventory provides a flexible, extensible inventory system for HPC hardware management with:

- **Multi-Version Schema Support** - Run multiple API versions simultaneously with automatic conversion
- **Template-Driven Generation** - Maintain consistency across APIs, storage, and clients
- **RESTful API** - Complete CRUD operations with OpenAPI documentation
- **Type-Safe Client** - Go client library with full type safety
- **CLI Tool** - User-friendly command-line interface
- **Pluggable Backends** - Flexible authentication and storage systems

## 🚀 Quick Start

### For Users (Running the System)

```bash
# Start the server (testing mode - no authentication)
./bin/server --port 9999 --disable-auth

# Use the CLI
./bin/inventory-cli --server http://localhost:9999 bmc list

# Populate with sample data
./bin/populate-bmcs
```

**→ See [User Guide](docs/user/USER-GUIDE.md) for complete usage instructions**

### For Developers (Modifying the System)

```bash
# Discover codebase structure
make discover

# Full development build
make dev

# Test changes
make test
```

**→ See [Development Guide](docs/developer/DEVELOPMENT.md) for development workflow**

## 📚 Documentation

### User Documentation

- **[User Guide](docs/user/USER-GUIDE.md)** - Installation, configuration, and usage ⭐
- **[CLI Reference](docs/user/CLI-REFERENCE.md)** - Complete CLI command reference
- **[API Reference](docs/user/API-REFERENCE.md)** - REST API endpoints and examples
- **[Version Negotiation](docs/user/VERSION-NEGOTIATION.md)** - Multi-version schema guide
- **[Authentication](docs/user/AUTHENTICATION.md)** - Security and policy management
- **[Troubleshooting](docs/user/TROUBLESHOOTING.md)** - Common issues and solutions

### Developer Documentation

- **[Development Guide](docs/developer/DEVELOPMENT.md)** - Architecture and development ⭐
- **[Code Generation](docs/developer/CODE-GENERATION.md)** - Template system guide
- **[Testing Guide](docs/developer/TESTING.md)** - Testing and development mode

### Quick Access

- **[Documentation Index](docs/README.md)** - Complete documentation map
- **`make discover`** - Interactive codebase explorer

## 🏗️ Architecture

OpenCHAMI Inventory uses template-based code generation to maintain consistency:

```
Resource Definition (Go struct)
    ↓
Code Generator
    ↓
Generated Code
    ├─ REST API Handlers
    ├─ Storage Operations
    ├─ HTTP Client Library
    └─ CLI Commands
```

**Key Components:**
- `pkg/resources/` - Resource type definitions (manual)
- `pkg/codegen/templates/` - Code generation templates (manual)
- `cmd/server/` - REST API server (generated)
- `internal/storage/` - Storage operations (generated)
- `pkg/client/` - HTTP client library (generated)
- `cmd/inventory-cli/` - CLI application (generated)

## ⚡ Key Features

### Multi-Version Schema Support

Support multiple API versions simultaneously with automatic conversion:

```bash
# Get resource as v1
./bin/inventory-cli --version v1 bmc get <uid>

# Get same resource as v2beta1
./bin/inventory-cli --version v2beta1 bmc get <uid>
```

The server automatically converts between versions, enabling:
- Backward compatibility for old clients
- Forward compatibility for new features
- Gradual migration at your own pace

**→ See [Version Negotiation Guide](docs/user/VERSION-NEGOTIATION.md)**

### Automatic Code Generation

Modify resource types and templates, regenerate everything:

```bash
# Edit resource definition
vim pkg/resources/bmc/bmc.go

# Or edit template
vim pkg/codegen/templates/handlers.go.tmpl

# Regenerate all code
make dev
```

This generates:
- REST API handlers with CRUD operations
- Type-safe storage layer
- HTTP client library methods
- CLI commands

**→ See [Code Generation Guide](docs/developer/CODE-GENERATION.md)**

### Flexible Authentication

Testing mode for development, policy-based auth for production:

```bash
# Testing mode: allows all operations
./bin/server --port 9999 --disable-auth

# Production mode: requires JWT authentication
./bin/server --port 9999
```

Create custom policies for fine-grained authorization control.

**→ See [Authentication Guide](docs/user/AUTHENTICATION.md)**

## 📦 Installation

### From Source

```bash
# Clone repository
git clone https://github.com/openchami/inventory.git
cd inventory

# Build everything
make build

# Binaries available in ./bin/
ls -la bin/
# bin/server
# bin/inventory-cli
# bin/populate-bmcs
```

**Requirements:**
- Go 1.24 or later
- Make

## 🔧 Configuration

### Server Configuration

```bash
./bin/server \
  --host 0.0.0.0 \
  --port 9999 \
  --disable-auth \
  --storage-path ./inventory \
  --log-level info
```

Or use environment variables:

```bash
export INVENTORY_HOST=0.0.0.0
export INVENTORY_PORT=9999
export INVENTORY_DISABLE_AUTH=true

./bin/server
```

**→ See [User Guide - Configuration](docs/user/USER-GUIDE.md#server-configuration)**

## 🧪 Testing

```bash
# Run tests
make test

# Start server in testing mode
./bin/server --disable-auth

# Run version negotiation demo
./scripts/demo-version-negotiation.sh
```

**→ See [Testing Guide](docs/developer/TESTING.md)**

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### New to the Project?

1. **Explore the codebase**: `make discover`
2. **Read the architecture**: [Development Guide](docs/developer/DEVELOPMENT.md)
3. **Try adding a resource**: Follow the [guide](docs/developer/DEVELOPMENT.md#how-to-add-a-new-resource-type)
4. **Understand templates**: [Code Generation Guide](docs/developer/CODE-GENERATION.md)

### Contributing Guidelines

- **Code of Conduct**: Be respectful and collaborative
- **Issues**: Check existing issues before creating new ones
- **Pull Requests**: Include tests and documentation
- **Commit Messages**: Use clear, descriptive messages

## 📝 License

MIT License - See [LICENSE](LICENSE)

## 🔗 Links

- **[GitHub Repository](https://github.com/openchami/inventory)**
- **[OpenCHAMI Project](https://openchami.org)**
- **[Issue Tracker](https://github.com/openchami/inventory/issues)**
- **[Discussions](https://github.com/openchami/inventory/discussions)**

## � Project Status

- **API Version**: inventory/v2
- **Schema Versions**: v1 (stable), v2beta1 (beta)
- **Go Version**: 1.24+
- **Status**: Active Development