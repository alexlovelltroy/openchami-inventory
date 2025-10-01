# OpenCHAMI Inventory

A modern inventory management system for high-performance computing environments, featuring automatic code generation and modular resource management.

## 🚀 Quick Start

```bash
# See codebase structure and current status
make discover

# Full development build
make dev

# Build specific components
make generate  # Generate all code
make build     # Build binaries
make test      # Run tests
```

## 📁 Project Structure

This project uses **code generation** to maintain consistency across REST APIs, storage operations, and client libraries. The generator reads resource type definitions and creates all the necessary boilerplate code.

```
pkg/resources/     # Resource definitions (EDIT THESE)
├── bmc/           # BMC hardware resources  
├── node/          # Compute node resources
├── fru/           # Field Replaceable Units
└── boot/          # Boot configurations

pkg/codegen/       # Code generation engine
└── templates/     # Template files (EDIT THESE)

cmd/server/        # Generated REST API server
internal/storage/  # Generated storage operations  
pkg/client/        # Generated HTTP client library
```

## 📖 Documentation

- **[DEVELOPMENT.md](docs/DEVELOPMENT.md)** - Comprehensive development guide
- **`make discover`** - Interactive codebase explorer
- **`./scripts/discover.sh`** - Detailed project status and structure
- **[Template Documentation](pkg/codegen/templates/README.md)** - Getting started guide for creating and editing the metaprogramming templates

## 🔧 Development Workflow

1. **Understand the codebase**: `make discover`
2. **Modify resource types**: Edit files in `pkg/resources/*/`
3. **Customize templates**: Edit files in `pkg/codegen/templates/`
4. **Regenerate code**: `make dev`
5. **Test changes**: APIs automatically updated

## ⚡ Key Features

- **Automatic API Generation**: REST endpoints generated from Go structs
- **Type-Safe Client Library**: HTTP client with full type safety
- **Modular Resource System**: Each resource type in separate package
- **Authentication Policies**: Configurable per-resource security
- **File-Based Storage**: Simple persistence with JSON serialization

## 🤝 Contributing

New to the codebase? Start with:
1. `make discover` - See current structure
2. Read `docs/DEVELOPMENT.md` - Understand the architecture  
3. Look at `pkg/resources/node/` - Example resource implementation
4. Check `pkg/codegen/templates/` - See what gets generated
5. Use `make templates` - View template content

## 📝 License

MIT