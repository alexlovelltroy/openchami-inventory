# Reconciliation & Events System Proposal

## Executive Summary

This proposal outlines the design and implementation of a **lightweight reconciliation and events system** for the inventory framework. The system will enable resources to automatically reconcile their desired state (Spec) with observed state (Status) using a pluggable workflow engine (supporting both embedded go-workflows and Temporal).

**Key Benefits:**
- **Declarative Infrastructure**: Spec → Status reconciliation loop
- **Event-Driven Architecture**: React to resource changes automatically
- **Pluggable Workflow Engine**: Choose between embedded or distributed execution
- **Code Generation**: Minimal boilerplate for resource reconcilers
- **HPC Use Cases**: Automatic BMC discovery, FRU inventory sync, health monitoring

---

## Table of Contents

1. [Motivation](#motivation)
2. [Architecture Overview](#architecture-overview)
3. [Core Components](#core-components)
4. [Code Generation](#code-generation)
5. [Workflow Engine Abstraction](#workflow-engine-abstraction)
6. [Implementation Plan](#implementation-plan)
7. [HPC Use Cases](#hpc-use-cases)
8. [API Reference](#api-reference)
9. [Examples](#examples)

---

## Motivation

### Current State

The inventory system currently supports:
- ✅ Resource CRUD operations via REST API
- ✅ Spec/Status pattern (Kubernetes-style)
- ✅ Conditions for status tracking
- ❌ **No automatic reconciliation** - Status must be manually updated
- ❌ **No event system** - Cannot react to resource changes
- ❌ **No workflow orchestration** - Complex operations require custom code

### Problems This Solves

**Problem 1: Manual Status Updates**
```go
// Current: User must manually crawl BMC and update status
bmc := getBMC("bmc-123")
crawlResult := crawlBMC(bmc.Spec.Address)
bmc.Status.Connected = crawlResult.Connected
bmc.Status.Version = crawlResult.Version
updateBMC(bmc)
```

**Problem 2: No Reactive Behavior**
- BMC address changed → Must manually re-discover FRUs
- Node marked for maintenance → Must manually update boot config
- FRU fails → Must manually update alerts

**Problem 3: Complex Workflows**
- Multi-step hardware provisioning
- Coordinated updates across resources
- Retry logic and error handling
- Long-running operations

### Desired State

**Declarative Reconciliation:**
```go
// Desired: Define reconciler, system handles the rest
type BMCReconciler struct {
    reconcile.BaseReconciler
}

func (r *BMCReconciler) Reconcile(ctx context.Context, bmc *bmc.BMC) error {
    // Connect to BMC
    client := connectToBMC(bmc.Spec.Address, bmc.Spec.Username, bmc.Spec.Password)

    // Update status
    bmc.Status.Connected = client.IsConnected()
    bmc.Status.Version = client.GetVersion()

    // Trigger FRU discovery if BMC newly connected
    if bmc.Status.Connected && !wasConnected {
        r.TriggerEvent("bmc.connected", bmc)
    }

    return nil
}
```

**Event-Driven Workflows:**
```go
// React to BMC connection event
func (r *FRUReconciler) OnBMCConnected(ctx context.Context, event Event) error {
    bmc := event.Resource.(*bmc.BMC)
    // Automatically discover FRUs for newly connected BMC
    return r.DiscoverFRUs(ctx, bmc)
}
```

---

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      REST API Layer                          │
│  (Resource Create/Update/Delete triggers reconciliation)    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Event Bus (In-Memory)                     │
│  • Resource change events                                    │
│  • Custom domain events                                      │
│  • Event routing to subscribers                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Reconciliation Controller                    │
│  • Watches resource changes                                  │
│  • Queues reconciliation requests                            │
│  • Manages reconciler lifecycle                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Workflow Engine Abstraction                     │
│  ┌──────────────────┐        ┌──────────────────┐          │
│  │  go-workflows    │   OR   │     Temporal      │          │
│  │  (embedded)      │        │   (distributed)   │          │
│  └──────────────────┘        └──────────────────┘          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Resource Reconcilers                       │
│  • BMCReconciler: Connect, check health, update status       │
│  • FRUReconciler: Discover FRUs, track inventory             │
│  • NodeReconciler: Configure node, update boot config        │
└─────────────────────────────────────────────────────────────┘
```

### Reconciliation Loop

```
   Desired State (Spec)
          │
          ▼
   ┌──────────────┐
   │  Reconciler  │ ← Periodic sync (configurable)
   └──────────────┘
          │
          ▼
   Observed State (Status)
          │
          ▼
   ┌──────────────┐
   │ Update Status│
   └──────────────┘
          │
          ▼
    Event Emission
          │
          ▼
   Other Reconcilers (React to events)
```

---

## Core Components

### 1. Event System (`pkg/events/`)

#### CloudEvents Standard Compliance

The event system adopts the [CloudEvents](https://cloudevents.io/) specification for maximum interoperability. CloudEvents is a CNCF standard for describing event data in a common way, enabling:
- **Interoperability** across different event systems
- **Protocol-agnostic** event transport
- **Integration** with existing cloud-native tooling
- **Standard tooling** support (tracing, routing, transformation)

#### Event Structure

```go
package events

import (
    "context"
    "time"

    cloudevents "github.com/cloudevents/sdk-go/v2"
)

// Event wraps CloudEvents specification
type Event struct {
    cloudevents.Event
}

// NewEvent creates a CloudEvents-compliant event
func NewEvent(eventType, source string, data interface{}) Event {
    event := cloudevents.NewEvent()
    event.SetID(generateEventID())
    event.SetType(eventType)
    event.SetSource(source)
    event.SetTime(time.Now())
    event.SetDataContentType("application/json")
    event.SetData(cloudevents.ApplicationJSON, data)

    return Event{Event: event}
}

// ResourceEvent extends CloudEvents with inventory-specific fields
type ResourceEvent struct {
    cloudevents.Event

    // Inventory-specific extensions (using CloudEvents extension attributes)
    ResourceKind string      `json:"resourcekind"` // Extension: inventory.resourcekind
    ResourceUID  string      `json:"resourceuid"`  // Extension: inventory.resourceuid
    Action       string      `json:"action"`       // Extension: inventory.action
}

// EventHandler processes CloudEvents
type EventHandler func(ctx context.Context, event cloudevents.Event) error

// EventBus manages event publishing and subscription
type EventBus interface {
    // Publish a CloudEvent
    Publish(ctx context.Context, event cloudevents.Event) error

    // Subscribe to events by type pattern (supports wildcards)
    Subscribe(eventType string, handler EventHandler) (SubscriptionID, error)

    // Unsubscribe from events
    Unsubscribe(id SubscriptionID) error

    // Close the event bus
    Close() error
}

// InMemoryEventBus implements EventBus with in-memory channels
type InMemoryEventBus struct {
    subscribers map[string][]subscription
    eventQueue  chan cloudevents.Event
    // ...
}

// NATSEventBus implements EventBus with NATS JetStream
type NATSEventBus struct {
    conn *nats.Conn
    js   nats.JetStreamContext
    // ...
}

// KafkaEventBus implements EventBus with Apache Kafka
type KafkaEventBus struct {
    producer *kafka.Producer
    consumer *kafka.Consumer
    // ...
}

// RedisEventBus implements EventBus with Redis Streams (optional)
type RedisEventBus struct {
    client *redis.Client
    // ...
}
```

#### CloudEvents Extension Attributes

The system uses CloudEvents extension attributes for inventory-specific metadata:

```json
{
  "specversion": "1.0",
  "type": "io.openchami.inventory.bmc.connected",
  "source": "/inventory/bmcs/bmc-abc123",
  "id": "evt-xyz789",
  "time": "2025-10-02T15:30:00Z",
  "datacontenttype": "application/json",
  "data": {
    "uid": "bmc-abc123",
    "address": "10.0.0.1",
    "version": "2.1.0"
  },
  "inventoryresourcekind": "BMC",
  "inventoryresourceuid": "bmc-abc123",
  "inventoryaction": "connected"
}
```

#### Event Type Naming Convention

Following CloudEvents best practices with reverse-DNS naming:

**System Events:**
- `io.openchami.inventory.<resource>.created` - Resource created
- `io.openchami.inventory.<resource>.updated` - Resource updated
- `io.openchami.inventory.<resource>.deleted` - Resource deleted
- `io.openchami.inventory.<resource>.status.changed` - Status fields changed

**Domain Events:**
- `io.openchami.inventory.bmc.connected` - BMC successfully connected
- `io.openchami.inventory.bmc.disconnected` - BMC connection lost
- `io.openchami.inventory.fru.discovered` - New FRU discovered
- `io.openchami.inventory.node.ready` - Node is ready for use
- `io.openchami.inventory.node.maintenance` - Node entering maintenance mode

**Wildcard Subscriptions:**
- `io.openchami.inventory.bmc.*` - All BMC events
- `io.openchami.inventory.*.created` - All creation events
- `io.openchami.inventory.*` - All inventory events

### 2. Reconciliation Framework (`pkg/reconcile/`)

#### Reconciler Interface

```go
package reconcile

import (
    "context"
    "time"
)

// Reconciler handles resource reconciliation
type Reconciler interface {
    // Reconcile brings resource to desired state
    Reconcile(ctx context.Context, resource interface{}) (Result, error)

    // GetResourceKind returns the resource kind this reconciler handles
    GetResourceKind() string
}

// Result indicates reconciliation outcome
type Result struct {
    Requeue      bool          // Should this be requeued?
    RequeueAfter time.Duration // Requeue after this duration
}

// BaseReconciler provides common reconciler functionality
type BaseReconciler struct {
    Client      ClientInterface
    EventBus    events.EventBus
    Logger      Logger
    WorkflowMgr WorkflowManager
}

// Helper methods
func (r *BaseReconciler) UpdateStatus(ctx context.Context, resource interface{}) error
func (r *BaseReconciler) EmitEvent(ctx context.Context, eventType string, resource interface{}) error
func (r *BaseReconciler) SetCondition(resource interface{}, condType, status, reason, message string) error
```

#### Controller

```go
// Controller manages reconciler lifecycle
type Controller struct {
    reconcilers map[string]Reconciler
    queue       WorkQueue
    eventBus    events.EventBus
    storage     storage.Backend
}

func (c *Controller) Start(ctx context.Context) error
func (c *Controller) Stop() error
func (c *Controller) RegisterReconciler(r Reconciler) error
func (c *Controller) Enqueue(request ReconcileRequest) error
```

#### Work Queue

```go
// WorkQueue manages reconciliation requests
type WorkQueue interface {
    Add(item interface{})
    Get() (item interface{}, shutdown bool)
    Done(item interface{})
    ShutDown()
}

// ReconcileRequest represents a reconciliation request
type ReconcileRequest struct {
    ResourceKind string
    ResourceUID  string
    Reason       string // Why reconciliation was triggered
}
```

### 3. Workflow Engine Abstraction (`pkg/workflows/`)

#### Workflow Interface

```go
package workflows

import "context"

// WorkflowManager abstracts workflow execution
type WorkflowManager interface {
    // Execute a workflow
    ExecuteWorkflow(ctx context.Context, workflow Workflow, input interface{}) (WorkflowExecution, error)

    // Get workflow execution
    GetExecution(ctx context.Context, id string) (WorkflowExecution, error)

    // Cancel workflow execution
    CancelExecution(ctx context.Context, id string) error

    // Close workflow manager
    Close() error
}

// Workflow represents a workflow definition
type Workflow interface {
    Name() string
    Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// WorkflowExecution represents a running workflow
type WorkflowExecution interface {
    ID() string
    Status() ExecutionStatus
    Result() (interface{}, error)
    Cancel() error
}

// ExecutionStatus represents workflow execution status
type ExecutionStatus string

const (
    StatusRunning   ExecutionStatus = "Running"
    StatusCompleted ExecutionStatus = "Completed"
    StatusFailed    ExecutionStatus = "Failed"
    StatusCancelled ExecutionStatus = "Cancelled"
)
```

#### go-workflows Implementation

```go
package workflows

import (
    "github.com/cschleiden/go-workflows/workflow"
    "github.com/cschleiden/go-workflows/backend/sqlite"
)

// GoWorkflowsManager implements WorkflowManager using go-workflows
type GoWorkflowsManager struct {
    backend workflow.Backend
    client  workflow.Client
}

func NewGoWorkflowsManager(dbPath string) (*GoWorkflowsManager, error) {
    backend := sqlite.NewSqliteBackend(dbPath)
    client := workflow.NewClient(backend)

    return &GoWorkflowsManager{
        backend: backend,
        client:  client,
    }, nil
}

func (m *GoWorkflowsManager) ExecuteWorkflow(ctx context.Context, wf Workflow, input interface{}) (WorkflowExecution, error) {
    // Wrap workflow for go-workflows
    instance, err := m.client.CreateWorkflowInstance(ctx, workflow.WorkflowDefinition{
        Name: wf.Name(),
        Func: wf.Execute,
    }, input)

    return &goWorkflowExecution{instance: instance}, err
}
```

#### Temporal Implementation

```go
package workflows

import (
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

// TemporalManager implements WorkflowManager using Temporal
type TemporalManager struct {
    client client.Client
    worker worker.Worker
}

func NewTemporalManager(hostPort string) (*TemporalManager, error) {
    c, err := client.Dial(client.Options{
        HostPort: hostPort,
    })
    if err != nil {
        return nil, err
    }

    w := worker.New(c, "inventory-task-queue", worker.Options{})

    return &TemporalManager{
        client: c,
        worker: w,
    }, nil
}

func (m *TemporalManager) ExecuteWorkflow(ctx context.Context, wf Workflow, input interface{}) (WorkflowExecution, error) {
    options := client.StartWorkflowOptions{
        TaskQueue: "inventory-task-queue",
    }

    we, err := m.client.ExecuteWorkflow(ctx, options, wf.Name(), input)
    return &temporalExecution{execution: we}, err
}
```

---

## Code Generation

### Template-Based Reconciler Generation

#### Template: `templates/reconciler.go.tmpl`

```go
// Code generated by inventory-codegen. DO NOT EDIT.
package reconcilers

import (
    "context"
    "time"

    "github.com/openchami/inventory-framework/pkg/reconcile"
    "github.com/openchami/inventory-framework/pkg/events"
    "{{ .Package }}"
)

// {{ .Name }}Reconciler reconciles {{ .Name }} resources
type {{ .Name }}Reconciler struct {
    reconcile.BaseReconciler
}

// NewDefaultReconciler creates a default {{ .Name }} reconciler
func NewDefault{{ .Name }}Reconciler(client reconcile.ClientInterface, eventBus events.EventBus) *{{ .Name }}Reconciler {
    return &{{ .Name }}Reconciler{
        BaseReconciler: reconcile.BaseReconciler{
            Client:   client,
            EventBus: eventBus,
        },
    }
}

// GetResourceKind returns the resource kind
func (r *{{ .Name }}Reconciler) GetResourceKind() string {
    return "{{ .Name }}"
}

// Reconcile brings {{ .Name }} to desired state
func (r *{{ .Name }}Reconciler) Reconcile(ctx context.Context, resource interface{}) (reconcile.Result, error) {
    res := resource.(*{{ .PackageAlias }}.{{ .Name }})

    // Call custom reconciliation logic
    if err := r.reconcile{{ .Name }}(ctx, res); err != nil {
        r.SetCondition(res, "Ready", "False", "ReconcileError", err.Error())
        return reconcile.Result{Requeue: true, RequeueAfter: 30 * time.Second}, err
    }

    r.SetCondition(res, "Ready", "True", "ReconcileSuccess", "Reconciliation successful")

    // Update status in storage
    if err := r.UpdateStatus(ctx, res); err != nil {
        return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, err
    }

    // Emit reconciled event
    r.EmitEvent(ctx, "{{ .PluralName }}.reconciled", res)

    return reconcile.Result{RequeueAfter: 5 * time.Minute}, nil
}

// reconcile{{ .Name }} contains custom reconciliation logic
// This is where developers implement domain-specific logic
func (r *{{ .Name }}Reconciler) reconcile{{ .Name }}(ctx context.Context, res *{{ .PackageAlias }}.{{ .Name }}) error {
    // TODO: Implement {{ .Name }}-specific reconciliation logic
    // This function is generated once and NOT overwritten
    return nil
}
```

#### Template: `templates/reconciler-registration.go.tmpl`

```go
// Code generated by inventory-codegen. DO NOT EDIT.
package main

import (
    "github.com/openchami/inventory-framework/pkg/reconcile"
    "github.com/openchami/inventory-framework/pkg/events"

    "{{ .ModulePath }}/pkg/reconcilers"
)

// RegisterReconcilers registers all reconcilers with the controller
func RegisterReconcilers(controller *reconcile.Controller, client reconcile.ClientInterface, eventBus events.EventBus) error {
    {{- range .Resources }}
    // Register {{ .Name }} reconciler
    {{ .PluralName }}Reconciler := reconcilers.NewDefault{{ .Name }}Reconciler(client, eventBus)
    if err := controller.RegisterReconciler({{ .PluralName }}Reconciler); err != nil {
        return err
    }
    {{- end }}

    return nil
}
```

#### Template: `templates/event-handlers.go.tmpl`

```go
// Code generated by inventory-codegen. DO NOT EDIT.
package reconcilers

import (
    "context"

    "github.com/openchami/inventory-framework/pkg/events"
)

// RegisterEventHandlers registers all event handlers
func RegisterEventHandlers(eventBus events.EventBus) error {
    {{- range .Resources }}
    {{- range .Events }}
    // Handle {{ .Type }} events
    if _, err := eventBus.Subscribe("{{ .Type }}", handle{{ .HandlerName }}); err != nil {
        return err
    }
    {{- end }}
    {{- end }}

    return nil
}

{{- range .Resources }}
{{- range .Events }}

// handle{{ .HandlerName }} handles {{ .Type }} events
func handle{{ .HandlerName }}(ctx context.Context, event events.Event) error {
    // TODO: Implement {{ .HandlerName }} logic
    return nil
}
{{- end }}
{{- end }}
```

### Generated Code Example

**Input Resource Definition:**
```go
// pkg/resources/bmc/bmc.go
package bmc

type BMC struct {
    resources.Resource
    Spec   BMCSpec
    Status BMCStatus
}

// Reconciler configuration (via struct tags or separate file)
// +reconcile:enabled=true
// +reconcile:interval=5m
// +reconcile:events=bmc.connected,bmc.disconnected
```

**Generated Output:**
```go
// pkg/reconcilers/bmc_reconciler.go (generated once)
func (r *BMCReconciler) reconcileBMC(ctx context.Context, res *bmc.BMC) error {
    // Connect to BMC
    client, err := r.connectToBMC(res.Spec.Address, res.Spec.Username, res.Spec.Password)
    if err != nil {
        res.Status.Connected = false
        res.Status.Reachable = false
        return err
    }

    // Update status
    res.Status.Connected = true
    res.Status.Reachable = true
    res.Status.Version = client.GetVersion()
    res.Status.LastSeen = time.Now().Format(time.RFC3339)

    return nil
}
```

---

## Workflow Engine Abstraction

### Configuration

```yaml
# config.yaml
reconciliation:
  enabled: true
  workflow_engine: "go-workflows"  # or "temporal"

  # go-workflows configuration
  go_workflows:
    db_path: "./workflows.db"
    worker_count: 10

  # Temporal configuration (alternative)
  temporal:
    host_port: "localhost:7233"
    namespace: "inventory"
    task_queue: "inventory-queue"
```

### Factory Pattern

```go
package workflows

import "github.com/spf13/viper"

// NewWorkflowManager creates workflow manager based on configuration
func NewWorkflowManager(cfg *viper.Viper) (WorkflowManager, error) {
    engine := cfg.GetString("reconciliation.workflow_engine")

    switch engine {
    case "go-workflows":
        dbPath := cfg.GetString("reconciliation.go_workflows.db_path")
        return NewGoWorkflowsManager(dbPath)

    case "temporal":
        hostPort := cfg.GetString("reconciliation.temporal.host_port")
        namespace := cfg.GetString("reconciliation.temporal.namespace")
        return NewTemporalManager(hostPort, namespace)

    default:
        return nil, fmt.Errorf("unknown workflow engine: %s", engine)
    }
}
```

### Engine Comparison

| Feature | go-workflows | Temporal |
|---------|-------------|----------|
| **Deployment** | Embedded (SQLite) | Distributed (requires server) |
| **Complexity** | Low | Medium |
| **Scalability** | Single instance | Highly scalable |
| **Visibility** | Basic | Advanced UI/monitoring |
| **Cost** | Free | Temporal Cloud or self-hosted |
| **Use Case** | Simple deployments | Production at scale |

---

## Event Bus Implementations

### Overview

The event bus abstraction supports multiple backends through CloudEvents-compliant implementations. Each backend provides different characteristics for deployment, scalability, and operational complexity.

### Configuration

```yaml
# config.yaml
events:
  # Backend type: memory, nats, kafka, redis
  backend: "memory"

  # CloudEvents settings
  cloudevents:
    version: "1.0"
    source_prefix: "/inventory"

  # In-memory configuration
  memory:
    buffer_size: 10000
    worker_count: 10

  # NATS configuration
  nats:
    url: "nats://localhost:4222"
    cluster_id: "inventory-cluster"
    durable_name: "inventory-durable"

    # JetStream settings (recommended for durability)
    jetstream:
      enabled: true
      stream_name: "INVENTORY_EVENTS"
      subjects:
        - "io.openchami.inventory.>"
      retention: "limits"  # limits, interest, or workqueue
      max_age: "168h"      # 7 days
      max_bytes: "10GB"
      replicas: 3

  # Kafka configuration
  kafka:
    brokers:
      - "localhost:9092"
      - "localhost:9093"
    topic: "inventory-events"
    consumer_group: "inventory-reconcilers"

    # Producer settings
    producer:
      compression: "snappy"
      max_message_bytes: 1048576
      required_acks: "all"

    # Consumer settings
    consumer:
      auto_offset_reset: "earliest"
      enable_auto_commit: true

  # Redis configuration (Streams)
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    stream_name: "inventory:events"
    consumer_group: "inventory-consumers"
    max_len: 100000
```

### Backend Comparison

| Feature | In-Memory | NATS JetStream | Kafka | Redis Streams |
|---------|-----------|----------------|-------|---------------|
| **Deployment** | Embedded | External server | External cluster | External server |
| **Durability** | None | Persistent | Persistent | Persistent |
| **Ordering** | FIFO | Per subject | Per partition | Per stream |
| **Scalability** | Single process | Multi-node cluster | Massive scale | Single/cluster |
| **Latency** | Microseconds | <1ms | <10ms | <5ms |
| **Complexity** | Minimal | Low-Medium | Medium-High | Low |
| **Replays** | No | Yes | Yes | Yes |
| **Filtering** | Basic | Subject-based | Topic-based | Key-based |
| **Best For** | Development/Testing | Production (small-medium) | Production (large) | Production (small) |

### Implementation Details

#### NATS JetStream Backend

```go
package events

import (
    "context"
    "encoding/json"

    "github.com/nats-io/nats.go"
    cloudevents "github.com/cloudevents/sdk-go/v2"
)

// NATSEventBus implements EventBus using NATS JetStream
type NATSEventBus struct {
    conn     *nats.Conn
    js       nats.JetStreamContext
    subs     map[string]*nats.Subscription
    handlers map[string]EventHandler
}

// NewNATSEventBus creates a new NATS-based event bus
func NewNATSEventBus(url string, config NATSConfig) (*NATSEventBus, error) {
    // Connect to NATS
    nc, err := nats.Connect(url)
    if err != nil {
        return nil, err
    }

    // Create JetStream context
    js, err := nc.JetStream()
    if err != nil {
        return nil, err
    }

    // Create or update stream
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     config.StreamName,
        Subjects: config.Subjects,
        Retention: nats.LimitsPolicy,
        MaxAge:   config.MaxAge,
        MaxBytes: config.MaxBytes,
        Replicas: config.Replicas,
    })
    if err != nil {
        return nil, err
    }

    return &NATSEventBus{
        conn:     nc,
        js:       js,
        subs:     make(map[string]*nats.Subscription),
        handlers: make(map[string]EventHandler),
    }, nil
}

// Publish publishes a CloudEvent to NATS
func (b *NATSEventBus) Publish(ctx context.Context, event cloudevents.Event) error {
    // Convert CloudEvent to JSON
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    // Publish to subject (derived from event type)
    subject := event.Type()
    _, err = b.js.Publish(subject, data)
    return err
}

// Subscribe subscribes to events matching a pattern
func (b *NATSEventBus) Subscribe(eventType string, handler EventHandler) (SubscriptionID, error) {
    // Convert event type pattern to NATS subject
    // io.openchami.inventory.bmc.* → io.openchami.inventory.bmc.>
    subject := convertToNATSSubject(eventType)

    // Create durable consumer
    sub, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
        // Parse CloudEvent
        var event cloudevents.Event
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            return
        }

        // Invoke handler
        if err := handler(context.Background(), event); err != nil {
            // Log error but ack message to prevent redelivery
            msg.Nak()
        } else {
            msg.Ack()
        }
    }, nats.Durable("inventory"), nats.ManualAck())

    if err != nil {
        return "", err
    }

    id := generateSubscriptionID()
    b.subs[id] = sub
    b.handlers[id] = handler

    return SubscriptionID(id), nil
}
```

#### Kafka Backend

```go
package events

import (
    "context"
    "encoding/json"

    "github.com/segmentio/kafka-go"
    cloudevents "github.com/cloudevents/sdk-go/v2"
)

// KafkaEventBus implements EventBus using Apache Kafka
type KafkaEventBus struct {
    writer   *kafka.Writer
    reader   *kafka.Reader
    handlers map[string]EventHandler
}

// NewKafkaEventBus creates a new Kafka-based event bus
func NewKafkaEventBus(brokers []string, topic string, config KafkaConfig) (*KafkaEventBus, error) {
    // Create Kafka writer (producer)
    writer := &kafka.Writer{
        Addr:         kafka.TCP(brokers...),
        Topic:        topic,
        Balancer:     &kafka.Hash{}, // Hash by event type for ordering
        Compression:  kafka.Snappy,
        RequiredAcks: kafka.RequireAll,
        Async:        false,
    }

    // Create Kafka reader (consumer)
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:        brokers,
        Topic:          topic,
        GroupID:        config.ConsumerGroup,
        MinBytes:       1,
        MaxBytes:       10e6, // 10MB
        CommitInterval: 1,     // Commit every second
    })

    return &KafkaEventBus{
        writer:   writer,
        reader:   reader,
        handlers: make(map[string]EventHandler),
    }, nil
}

// Publish publishes a CloudEvent to Kafka
func (b *KafkaEventBus) Publish(ctx context.Context, event cloudevents.Event) error {
    // Convert CloudEvent to JSON
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    // Use event type as key for partitioning (maintains ordering per event type)
    return b.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(event.Type()),
        Value: data,
    })
}

// Subscribe subscribes to events (Kafka doesn't support per-consumer filtering)
func (b *KafkaEventBus) Subscribe(eventType string, handler EventHandler) (SubscriptionID, error) {
    id := generateSubscriptionID()
    b.handlers[id] = handler

    // Start consumer goroutine (filters events in-process)
    go func() {
        for {
            msg, err := b.reader.ReadMessage(context.Background())
            if err != nil {
                continue
            }

            // Parse CloudEvent
            var event cloudevents.Event
            if err := json.Unmarshal(msg.Value, &event); err != nil {
                continue
            }

            // Filter by event type pattern
            if !matchesPattern(event.Type(), eventType) {
                continue
            }

            // Invoke handler
            if err := handler(context.Background(), event); err != nil {
                // Log error
            }
        }
    }()

    return SubscriptionID(id), nil
}
```

### Event Routing and Filtering

#### Pattern Matching

```go
// Pattern matching for event subscriptions
// Supports wildcards: * (single segment), ** (multiple segments)

func matchesPattern(eventType, pattern string) bool {
    // io.openchami.inventory.bmc.connected matches:
    // - io.openchami.inventory.bmc.connected (exact)
    // - io.openchami.inventory.bmc.* (wildcard)
    // - io.openchami.inventory.*.connected (wildcard)
    // - io.openchami.inventory.** (multi-segment)

    eventParts := strings.Split(eventType, ".")
    patternParts := strings.Split(pattern, ".")

    if len(patternParts) > len(eventParts) {
        return false
    }

    for i, p := range patternParts {
        if p == "*" {
            continue
        }
        if p == "**" {
            return true // Match everything after
        }
        if i >= len(eventParts) || p != eventParts[i] {
            return false
        }
    }

    return len(eventParts) == len(patternParts)
}
```

### CloudEvents HTTP Bridge

For external integrations, expose CloudEvents via HTTP:

```go
package events

import (
    "net/http"

    cloudevents "github.com/cloudevents/sdk-go/v2"
)

// HTTPEventBridge exposes CloudEvents over HTTP
type HTTPEventBridge struct {
    eventBus EventBus
    client   cloudevents.Client
}

func NewHTTPEventBridge(eventBus EventBus, port int) (*HTTPEventBridge, error) {
    // Create CloudEvents HTTP client
    p, err := cloudevents.NewHTTP(cloudevents.WithPort(port))
    if err != nil {
        return nil, err
    }

    c, err := cloudevents.NewClient(p)
    if err != nil {
        return nil, err
    }

    bridge := &HTTPEventBridge{
        eventBus: eventBus,
        client:   c,
    }

    // Start HTTP receiver
    go c.StartReceiver(context.Background(), bridge.receiveEvent)

    return bridge, nil
}

func (b *HTTPEventBridge) receiveEvent(ctx context.Context, event cloudevents.Event) error {
    // Forward to event bus
    return b.eventBus.Publish(ctx, event)
}

// POST /events - Publish event via HTTP
// GET /events/{uid}/stream - SSE stream of events for resource
```

### Observability and Monitoring

#### Metrics

```go
// Prometheus metrics for event bus
var (
    eventsPublished = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "inventory_events_published_total",
            Help: "Total number of events published",
        },
        []string{"type", "backend"},
    )

    eventsProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "inventory_events_processed_total",
            Help: "Total number of events processed",
        },
        []string{"type", "backend", "result"},
    )

    eventLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "inventory_event_latency_seconds",
            Help:    "Event processing latency",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
        },
        []string{"type", "backend"},
    )
)
```

#### Distributed Tracing

```go
// OpenTelemetry tracing for events
import "go.opentelemetry.io/otel/trace"

func (b *NATSEventBus) Publish(ctx context.Context, event cloudevents.Event) error {
    tracer := otel.Tracer("event-bus")
    ctx, span := tracer.Start(ctx, "event.publish",
        trace.WithAttributes(
            attribute.String("event.type", event.Type()),
            attribute.String("event.id", event.ID()),
        ),
    )
    defer span.End()

    // Inject trace context into CloudEvent
    event.SetExtension("traceparent", span.SpanContext().TraceID().String())

    // Publish event
    return b.doPublish(ctx, event)
}
```

### Migration Strategy

#### Phase 1: In-Memory (Development)
```yaml
events:
  backend: "memory"
  memory:
    buffer_size: 1000
```

#### Phase 2: NATS (Small Production)
```yaml
events:
  backend: "nats"
  nats:
    url: "nats://nats-cluster:4222"
    jetstream:
      enabled: true
      stream_name: "INVENTORY_EVENTS"
      replicas: 3
```

#### Phase 3: Kafka (Large Scale)
```yaml
events:
  backend: "kafka"
  kafka:
    brokers:
      - "kafka-1:9092"
      - "kafka-2:9092"
      - "kafka-3:9092"
    topic: "inventory-events"
    consumer_group: "inventory-reconcilers"
```

### Event Bus Factory

```go
package events

import "github.com/spf13/viper"

// NewEventBus creates an event bus based on configuration
func NewEventBus(cfg *viper.Viper) (EventBus, error) {
    backend := cfg.GetString("events.backend")

    switch backend {
    case "memory":
        return NewInMemoryEventBus(
            cfg.GetInt("events.memory.buffer_size"),
            cfg.GetInt("events.memory.worker_count"),
        )

    case "nats":
        return NewNATSEventBus(
            cfg.GetString("events.nats.url"),
            NATSConfig{
                StreamName: cfg.GetString("events.nats.jetstream.stream_name"),
                Subjects:   cfg.GetStringSlice("events.nats.jetstream.subjects"),
                MaxAge:     cfg.GetDuration("events.nats.jetstream.max_age"),
                Replicas:   cfg.GetInt("events.nats.jetstream.replicas"),
            },
        )

    case "kafka":
        return NewKafkaEventBus(
            cfg.GetStringSlice("events.kafka.brokers"),
            cfg.GetString("events.kafka.topic"),
            KafkaConfig{
                ConsumerGroup: cfg.GetString("events.kafka.consumer_group"),
            },
        )

    case "redis":
        return NewRedisEventBus(
            cfg.GetString("events.redis.addr"),
            cfg.GetString("events.redis.password"),
            cfg.GetInt("events.redis.db"),
            cfg.GetString("events.redis.stream_name"),
        )

    default:
        return nil, fmt.Errorf("unsupported event backend: %s", backend)
    }
}
```

---

## Implementation Plan

### Phase 1: Core Framework (2-3 weeks)

**Week 1: Event System**
- [ ] Implement `pkg/events/` package
  - [ ] Event struct and types
  - [ ] InMemoryEventBus implementation
  - [ ] Event pattern matching/routing
  - [ ] Unit tests
- [ ] Add event emission to storage layer
  - [ ] Emit events on Create/Update/Delete
  - [ ] Include before/after resource state

**Week 2: Reconciliation Framework**
- [ ] Implement `pkg/reconcile/` package
  - [ ] Reconciler interface
  - [ ] BaseReconciler with helpers
  - [ ] Controller implementation
  - [ ] WorkQueue implementation
- [ ] Integrate with events system
  - [ ] Watch for resource change events
  - [ ] Queue reconciliation requests

**Week 3: Workflow Abstraction**
- [ ] Implement `pkg/workflows/` package
  - [ ] WorkflowManager interface
  - [ ] go-workflows implementation
  - [ ] Temporal implementation (basic)
- [ ] Add workflow support to reconcilers
  - [ ] Long-running operations
  - [ ] Retry logic

### Phase 2: Code Generation (1-2 weeks)

**Week 4: Reconciler Templates**
- [ ] Create reconciler templates
  - [ ] Base reconciler template
  - [ ] Event handler template
  - [ ] Registration template
- [ ] Update code generator
  - [ ] Parse reconciliation annotations
  - [ ] Generate reconciler scaffolding
- [ ] Documentation
  - [ ] Template customization guide
  - [ ] Reconciler development guide

### Phase 3: HPC Implementation (2-3 weeks)

**Week 5-6: HPC Reconcilers**
- [ ] BMC Reconciler
  - [ ] Connect to BMC
  - [ ] Update connection status
  - [ ] Emit connection events
- [ ] FRU Reconciler
  - [ ] Listen to BMC connection events
  - [ ] Discover FRUs via Redfish
  - [ ] Update FRU inventory
- [ ] Node Reconciler
  - [ ] Monitor node health
  - [ ] Update boot configuration
  - [ ] Handle maintenance mode

**Week 7: Integration**
- [ ] Wire up reconciliation controller
- [ ] Add CLI flags for reconciliation config
- [ ] Performance testing
- [ ] Documentation

### Phase 4: Advanced Features (2-3 weeks)

**Week 8-9: Workflows**
- [ ] Complex provisioning workflows
- [ ] Multi-resource coordination
- [ ] Temporal integration refinement

**Week 10: Observability**
- [ ] Reconciliation metrics
- [ ] Event tracing
- [ ] Workflow visibility

---

## HPC Use Cases

### Use Case 1: Automatic BMC Discovery

**Scenario:** New BMC added to inventory, system automatically discovers it

```go
// BMC Reconciler
func (r *BMCReconciler) reconcileBMC(ctx context.Context, bmc *bmc.BMC) error {
    // Connect to BMC
    client, err := redfish.Connect(bmc.Spec.Address, bmc.Spec.Username, bmc.Spec.Password)
    if err != nil {
        bmc.Status.Connected = false
        return err
    }
    defer client.Close()

    // Update status
    bmc.Status.Connected = true
    bmc.Status.Version = client.GetFirmwareVersion()
    bmc.Status.LastSeen = time.Now().Format(time.RFC3339)

    // Emit connection event if newly connected
    if !wasConnected(bmc) {
        r.EmitEvent(ctx, "bmc.connected", bmc)
    }

    return nil
}

// FRU Reconciler listens for BMC connection events
func (r *FRUReconciler) OnBMCConnected(ctx context.Context, event events.Event) error {
    bmc := event.Resource.(*bmc.BMC)

    // Start FRU discovery workflow
    workflow := &DiscoverFRUsWorkflow{
        BMCUID:   bmc.GetUID(),
        Address:  bmc.Spec.Address,
        Username: bmc.Spec.Username,
        Password: bmc.Spec.Password,
    }

    _, err := r.WorkflowMgr.ExecuteWorkflow(ctx, workflow, nil)
    return err
}
```

### Use Case 2: FRU Inventory Synchronization

**Scenario:** Periodically sync FRU inventory with actual hardware

```go
// FRU Reconciler
func (r *FRUReconciler) reconcileFRU(ctx context.Context, fru *fru.FRU) error {
    // Get associated BMC
    bmc, err := r.Client.Get(ctx, "BMC", fru.Spec.Location.BMCUID)
    if err != nil {
        return err
    }

    // Connect and verify FRU still exists
    client := redfish.Connect(bmc.Spec.Address, ...)
    fruData, err := client.GetFRU(fru.Spec.RedfishPath)
    if err != nil {
        // FRU not found - mark as removed
        fru.Status.Present = false
        r.EmitEvent(ctx, "fru.removed", fru)
        return nil
    }

    // Update FRU status
    fru.Status.Present = true
    fru.Status.Health = fruData.Health
    fru.Status.LastVerified = time.Now()

    return nil
}
```

### Use Case 3: Node Health Monitoring

**Scenario:** Monitor node health and react to failures

```go
// Node Reconciler
func (r *NodeReconciler) reconcileNode(ctx context.Context, node *node.Node) error {
    // Check BMC connection
    bmc, _ := r.Client.Get(ctx, "BMC", node.Spec.BMCUID)
    if bmc.Status.Connected {
        node.Status.BMCReachable = true
    }

    // Check node power state
    powerState, err := r.checkPowerState(node)
    if err == nil {
        node.Status.PowerState = powerState

        // Emit event if power state changed
        if powerState != node.Status.PowerState {
            r.EmitEvent(ctx, "node.power.changed", node)
        }
    }

    // Update overall health condition
    healthy := node.Status.BMCReachable && node.Status.PowerState == "On"
    if healthy {
        r.SetCondition(node, "Healthy", "True", "AllChecksPass", "Node is healthy")
    } else {
        r.SetCondition(node, "Healthy", "False", "HealthCheckFailed", "Node health check failed")
        r.EmitEvent(ctx, "node.unhealthy", node)
    }

    return nil
}
```

### Use Case 4: Automated Provisioning Workflow

**Scenario:** Multi-step node provisioning with coordination

```go
// Provisioning Workflow
type ProvisionNodeWorkflow struct {
    NodeUID string
}

func (w *ProvisionNodeWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Step 1: Configure BMC
    if err := w.configureBMC(ctx); err != nil {
        return nil, err
    }

    // Step 2: Discover FRUs
    if err := w.discoverFRUs(ctx); err != nil {
        return nil, err
    }

    // Step 3: Configure boot parameters
    if err := w.configureBootConfig(ctx); err != nil {
        return nil, err
    }

    // Step 4: Power on node
    if err := w.powerOnNode(ctx); err != nil {
        return nil, err
    }

    // Step 5: Wait for node ready
    if err := w.waitForNodeReady(ctx); err != nil {
        return nil, err
    }

    return "provisioned", nil
}

// Trigger workflow via event
func (r *NodeReconciler) OnNodeCreated(ctx context.Context, event events.Event) error {
    node := event.Resource.(*node.Node)

    // Auto-provision if annotation present
    if node.Metadata.Annotations["auto-provision"] == "true" {
        workflow := &ProvisionNodeWorkflow{NodeUID: node.GetUID()}
        _, err := r.WorkflowMgr.ExecuteWorkflow(ctx, workflow, nil)
        return err
    }

    return nil
}
```

### Use Case 5: Maintenance Mode Coordination

**Scenario:** Node entering maintenance drains workloads and updates boot config

```go
// Node Reconciler
func (r *NodeReconciler) reconcileNode(ctx context.Context, node *node.Node) error {
    maintenanceMode := node.Spec.MaintenanceMode

    if maintenanceMode && !node.Status.InMaintenance {
        // Entering maintenance mode
        workflow := &MaintenanceModeWorkflow{
            NodeUID: node.GetUID(),
            Action:  "enter",
        }
        _, err := r.WorkflowMgr.ExecuteWorkflow(ctx, workflow, nil)
        if err != nil {
            return err
        }

        node.Status.InMaintenance = true
        r.EmitEvent(ctx, "node.maintenance.entered", node)
    } else if !maintenanceMode && node.Status.InMaintenance {
        // Exiting maintenance mode
        workflow := &MaintenanceModeWorkflow{
            NodeUID: node.GetUID(),
            Action:  "exit",
        }
        _, err := r.WorkflowMgr.ExecuteWorkflow(ctx, workflow, nil)
        if err != nil {
            return err
        }

        node.Status.InMaintenance = false
        r.EmitEvent(ctx, "node.maintenance.exited", node)
    }

    return nil
}
```

---

## API Reference

### Server Configuration

```go
// cmd/server/main.go
func runServer(cmd *cobra.Command, args []string) {
    // ... existing setup ...

    // Create event bus
    eventBus := events.NewInMemoryEventBus()

    // Create workflow manager
    workflowMgr, err := workflows.NewWorkflowManager(viper.GetViper())
    if err != nil {
        log.Fatalf("Failed to create workflow manager: %v", err)
    }
    defer workflowMgr.Close()

    // Create reconciliation controller
    controller := reconcile.NewController(eventBus, storageBackend)

    // Register reconcilers (generated code)
    if err := RegisterReconcilers(controller, client, eventBus); err != nil {
        log.Fatalf("Failed to register reconcilers: %v", err)
    }

    // Start controller
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := controller.Start(ctx); err != nil {
        log.Fatalf("Failed to start reconciliation controller: %v", err)
    }

    // ... start server ...
}
```

### CLI Flags

```bash
# Start server with reconciliation enabled
./bin/server \
  --reconcile-enabled=true \
  --reconcile-engine=go-workflows \
  --reconcile-interval=5m

# Start with Temporal
./bin/server \
  --reconcile-enabled=true \
  --reconcile-engine=temporal \
  --temporal-host=localhost:7233

# Disable reconciliation (API-only mode)
./bin/server --reconcile-enabled=false
```

### REST API Extensions

```bash
# Trigger manual reconciliation
POST /api/v1/bmcs/{uid}/reconcile

# View reconciliation status
GET /api/v1/bmcs/{uid}/reconciliation-status
{
  "lastReconcile": "2025-10-02T10:30:00Z",
  "status": "Success",
  "message": "Reconciliation successful",
  "nextReconcile": "2025-10-02T10:35:00Z"
}

# View events for resource
GET /api/v1/bmcs/{uid}/events
{
  "events": [
    {
      "id": "evt-123",
      "type": "bmc.connected",
      "timestamp": "2025-10-02T10:30:00Z",
      "message": "BMC successfully connected"
    }
  ]
}

# Trigger workflow
POST /api/v1/workflows/provision-node
{
  "nodeUID": "nd-456",
  "parameters": {
    "autoStart": true
  }
}
```

---

## Examples

### Example 1: Minimal BMC Reconciler

```go
package reconcilers

import (
    "context"
    "time"

    "github.com/openchami/inventory/pkg/resources/bmc"
    "github.com/openchami/inventory-framework/pkg/reconcile"
)

type BMCReconciler struct {
    reconcile.BaseReconciler
}

func (r *BMCReconciler) reconcileBMC(ctx context.Context, b *bmc.BMC) error {
    // Simple health check
    if err := pingBMC(b.Spec.Address); err != nil {
        b.Status.Reachable = false
        return err
    }

    b.Status.Reachable = true
    b.Status.LastSeen = time.Now().Format(time.RFC3339)
    return nil
}
```

### Example 2: Event-Driven FRU Discovery

```go
package reconcilers

import (
    "context"

    "github.com/openchami/inventory/pkg/crawler"
    "github.com/openchami/inventory-framework/pkg/events"
)

// RegisterEventHandlers registers FRU event handlers
func RegisterFRUEventHandlers(eventBus events.EventBus, reconciler *FRUReconciler) error {
    // React to BMC connection
    _, err := eventBus.Subscribe("bmc.connected", func(ctx context.Context, event events.Event) error {
        bmc := event.Resource.(*bmc.BMC)
        return reconciler.DiscoverFRUs(ctx, bmc)
    })
    return err
}

func (r *FRUReconciler) DiscoverFRUs(ctx context.Context, bmc *bmc.BMC) error {
    // Use crawler to discover FRUs
    client, _ := redfish.Connect(bmc.Spec.Address, bmc.Spec.Username, bmc.Spec.Password)
    fruSpecs, err := crawler.CoerceAll(client, bmc.GetUID(), "")
    if err != nil {
        return err
    }

    // Create FRU resources
    for _, spec := range fruSpecs {
        fru := &fru.FRU{Spec: spec}
        if err := r.Client.Create(ctx, fru); err != nil {
            r.Logger.Errorf("Failed to create FRU: %v", err)
        }
    }

    return nil
}
```

### Example 3: Long-Running Workflow

```go
package workflows

import (
    "context"
    "time"
)

type FirmwareUpdateWorkflow struct {
    BMCUID         string
    FirmwareURL    string
    VerifyChecksum bool
}

func (w *FirmwareUpdateWorkflow) Name() string {
    return "firmware-update"
}

func (w *FirmwareUpdateWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Step 1: Download firmware
    if err := w.downloadFirmware(ctx); err != nil {
        return nil, err
    }

    // Step 2: Verify checksum (if enabled)
    if w.VerifyChecksum {
        if err := w.verifyChecksum(ctx); err != nil {
            return nil, err
        }
    }

    // Step 3: Upload to BMC
    if err := w.uploadToBMC(ctx); err != nil {
        return nil, err
    }

    // Step 4: Trigger update
    if err := w.triggerUpdate(ctx); err != nil {
        return nil, err
    }

    // Step 5: Wait for BMC reboot (with timeout)
    time.Sleep(2 * time.Minute)

    // Step 6: Verify update
    if err := w.verifyUpdate(ctx); err != nil {
        return nil, err
    }

    return "updated", nil
}
```

---

## Testing Strategy

### Unit Tests

```go
// Test reconciler logic
func TestBMCReconciler_Reconcile(t *testing.T) {
    reconciler := &BMCReconciler{
        BaseReconciler: reconcile.BaseReconciler{
            Client:   mockClient,
            EventBus: mockEventBus,
        },
    }

    bmc := &bmc.BMC{
        Spec: bmc.BMCSpec{Address: "10.0.0.1"},
    }

    result, err := reconciler.Reconcile(context.Background(), bmc)
    assert.NoError(t, err)
    assert.True(t, result.Requeue)
}
```

### Integration Tests

```go
// Test event flow
func TestEventFlow_BMCConnectionTriggersDiscovery(t *testing.T) {
    // Setup
    eventBus := events.NewInMemoryEventBus()
    controller := reconcile.NewController(eventBus, storage)

    // Register reconcilers
    controller.RegisterReconciler(bmcReconciler)
    controller.RegisterReconciler(fruReconciler)

    // Start controller
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    controller.Start(ctx)

    // Create BMC
    bmc := &bmc.BMC{Spec: bmc.BMCSpec{Address: "10.0.0.1"}}
    storage.Save(ctx, "BMC", bmc.GetUID(), bmc)

    // Wait for reconciliation and FRU discovery
    time.Sleep(1 * time.Second)

    // Verify FRUs created
    frus, _ := storage.LoadAll(ctx, "FRU")
    assert.Greater(t, len(frus), 0)
}
```

---

## Migration Path

### For Existing Deployments

**Phase 1: Add reconciliation (opt-in)**
```bash
# Default: reconciliation disabled for backward compatibility
./bin/server

# Enable reconciliation explicitly
./bin/server --reconcile-enabled=true
```

**Phase 2: Gradual rollout**
- Deploy with reconciliation enabled
- Monitor metrics and logs
- Verify status updates happen automatically

**Phase 3: Full adoption**
- Make reconciliation default
- Remove manual status update code
- Add custom event handlers

---

## Performance Considerations

### Scaling

**Single Instance (go-workflows):**
- 100s of resources: Excellent
- 1000s of resources: Good
- 10000+ resources: Consider Temporal

**Distributed (Temporal):**
- Unlimited horizontal scaling
- Multiple worker instances
- Workflow sharding

### Optimization Strategies

1. **Reconciliation Intervals:**
   ```go
   // Fast reconciliation for critical resources
   bmcReconciler.SetInterval(30 * time.Second)

   // Slower for stable resources
   fruReconciler.SetInterval(15 * time.Minute)
   ```

2. **Event Batching:**
   ```go
   // Batch events to reduce reconciliation load
   eventBus.SetBatchSize(100)
   eventBus.SetBatchInterval(5 * time.Second)
   ```

3. **Rate Limiting:**
   ```go
   // Limit reconciliation rate per resource type
   controller.SetRateLimit("BMC", 10) // 10 per second
   ```

---

## Future Enhancements

### Phase 2 Features

- **Webhook Support:** Trigger reconciliation via webhooks
- **Scheduled Reconciliation:** Cron-style reconciliation schedules
- **Leader Election:** Multi-instance reconciliation with leader election
- **Event Persistence:** Durable event storage (Redis, Kafka)
- **Observability:** Prometheus metrics, OpenTelemetry tracing

### Phase 3 Features

- **Custom Workflow Languages:** Support for workflow DSLs
- **Conditional Reconciliation:** Only reconcile if conditions met
- **Resource Dependencies:** Reconcile in dependency order
- **Rollback Workflows:** Automatic rollback on failure

---

## Conclusion

This proposal introduces a **lightweight, pluggable reconciliation and events system** that:

✅ **Enables declarative infrastructure management** (Spec → Status)
✅ **Provides event-driven architecture** for reactive behaviors
✅ **Supports both embedded and distributed workflows** (go-workflows & Temporal)
✅ **Minimizes boilerplate** through code generation
✅ **Solves real HPC problems** (BMC discovery, FRU sync, health monitoring)

**Next Steps:**
1. Review and approve proposal
2. Create `pkg/events/`, `pkg/reconcile/`, `pkg/workflows/` packages
3. Update code generator with reconciler templates
4. Implement BMC reconciler as proof-of-concept
5. Document reconciler development guide

**Estimated Timeline:** 6-8 weeks for complete implementation

---

## Appendix A: Configuration Reference

```yaml
# Full configuration example
reconciliation:
  enabled: true
  workflow_engine: "go-workflows"  # or "temporal"

  # Global settings
  default_interval: "5m"
  max_concurrent_reconciles: 10

  # go-workflows settings
  go_workflows:
    db_path: "./data/workflows.db"
    worker_count: 10
    max_workflow_runtime: "1h"

  # Temporal settings
  temporal:
    host_port: "localhost:7233"
    namespace: "inventory"
    task_queue: "inventory-queue"
    worker_count: 20
    max_concurrent_workflows: 100

  # Per-resource configuration
  resources:
    BMC:
      interval: "30s"
      max_retries: 3
      timeout: "1m"
    FRU:
      interval: "15m"
      max_retries: 1
    Node:
      interval: "1m"
      max_retries: 5

# Event bus configuration
events:
  type: "memory"  # or "redis", "kafka"
  buffer_size: 1000

  # Redis settings (optional)
  redis:
    addr: "localhost:6379"
    db: 0
```

## Appendix B: Metrics & Observability

```go
// Prometheus metrics
reconcile_duration_seconds{resource="BMC",result="success"}
reconcile_total{resource="BMC",result="success"}
reconcile_errors_total{resource="BMC"}
event_published_total{type="bmc.connected"}
event_handled_total{type="bmc.connected",result="success"}
workflow_duration_seconds{workflow="provision-node"}
workflow_total{workflow="provision-node",status="completed"}
```

---

## Appendix C: CloudEvents & Event Bus Summary

### CloudEvents Adoption Benefits

**Interoperability:**
- ✅ CNCF-standard event format
- ✅ Works with existing CloudEvents tooling
- ✅ Language and protocol agnostic
- ✅ HTTP, NATS, Kafka, AMQP support out-of-box

**Event Structure Example:**
```json
{
  "specversion": "1.0",
  "type": "io.openchami.inventory.bmc.connected",
  "source": "/inventory/bmcs/bmc-abc123",
  "id": "evt-xyz789",
  "time": "2025-10-02T15:30:00Z",
  "datacontenttype": "application/json",
  "data": {
    "uid": "bmc-abc123",
    "address": "10.0.0.1",
    "version": "2.1.0"
  },
  "inventoryresourcekind": "BMC",
  "inventoryresourceuid": "bmc-abc123"
}
```

### Event Bus Backend Support

| Backend | Deployment | Durability | Latency | Best For |
|---------|-----------|------------|---------|----------|
| **In-Memory** | Embedded | None | µs | Development |
| **NATS JetStream** | External | Persistent | <1ms | Production (small-medium) |
| **Kafka** | Cluster | Persistent | <10ms | Production (large scale) |
| **Redis Streams** | External | Persistent | <5ms | Production (small) |

### Key Design Decisions

**1. CloudEvents Standard:**
- Use CloudEvents SDK for Go
- Reverse-DNS event type naming (`io.openchami.inventory.*`)
- Extension attributes for inventory-specific metadata
- Compatible with HTTP, NATS, Kafka transports

**2. Pluggable Event Bus:**
- Abstract `EventBus` interface
- Factory pattern for backend selection
- Configuration-driven backend choice
- Support for wildcard subscriptions

**3. External Integration:**
- CloudEvents HTTP bridge for webhooks
- Server-Sent Events (SSE) for streaming
- OpenTelemetry tracing support
- Prometheus metrics

### Migration Path

**Development → Production:**
```yaml
# Development (in-memory)
events:
  backend: "memory"

# Small production (NATS)
events:
  backend: "nats"
  nats:
    jetstream:
      enabled: true
      replicas: 3

# Large scale (Kafka)
events:
  backend: "kafka"
  kafka:
    brokers: ["kafka-1:9092", "kafka-2:9092"]
```

### Integration Examples

**NATS Example:**
```bash
# Subscribe to all inventory events
nats sub "io.openchami.inventory.>"

# Subscribe to BMC events only
nats sub "io.openchami.inventory.bmc.*"
```

**Kafka Example:**
```bash
# Consume from inventory events topic
kafka-console-consumer --topic inventory-events \
  --bootstrap-server localhost:9092
```

**HTTP Webhook:**
```bash
# External system publishes event
curl -X POST http://inventory:8080/events \
  -H "Ce-Specversion: 1.0" \
  -H "Ce-Type: io.openchami.inventory.bmc.created" \
  -H "Ce-Source: /external/system" \
  -H "Content-Type: application/json" \
  -d '{"uid": "bmc-xyz", "address": "10.0.0.5"}'
```

---

**Document Version:** 1.1
**Last Updated:** 2025-10-02
**Authors:** System Architecture Team
**Status:** Proposed (Updated with CloudEvents & Event Bus Backend Support)
