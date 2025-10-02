# API Reference

REST API documentation for OpenCHAMI Inventory.

## Base URL

```
http://localhost:8080
```

Configure with `--host` and `--port` flags when starting the server.

## Table of Contents

- [Common Headers](#common-headers)
- [Response Format](#response-format)
- [Error Responses](#error-responses)
- [Version Negotiation](#version-negotiation)
- [BMC Endpoints](#bmc-endpoints)
- [Node Endpoints](#node-endpoints)
- [FRU Endpoints](#fru-endpoints)
- [Boot Configuration Endpoints](#boot-configuration-endpoints)
- [System Endpoints](#system-endpoints)

## Common Headers

### Request Headers

| Header | Description | Example |
|--------|-------------|---------|
| `Content-Type` | Request body format | `application/json` or `application/json;version=v2beta1` |
| `Accept` | Desired response format | `application/json` or `application/json;version=v1` |
| `Authorization` | Authentication token (when auth enabled) | `Bearer <jwt-token>` |

### Response Headers

| Header | Description |
|--------|-------------|
| `Content-Type` | Response body format |
| `X-Schema-Version` | Actual schema version returned |

## Response Format

### Success Response

All successful responses return the resource(s) directly:

**Single resource (GET, POST, PUT):**
```json
{
  "apiVersion": "inventory/v2",
  "kind": "BMC",
  "schemaVersion": "v1",
  "metadata": {
    "name": "node001-bmc",
    "uid": "550e8400-e29b-41d4-a716-446655440000",
    "creationTimestamp": "2025-10-02T10:00:00Z",
    "labels": {},
    "annotations": {}
  },
  "spec": {
    "address": "https://10.1.1.100",
    "username": "admin",
    "type": "Redfish"
  },
  "status": {
    "conditions": []
  }
}
```

**Multiple resources (LIST):**
```json
[
  {
    "apiVersion": "inventory/v2",
    "kind": "BMC",
    "metadata": {...},
    "spec": {...}
  },
  {
    "apiVersion": "inventory/v2",
    "kind": "BMC",
    "metadata": {...},
    "spec": {...}
  }
]
```

## Error Responses

### Error Format

```json
{
  "error": "error message",
  "details": "additional context (optional)"
}
```

### Status Codes

| Code | Meaning | Common Causes |
|------|---------|---------------|
| 200 | OK | Successful GET, PUT, DELETE |
| 201 | Created | Successful POST |
| 400 | Bad Request | Invalid JSON, missing required fields |
| 401 | Unauthorized | Authentication required but not provided |
| 403 | Forbidden | Authentication provided but insufficient permissions |
| 404 | Not Found | Resource UID doesn't exist |
| 409 | Conflict | Resource already exists (duplicate UID/name) |
| 500 | Internal Server Error | Server-side error |

## Version Negotiation

### Requesting Specific Versions

Use `Accept` and `Content-Type` headers with version parameter:

```bash
# Request v1
curl -H "Accept: application/json;version=v1" \
  http://localhost:9999/bmcs/<uid>

# Request v2beta1
curl -H "Accept: application/json;version=v2beta1" \
  http://localhost:9999/bmcs/<uid>

# Create with v2beta1
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json;version=v2beta1" \
  -H "Accept: application/json;version=v2beta1" \
  -d '{...}'
```

### Version Discovery

**Endpoint:** `GET /version-info`

**Response:**
```json
{
  "apiVersion": "inventory/v2",
  "defaultSchemaVersion": "v1",
  "supportedResourceVersions": {
    "BMC": ["v1", "v2beta1"],
    "Node": ["v1"],
    "FRU": ["v1"],
    "BootConfiguration": ["v1"]
  }
}
```

## BMC Endpoints

Baseboard Management Controller resources.

### List BMCs

```
GET /bmcs
```

**Example:**
```bash
curl http://localhost:9999/bmcs | jq
```

**Response:** Array of BMC resources

### Get BMC

```
GET /bmcs/{uid}
```

**Parameters:**
- `uid` - Resource UID

**Example:**
```bash
curl http://localhost:9999/bmcs/550e8400-e29b-41d4-a716-446655440000 | jq
```

**Response:** Single BMC resource

### Create BMC

```
POST /bmcs
```

**Request Body (v1):**
```json
{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "username": "admin",
  "password": "changeme",
  "type": "Redfish"
}
```

**Request Body (v2beta1):**
```json
{
  "name": "node001-bmc",
  "address": "https://10.1.1.100",
  "type": "Redfish",
  "authentication": {
    "method": "basic",
    "basic": {
      "username": "admin",
      "password": "changeme"
    }
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:9999/bmcs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "node001-bmc",
    "address": "https://10.1.1.100",
    "username": "admin",
    "password": "changeme",
    "type": "Redfish"
  }' | jq
```

**Response:** Created BMC resource with generated UID

### Update BMC

```
PUT /bmcs/{uid}
```

**Parameters:**
- `uid` - Resource UID

**Request Body:** Partial resource (only fields to update)

**Example:**
```bash
curl -X PUT http://localhost:9999/bmcs/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "password": "new-secure-password"
  }' | jq
```

**Response:** Updated BMC resource

### Delete BMC

```
DELETE /bmcs/{uid}
```

**Parameters:**
- `uid` - Resource UID

**Example:**
```bash
curl -X DELETE http://localhost:9999/bmcs/550e8400-e29b-41d4-a716-446655440000
```

**Response:** 200 OK (empty body)

## Node Endpoints

Compute node resources.

### List Nodes

```
GET /nodes
```

**Example:**
```bash
curl http://localhost:9999/nodes | jq
```

### Get Node

```
GET /nodes/{uid}
```

**Example:**
```bash
curl http://localhost:9999/nodes/660e8400-e29b-41d4-a716-446655440001 | jq
```

### Create Node

```
POST /nodes
```

**Request Body:**
```json
{
  "name": "node001",
  "xname": "x1000c0s0b0n0",
  "nid": 1,
  "role": "compute"
}
```

**Example:**
```bash
curl -X POST http://localhost:9999/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "node001",
    "xname": "x1000c0s0b0n0",
    "nid": 1,
    "role": "compute"
  }' | jq
```

### Update Node

```
PUT /nodes/{uid}
```

**Example:**
```bash
curl -X PUT http://localhost:9999/nodes/660e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -d '{
    "role": "login",
    "status": {
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "reason": "NodeReady",
          "message": "Node is ready"
        }
      ]
    }
  }' | jq
```

### Delete Node

```
DELETE /nodes/{uid}
```

**Example:**
```bash
curl -X DELETE http://localhost:9999/nodes/660e8400-e29b-41d4-a716-446655440001
```

## FRU Endpoints

Field Replaceable Unit resources.

### List FRUs

```
GET /frus
```

### Get FRU

```
GET /frus/{uid}
```

### Create FRU

```
POST /frus
```

**Request Body:**
```json
{
  "name": "chassis-001",
  "type": "chassis",
  "manufacturer": "HPE",
  "model": "Apollo 6500",
  "serialNumber": "SN123456",
  "partNumber": "PN789012"
}
```

### Update FRU

```
PUT /frus/{uid}
```

### Delete FRU

```
DELETE /frus/{uid}
```

## Boot Configuration Endpoints

Boot configuration resources.

### List Boot Configurations

```
GET /bootconfigurations
```

### Get Boot Configuration

```
GET /bootconfigurations/{uid}
```

### Create Boot Configuration

```
POST /bootconfigurations
```

**Request Body:**
```json
{
  "name": "compute-boot",
  "kernel": "vmlinuz-5.15.0",
  "initrd": "initrd-5.15.0.img",
  "cmdline": "console=ttyS0 root=/dev/sda1"
}
```

### Update Boot Configuration

```
PUT /bootconfigurations/{uid}
```

### Delete Boot Configuration

```
DELETE /bootconfigurations/{uid}
```

## System Endpoints

System information and health check endpoints.

### Health Check

```
GET /health
```

**Response:**
```json
{
  "status": "ok"
}
```

**Example:**
```bash
curl http://localhost:9999/health
```

### Version Information

```
GET /version-info
```

**Response:**
```json
{
  "apiVersion": "inventory/v2",
  "defaultSchemaVersion": "v1",
  "supportedResourceVersions": {
    "BMC": ["v1", "v2beta1"],
    "Node": ["v1"]
  }
}
```

**Example:**
```bash
curl http://localhost:9999/version-info | jq
```

### OpenAPI Specification

```
GET /openapi.json
```

**Response:** OpenAPI 3.0 specification document

**Example:**
```bash
curl http://localhost:9999/openapi.json | jq
```

## Advanced Examples

### Bulk Create

```bash
#!/bin/bash
# Create multiple BMCs

for i in {1..10}; do
  curl -X POST http://localhost:9999/bmcs \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"node$(printf %03d $i)-bmc\",
      \"address\": \"https://10.1.1.$((100+i))\",
      \"username\": \"admin\",
      \"password\": \"changeme\",
      \"type\": \"Redfish\"
    }"
  echo
done
```

### Query with jq

```bash
# Get all BMCs in datacenter dc1
curl http://localhost:9999/bmcs | \
  jq '.[] | select(.metadata.labels.datacenter == "dc1")'

# Count BMCs by type
curl http://localhost:9999/bmcs | \
  jq 'group_by(.spec.type) | map({type: .[0].spec.type, count: length})'

# Extract just names and addresses
curl http://localhost:9999/bmcs | \
  jq '.[] | {name: .metadata.name, address: .spec.address}'
```

### Version Comparison

```bash
# Get resource in both versions
curl -H "Accept: application/json;version=v1" \
  http://localhost:9999/bmcs/<uid> > v1.json

curl -H "Accept: application/json;version=v2beta1" \
  http://localhost:9999/bmcs/<uid> > v2beta1.json

# Compare
diff -u v1.json v2beta1.json
```

### Conditional Updates

```bash
# Update only if condition exists
BMC_JSON=$(curl -s http://localhost:9999/bmcs/<uid>)

if echo "$BMC_JSON" | jq -e '.status.conditions[]? | select(.type == "Ready")' > /dev/null; then
  echo "BMC is ready, updating..."
  curl -X PUT http://localhost:9999/bmcs/<uid> \
    -H "Content-Type: application/json" \
    -d '{"password": "new-password"}'
fi
```

## Rate Limiting

Currently no rate limiting is implemented. In production, consider:
- Implementing rate limiting middleware
- Using API gateway with rate limits
- Monitoring API usage

## Authentication

### Testing Mode (No Auth)

Start server with `--disable-auth`:
```bash
./bin/server --disable-auth
```

No authentication required for requests.

### Production Mode (Auth Required)

Without `--disable-auth`, all requests require JWT authentication:

```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:9999/bmcs
```

See [AUTHENTICATION.md](./AUTHENTICATION.md) for details.

## Best Practices

1. **Always specify Content-Type** when sending JSON
2. **Use version headers** for production code to ensure compatibility
3. **Check status codes** before parsing response
4. **Handle errors gracefully** with proper error messages
5. **Use HTTPS** in production (terminate TLS at load balancer)
6. **Validate JSON** before sending to avoid 400 errors
7. **Use resource UIDs** not names for gets/updates/deletes
8. **Store UIDs** after creating resources

## See Also

- [User Guide](./USER-GUIDE.md) - Complete usage guide
- [CLI Reference](./CLI-REFERENCE.md) - CLI tool documentation
- [Version Negotiation](./VERSION-NEGOTIATION.md) - Multi-version guide
- [Authentication](./AUTHENTICATION.md) - Security and policies
