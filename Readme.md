# Notification Service

A scalable Go-based SMS notification service built with microservices architecture, featuring asynchronous message processing using Kafka, data persistence with ScyllaDB, and Redis-based blacklisting.

##  Architecture

- **API Layer**: Gin HTTP framework with middleware for authentication and request tracing
- **Message Queue**: Kafka for asynchronous SMS processing
- **Database**: ScyllaDB for SMS request storage and tracking
- **Cache**: Redis for phone number blacklisting
- **Containerization**: Docker Compose for easy deployment

##  Features

- **SMS Management**: Send, track, and retrieve SMS request details
- **Blacklist Management**: Add/remove phone numbers from blacklist
- **Asynchronous Processing**: Kafka-based message queuing for reliability
- **Request Tracing**: UUID-based tracking for all requests
- **Authentication**: Bearer token-based API security
- **Health Checks**: Service health monitoring endpoints

##  Prerequisites

- Go 1.24+ 
- Docker & Docker Compose
- Git

## 🛠️ Setup Instructions

### 1. Clone the Repository
```bash
git clone <your-repo-url>
cd NotificationService
```

### 2. Install Dependencies
```bash
go mod download
```

### 3. Start Infrastructure Services
```bash
# Start all services (Kafka, ScyllaDB, Redis)
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 4. Create ScyllaDB Schema
```bash
# Connect to ScyllaDB
docker exec -it scylla cqlsh

# Create keyspace and table
CREATE KEYSPACE IF NOT EXISTS notificationservice 
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE notificationservice;

CREATE TABLE IF NOT EXISTS sms_requests (
    id TEXT PRIMARY KEY,
    phone_number TEXT,
    message TEXT,
    status TEXT,
    failure_code TEXT,
    failure_comments TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

exit;
```

### 5. Configuration
Create `configs/app_config.yaml`:
```yaml
kafka:
  bootstrapservers: "localhost:9092"
  groupid: "notification-service-group"
  autooffsetreset: "earliest"

redis:
  addr: "localhost:6379"
  db: 0
  pwd: ""

scylla:
  hosts: "localhost"
  keyspace: "notificationservice"
```

### 6. Run the Application
```bash
go run cmd/main.go
```

The service will start on `http://localhost:3333`

## 📡 API Endpoints

### Authentication
All endpoints require `Authorization: Bearer password123` header.

### SMS Operations

#### Send SMS
```bash
POST /v1/sms/send
Content-Type: application/json

{
    "phone_number": "1234567890",
    "message": "Hello World!"
}
```

**Response:**
```json
{
    "request_id": "uuid-here",
    "message": "message sent successfully!"
}
```

#### Get SMS Details
```bash
GET /v1/sms/{request_id}
```

### Blacklist Operations

#### Get Blacklisted Numbers
```bash
GET /v1/blacklist
```

#### Add to Blacklist
```bash
POST /v1/blacklist
Content-Type: application/json

{
    "phone_numbers": "1234567890"
}
```

#### Remove from Blacklist
```bash
DELETE /v1/blacklist/{phone_number}
```

### Health Check
```bash
GET /health
```

## Testing

### Basic Health Check
```bash
curl http://localhost:3333/health
```

### Send Test SMS
```bash
curl -X POST http://localhost:3333/v1/sms/send \
  -H "Authorization: Bearer password123" \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "1234567890", "message": "Test message"}'
```

### Verify Data in ScyllaDB
```bash
docker exec -it scylla cqlsh -e "USE notificationservice; SELECT * FROM sms_requests;"
```

### Check Kafka Messages
```bash
# View producer logs
docker logs kafka

# Monitor consumer logs in application output
```

##  Development

### Project Structure
```
├── cmd/                    # Application entry points
├── config/                 # Configuration management
├── dao/                    # Data Access Objects
├── internal/
│   ├── handlers/          # HTTP request handlers
│   ├── middlewares/       # HTTP middlewares
│   ├── models/           # Data models
│   ├── repo/             # Service layer
│   └── utils/            # Utility functions
├── kafka/                 # Kafka DAO implementation
└── docker-compose.yml     # Infrastructure setup
```

### Adding New Features
1. Define models in `internal/models/`
2. Create DAO methods in appropriate `dao/` files
3. Implement business logic in `internal/repo/`
4. Add HTTP handlers in `internal/handlers/`
5. Register routes in `internal/routes/`

##  Troubleshooting

### Common Issues

**Service won't start:**
- Ensure Docker services are running: `docker-compose ps`
- Check port conflicts (3333, 9042, 6379, 9092)

**Database connection errors:**
- Verify ScyllaDB is healthy: `docker-compose logs scylla`
- Ensure keyspace and table exist

**Kafka issues:**
- Check Kafka logs: `docker-compose logs kafka`
- Verify topic creation: Consumer logs should show subscription

**No data in database:**
- Check application logs for insert errors
- Verify table schema matches model structure

### Logs and Monitoring
- Application logs show detailed request/response information
- Kafka consumer logs show message processing status
- Database operation results are logged with request IDs

##  Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request
