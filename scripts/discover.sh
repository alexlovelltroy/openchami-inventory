#!/bin/bash

# Developer Discovery Script for OpenCHAMI Inventory
# This script helps developers understand the codebase structure

echo "🔍 OpenCHAMI Inventory Codebase Discovery"
echo "========================================"
echo

echo "📁 DIRECTORY STRUCTURE:"
echo "├── pkg/resources/         # Resource type definitions (EDIT THESE)"
echo "│   ├── bmc/               # BMC hardware resources"
echo "│   ├── node/              # Compute node resources"
echo "│   ├── fru/               # Field Replaceable Units"
echo "│   └── boot/              # Boot configurations"
echo "├── pkg/codegen/           # Code generation engine"
echo "│   └── templates/         # Template files (EDIT THESE)"
echo "├── cmd/server/            # Generated REST API server"
echo "├── cmd/crawler/           # Hardware discovery agent"
echo "├── internal/storage/      # Generated storage operations"
echo "├── pkg/client/            # Generated HTTP client library"
echo "└── pkg/policies/          # Authentication/authorization"
echo

echo "🔧 RESOURCE TYPES:"
for dir in pkg/resources/*/; do
    if [ -d "$dir" ] && [ "$(basename "$dir")" != "resources" ]; then
        resource=$(basename "$dir")
        echo "  $resource:"
        if [ -f "$dir/${resource}.go" ]; then
            types=$(grep "^type.*struct" "$dir/${resource}.go" | sed 's/type //; s/ struct.*//' | tr '\n' ', ')
            echo "    Types: ${types%, }"
        fi
        if [ -f "$dir/policy.go" ]; then
            echo "    Auth: Custom policy defined"
        else
            echo "    Auth: Default policy"
        fi
        if ls cmd/server/*${resource}* >/dev/null 2>&1; then
            echo "    Generated: REST handlers, storage, client"
        fi
        echo
    fi
done

echo "⚡ GENERATED FILES (DO NOT EDIT):"
echo "  cmd/server/*_handlers_generated.go  # REST API endpoints"
echo "  cmd/server/models_generated.go      # Request/response types"
echo "  cmd/server/routes_generated.go      # URL routing"
echo "  internal/storage/storage_generated.go # Data persistence"
echo "  pkg/client/*.go                     # HTTP client library"
echo

echo "✏️  EDITABLE FILES:"
echo "  pkg/resources/*/                    # Add/modify resource types here"
echo "  pkg/codegen/templates/              # Modify code generation templates"
echo "  pkg/policies/                       # Authentication logic"
echo "  cmd/server/main.go                  # Server configuration"
echo "  cmd/crawler/                        # Hardware discovery"
echo

echo "🚀 COMMON WORKFLOWS:"
echo "  # Add new resource type:"
echo "  1. Create pkg/resources/newtype/newtype.go"
echo "  2. Add to cmd/codegen/main.go"
echo "  3. Run 'make dev'"
echo
echo "  # Modify existing resource:"
echo "  1. Edit pkg/resources/typename/typename.go"
echo "  2. Run 'make dev'"
echo
echo "  # Customize API behavior:"
echo "  1. Edit templates in pkg/codegen/templates/"
echo "  2. Run 'make dev'"
echo

echo "📊 CURRENT STATUS:"
echo "  Generated files last updated: $(date)"

if [ -f "internal/storage/storage_generated.go" ]; then
    echo "  ✅ Storage layer: Generated"
else
    echo "  ❌ Storage layer: Missing (run 'make dev')"
fi

if [ -f "cmd/server/routes_generated.go" ]; then
    echo "  ✅ API routes: Generated"
    routes=$(grep -c "fuego\." cmd/server/routes_generated.go 2>/dev/null || echo "0")
    echo "    → $routes endpoints registered"
else
    echo "  ❌ API routes: Missing (run 'make dev')"
fi

if [ -f "pkg/client/client.go" ]; then
    echo "  ✅ Client library: Generated"
else
    echo "  ❌ Client library: Missing (run 'make dev')"
fi

echo
echo "📖 For detailed documentation, see:"
echo "   docs/DEVELOPMENT.md"
echo
echo "❓ Need help? Check the templates in pkg/codegen/templates/"
echo "   Each template is a separate file for easier editing."
echo "   Use 'make templates' to view template content."
echo ""
echo "🔧 Storage System: Interface-based with pluggable backends"
echo "   Default: File-based storage in ./inventory/"
echo "   See: internal/storage/README.md for details"