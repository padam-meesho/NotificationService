# Go Codebase Analysis & Best Practices

A compilation of learnings from reviewing various Go codebases, focusing on architectural patterns, best practices, and implementation strategies for microservices.

---

## 🏗️ Architecture & Design Patterns

### Q: How do services interact with DAOs and what design patterns are commonly used?

**Service-DAO Interaction Pattern:**
- We define a services struct with the DAO object as a field
- Services are typically structs that hold references to DAOs or repositories as fields
- These references are set during service initialization, often via constructor functions (e.g., NewOfferOperationsService)
- The DAOs themselves encapsulate the logic for interacting with databases (Scylla, Redis, etc.)

**Constructor Injection & Singleton Pattern:**
This is all part of Constructor Injection and singleton pattern usage. DAOs are structs that encapsulate database/cache clients (e.g., Redis, Scylla session) as fields. When creating an object of a DAO type, we use a `sync.Once` variable's `Do` function to ensure that initialization is done only once.

### Q: What's the typical control flow for initialization in Go applications?

```
main.go -> app.NewApp() -> Initialize all components
```

All the initialization functions are standalone functions. If a function is standalone, then it is basically a part of the package in which it is defined. If a function is a part of a struct, then it can be accessed by creating an object of the struct.

### Q: How do DAOs typically get initialized and accessed?

Each DAO typically has a constructor function (commonly named `NewXxxDao` or `NewXxxRepository`), which:
- Accepts configuration or client/session as arguments
- Initializes the struct with the required dependencies
- Uses `sync.Once` to ensure singleton behavior

For any DAO:
1. Have a struct that has a DAO client
2. Have an interface which has all the methods that the DAO needs to implement
3. Create an object of the DAO struct (within the package so it's accessible to other global or standalone functions)
4. Define all the methods of the interface for the struct (this automatically implements the interface in Go)
5. Have a global method (standalone, belongs to the package) for struct access by other services
6. Call `<package_name>.<global_method>` which initializes the DAO client (Redis, Scylla, or Kafka)

---

## 🔧 Go Language Fundamentals & Patterns

### Q: How does Go compare to Java in terms of structure and patterns?

**Correlation with Java:**
- **Standalone functions** → Static methods
- **Package** → Class
- **Struct** → Class
- **Methods** → Non-static functions

### Q: What are the naming conventions for interfaces and structs in Go?

**Naming Conventions:**
- **Interface**: Normal name + "Dao/Service/Controller" (e.g., `GlobalUserOfferRedemptionDao`)
- **Struct**: InterfaceName + "Impl" (e.g., `GlobalUserOfferRedemptionDaoImpl`)
- Have a global constructor method (belongs to the package)
- Have a global getter method (belongs to the package)

### Q: How do Go interfaces work and what makes them unique?

In Go, it doesn't matter if a struct has extra fields (or even extra methods) beyond what the interface specifies.

**How Go Interfaces Work:**
- Interfaces only care about methods
- If a struct has all the methods required by an interface (with the correct signatures), it implements that interface—regardless of any extra fields or methods
- In Go, we don't need to mention inheritance or implementation explicitly

### Q: Can struct methods access global variables within a package?

Yes, a struct method in Go can access global (package-level) variables. Any function or method defined in a package can access all variables declared at the package level, as long as they are in the same package and are not shadowed by local variables.

```go
// Example
package example

var globalCounter int // package-level variable

type MyStruct struct{}

func (m *MyStruct) IncrementGlobal() {
    globalCounter++ // Accessing and modifying the global variable
}
```

### Q: What is sync.Once and when should it be used?

`sync.Once` is used in Go to ensure singleton initialization. When the new app is being initialized, we call the new or constructor method, but when we need a client to create a DAO object, we call the get client method, since the constructor would have already initialized the global object by then.

---

## 💾 Database Implementation & Patterns

### Q: How is ScyllaDB typically integrated in Go applications?

**ScyllaDB (Cassandra) Usage:**
The Scylla session is created and initialized at application startup. DAOs that need Scylla access have a Session field. The session is set when the DAO is constructed, usually via a singleton getter or constructor.

**Query Building Pattern:**
We use the `qb` (query builder) package for ScyllaDB operations. There are various functions such as `select`, `insert`. The methods which have "release" in their name are used to execute the query; other methods are for building the query.

### Q: What naming conventions should be followed for database schemas?

**Naming Conventions:**
- Tables: lowercase and underscore separated
- Be careful about verb/tense confusion in naming

### Q: How should models be organized in a Go project?

Models should be divided into different types:
1. **Entity**: Used for defining database schemas
2. **Domain**: Used for core business objects or entities
3. **DTO**: Used for transferring data between layers
4. **Request**: Represents the structure of incoming API requests
5. **Response**: Represents the structure of outgoing API responses

---

## ⚙️ Configuration Management & Environment Setup

### Q: How should configuration be handled in Go applications?

**Viper Usage:**
We use Viper to read environment variables and process them into a struct. This provides a clean way to manage configuration across different environments.

### Q: What are the benefits of using Docker in Go applications?

**Docker Integration Benefits:**

Without Docker - you need to install everything locally:
```bash
brew install kafka
brew install redis
go install ...
go run main.go
```

With Docker - everything is packaged:
```bash
docker-compose up  # Kafka, Redis, your app all start together
```

---

## 🚀 Kafka & Message Processing Patterns

### Q: What's a good approach for Kafka payload design?

**Message Processing Philosophy:**
The Kafka payload should contain only the request ID. The consumer is expected to fetch the rest of the details from the different DAOs. This maintains loose coupling and better scalability.

**Kafka Payload Structure:**
```json
{
  "Type": "SMS_REQUEST",
  "Data": {
    "MessageId": "unique-message-id"
  }
}
```

### Q: How can you run multiple servers on the same port?

**Server Architecture with CMUX:**
You can have both HTTP and gRPC servers running on the same port using Connection Multiplexing (CMUX):
- A single TCP listener is created on the configured port
- The listener is wrapped with cmux for protocol matching:
  - `grpcListener := mux.Match(cmux.HTTP2())` - Matches HTTP/2 traffic (gRPC)
  - `httpListener := mux.Match(cmux.Any())` - Matches other traffic (HTTP/1.x)

---

## 📝 Logging & Best Practices

### Q: What are the key principles for effective logging in Go applications?

**Good Practices:**
- Try to log specific errors in the service layer
- Log generic errors with the request_id of the failed payload in the Kafka layer
- Maintain a consistent format for all logs - this is good practice and makes it easier for parsing/debugging

### Q: How should context be managed in Go applications?

**Context Management:**
Context should be created directly at the entry point (as top of the chain as possible) and propagated through the application layers.

### Q: How can you implement request tracing in Go applications?

**Trace ID Implementation:**
Use middleware to generate a traceID for each incoming request and add it to the headers. You can create a custom logger function wrapper around zerolog to print logs with trace details for better request tracking.

---

## 📊 Structured Logging Architecture Implementation

### Q: What logging utilities should be implemented for enterprise-grade applications?

This project implements enterprise-grade structured logging using **Zerolog** with custom logging utilities for consistent, searchable, and machine-readable logs across all components.

**Logging Utilities (`internal/utils/logger.go`):**

1. **ComponentLogger(component string)** - For general component-level operations without request context
2. **RequestLogger(ctx, component, operation string)** - For request-scoped operations with trace ID propagation
3. **DatabaseLogger(ctx, operation, table, requestID string)** - Specialized for database operations with enhanced context
4. **KafkaLogger(operation, topic string)** - Specialized for Kafka operations with topic context

### Q: What fields should be standardized across all logs?

**Common Structured Fields:**
| Field | Description | Example |
|-------|-------------|---------|
| `component` | System component | `"kafka"`, `"scylla"`, `"service"` |
| `operation` | Action being performed | `"insert"`, `"produce"`, `"send_sms"` |
| `trace_id` | Request tracing ID | `"abc-123-def-456"` |
| `request_id` | SMS request identifier | `"uuid-request-id"` |
| `phone_number` | Target phone number | `"1234567890"` |
| `level` | Log severity | `"info"`, `"error"`, `"warn"` |
| `time` | Timestamp | `"2024-01-01T10:00:00Z"` |
| `message` | Human-readable message | `"SMS request processed successfully"` |

### Q: How should different log levels be used?

**Log Level Guidelines:**
- **Info Level**: Normal operations, successful actions, service initialization, request processing milestones
- **Error Level**: Operation failures with detailed context, database connection issues, external service failures
- **Warn Level**: Business logic warnings, blacklisted operations, rate limiting, degraded performance
- **Debug Level**: Detailed operation traces, cache hits/misses, internal state changes, performance metrics

---

## 🔄 Project-Specific Implementation Notes

### Q: What are the current tasks and improvements for this SMS notification service?

**Current Tasks:**
- [x] Log correction and generalization across the system
- [x] Service verification - ensure all components are running and building correctly
- [ ] Define the schema for the RDBMS database
- [x] Integrate Redis DAO methods into the service layer
- [x] Implement global config struct for parsing configs (stored in .yml file)

**Future Improvements:**
- [ ] Check and understand the complete Docker setup significance
- [x] Organize the README.md documentation
- [x] Use traceId in logs for better tracking
- [ ] Organize Go language learning materials

### Q: What specific architectural decisions were made for this SMS service?

**Core Architecture Components:**
- **API Layer**: HTTP endpoints with Gin framework
- **Message Queue**: Kafka for asynchronous SMS processing
- **Database**: ScyllaDB for SMS request storage and tracking
- **Cache**: Redis for phone number blacklisting
- **Containerization**: Docker Compose for service orchestration

**Key Implementation Details:**
- Services are structs that have DAOs as their fields
- Since the service is a struct, it can be initialized inside a controller
- Field values inside the service are basically DAO objects, through which their methods can be accessed
- DAO internally is a struct which has different Redis clients or Scylla sessions as fields, so DAO methods can access them

---

## 📈 Monitoring & Observability Patterns

### Q: How can structured logging support monitoring and alerting?

**Log Aggregation Integration:**
- **ELK Stack**: Easy integration with Elasticsearch for searching
- **Prometheus**: Metrics extraction from structured logs
- **Grafana**: Dashboard creation using log-based metrics
- **Jaeger**: Distributed tracing using trace IDs

**Search Patterns:**
```bash
# Find all errors for a specific request
component:"service" AND request_id:"uuid-456-789" AND level:"error"

# Monitor Kafka operations
component:"kafka" AND operation:"produce"

# Database performance analysis
component:"database" AND table:"sms_requests" AND operation:"insert"

# Trace complete request flow
trace_id:"abc-123-def"
```

This structured approach ensures enterprise-grade observability, making debugging, monitoring, and performance analysis significantly more efficient and effective.



