# Version Demo

This command demonstrates the multi-version support capability implemented in Phase 1 of the Multi-Schema Version Support proposal.

## Purpose

The demo shows how to:
- Register multiple schema versions for the same resource type (Node v1, Node v2beta1, BMC v1)
- Query the version registry for available versions
- Negotiate versions based on client requests
- Handle default version fallback when requested versions are unavailable

## Running the Demo

### Using Make (recommended)
```bash
make demo
```

### Using go run
```bash
go run ./cmd/version-demo/main.go
```

### Build and run binary
```bash
go build -o bin/version-demo ./cmd/version-demo
./bin/version-demo
```

## Output

The demo will display:
1. **Registration**: Shows successful registration of multiple versions
2. **Registry Queries**: Lists all registered resource kinds and their versions
3. **Version Negotiation**: Demonstrates different client request scenarios and how the system responds

## Example Output

```
=== Multi-Version Support Demo ===
✓ Registered Node v1 (stable)
✓ Registered Node v2beta1 (beta)
✓ Registered BMC v1 (stable)

=== Version Registry Queries ===
Registered resource kinds: [BMC Node]
Node versions: [v1 v2beta1] (default: v1)
  v1: stability=stable, deprecated=false, package=github.com/openchami/inventory/pkg/resources/node
  v2beta1: stability=beta, deprecated=false, package=github.com/openchami/inventory/pkg/resources/node/v2beta1

=== Version Negotiation Scenarios ===
✓ Client requests default Node version: requested=, served=v1
✓ Client requests Node v2beta1: requested=v2beta1, served=v2beta1
✓ Client requests non-existent Node v3: requested=v3, served=v1
```

## Related Documentation

- [PROPOSAL-Multi-Schema-Versioning.md](../../PROPOSAL-Multi-Schema-Versioning.md) - Full proposal
- [PHASE1-SUMMARY.md](../../PHASE1-SUMMARY.md) - Phase 1 implementation summary
- [pkg/versioning](../../pkg/versioning/) - Versioning package implementation
