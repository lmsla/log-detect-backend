# Log Detect Backend - Quick Reference Guide

## Project Overview

**Name**: Log Detect Backend
**Purpose**: Centralized log monitoring, device health tracking, and anomaly detection system
**Language**: Go 1.24.0
**Framework**: Gin Web Framework
**Databases**: MySQL, TimescaleDB, Elasticsearch
**Port**: 8006

---

## Key Features at a Glance

| Feature | Purpose | Key Files |
|---------|---------|-----------|
| **Device Monitoring** | Detect online/offline status | detect.go, center.go |
| **Elasticsearch Health** | Monitor ES cluster health | es_monitor_service.go, es_scheduler.go |
| **Email Alerts** | Send notifications | mail.go, detect.go |
| **User Management** | Authentication & RBAC | auth.go, middleware/auth.go |
| **Time-Series Data** | Store metrics efficiently | batch_writer.go (TimescaleDB) |
| **Dashboard** | Visualization & Analytics | dashboard.go, controller/dashboard.go |
| **REST API** | HTTP interface | router/router.go, controller/* |
| **Cron Jobs** | Scheduled monitoring | center.go, services/detect.go |

---

## Core Data Models

### Monitoring Entities
```
Target (monitoring config)
  ├─ Indices (many-to-many)
  │   ├─ Pattern: ES index pattern
  │   ├─ Logname: log type name
  │   ├─ Period: minutes/hours
  │   └─ Unit: frequency value
  └─ Receivers (email list)

Device
  ├─ DeviceGroup: logical grouping
  └─ Name: device hostname

History (monitoring result)
  ├─ Status: online/offline/warning/error
  ├─ Lost: true/false
  ├─ DateTime: when checked
  ├─ ResponseTime: milliseconds
  └─ DataCount: records found
```

### Authentication Entities
```
User
  └─ Role (1-to-1 relationship)
      └─ Permissions (many-to-many)
          ├─ Resource: device, target, indices, elasticsearch
          └─ Action: create, read, update, delete
```

### Elasticsearch Monitoring
```
ElasticsearchMonitor (config)
  ├─ Host/Port/Auth
  ├─ CheckInterval: seconds
  ├─ Thresholds: CPU/Memory/Disk/Response
  └─ Receivers: email alerts

ESMetric (measurements)
  ├─ Time: timestamp
  ├─ CPU/Memory/Disk: percentages
  ├─ NodeCount: active nodes
  └─ UnassignedShards: shard status

ESAlert (warnings)
  ├─ Severity: critical/high/medium/low
  ├─ Status: active/resolved/acknowledged
  └─ Message: alert details
```

---

## Important File Locations

### Configuration Files
```
config.yml              # Main configuration
setting.yml            # Monitoring targets configuration
```

### Source Code Directories
```
controller/            # HTTP handlers (9 files)
services/              # Business logic (22 services)
entities/              # Data models (targets.go, elasticsearch.go, menu.go)
models/                # Response structures
middleware/            # Authentication & RBAC
clients/               # Database connections
router/                # Route definitions
```

### Database Schema
```
MySQL (logdetect database)
  ├─ users, roles, permissions (RBAC)
  ├─ devices, targets, indices (config)
  ├─ history, alert_history, mail_history (logs)
  └─ cron_lists, elasticsearch_monitors (management)

TimescaleDB (monitoring database)
  ├─ device_metrics (time-series)
  ├─ es_metrics (time-series)
  └─ alert_history (alerts)
```

---

## API Endpoints Quick List

### Authentication
```
POST   /auth/login
POST   /api/v1/auth/register
GET    /api/v1/auth/profile
POST   /api/v1/auth/refresh
```

### Device Management
```
GET    /api/v1/Device/GetAll
POST   /api/v1/Device/Create
PUT    /api/v1/Device/Update
DELETE /api/v1/Device/Delete/:id
GET    /api/v1/Device/GetGroup
GET    /api/v1/Device/count
```

### Target Management (Monitoring Configs)
```
GET    /api/v1/Target/GetAll
POST   /api/v1/Target/Create
PUT    /api/v1/Target/Update
DELETE /api/v1/Target/Delete/:id
```

### Elasticsearch Monitoring
```
GET    /api/v1/elasticsearch/monitors
POST   /api/v1/elasticsearch/monitors
GET    /api/v1/elasticsearch/status
GET    /api/v1/elasticsearch/alerts
POST   /api/v1/elasticsearch/monitors/:id/test
```

### Dashboard & Analytics
```
GET    /api/v1/dashboard/overview
GET    /api/v1/dashboard/statistics
GET    /api/v1/dashboard/trends
GET    /api/v1/dashboard/devices/:device_name/timeline
```

### History & Monitoring
```
GET    /api/v1/History/GetData/:logname
GET    /api/v1/History/GetLognameData
```

---

## Workflow Examples

### How Device Monitoring Works

1. **Configuration**
   - Admin creates Target (email recipients)
   - Admin creates Index (ES pattern + logname)
   - Admin links Index to Target

2. **Scheduling**
   - Control_center() reads all enabled Targets
   - For each Index, creates cron job expression
   - Registers with robfig/cron scheduler
   - Example: "*/5 * * * *" = every 5 minutes

3. **Execution**
   - Cron triggers Detect() function
   - Queries Elasticsearch for devices in timeframe
   - Compares with Device database
   - Creates History records for each device
   - Sends email for offline devices
   - Stores metrics in TimescaleDB (via BatchWriter)

### How ES Monitoring Works

1. **Configuration**
   - Admin creates ElasticsearchMonitor
   - Sets check interval (e.g., 60 seconds)
   - Sets alert thresholds

2. **Scheduling**
   - ESMonitorScheduler loads all enabled monitors
   - Creates ticker for each monitor
   - Starts background goroutine

3. **Health Check**
   - Periodically queries ES cluster health
   - Collects CPU, memory, disk, shard data
   - Compares against thresholds
   - Creates ESAlert if threshold exceeded
   - Sends email notification
   - Stores metrics in TimescaleDB

### How Batch Writing Works

1. **Accumulation**
   - Services generate History or ESMetric records
   - BatchWriter.AddHistory() queues them
   - Held in memory

2. **Flushing Trigger**
   - Size: >= 50 records
   - Time: 5 seconds elapsed (configurable)

3. **Batch Insert**
   - Begin transaction
   - Insert all records using prepared statement
   - Commit (atomic, all-or-nothing)
   - Log success count

---

## Key Services & Their Responsibilities

| Service | Main Functions | Key Logic |
|---------|---|---|
| **auth.go** | HashPassword, GenerateJWT, ValidateJWT, CheckPermission | RBAC enforcement, token management |
| **detect.go** | Detect, SearchRequest, GetDevicesDataByGroupName | Core monitoring: ES query, device comparison, alerts |
| **center.go** | ExecuteCrontab, Control_center | Cron job registration & bootstrap |
| **batch_writer.go** | AddHistory, flushDeviceMetrics, flushESMetrics | Time-series data batching & insertion |
| **es_monitor_service.go** | CreateESMonitor, UpdateESMonitor, GetAllESMonitors | ES monitor CRUD operations |
| **es_scheduler.go** | InitESScheduler, StartMonitor, LoadAllMonitors | ES health check scheduling |
| **mail.go** | Mail4 | SMTP email sending with HTML templates |
| **history.go** | CreateHistory, GetHistoryDataByDeviceName | Historical data management |
| **device.go** | CreateDevice, UpdateDevice, GetDevicesDataByGroupName | Device CRUD & grouping |
| **target.go** | CreateTarget, UpdateTarget, GetAllTargets | Target CRUD & orchestration |
| **indices.go** | CreateIndices, UpdateIndices, GetIndicesDataByLogname | Index pattern CRUD |
| **dashboard.go** | GetDashboardData, GetHistoryStatistics, GetTrendData | Analytics & visualization |

---

## Common Operations

### Start Monitoring a New Log Source

```go
// 1. Create Index (ES pattern definition)
Index{
  Pattern: "logstash-prod-*",
  Logname: "webserver",
  Period: "minutes",
  Unit: 5,           // Check every 5 minutes
  Field: "host.keyword"
}

// 2. Create Target (recipient config)
Target{
  Subject: "Web Server Monitoring",
  To: ["admin@company.com"],
  Enable: true,
  Indices: [...]  // Link the index
}

// 3. System automatically:
// - Registers cron job "*/5 * * * *"
// - Starts detect() execution
// - Stores results in History
// - Sends alerts for offline devices
```

### Create Elasticsearch Health Monitor

```go
ElasticsearchMonitor{
  Name: "Production ES Cluster",
  Host: "10.99.1.213",
  Port: 9200,
  Interval: 60,        // Check every 60 seconds
  CPUUsageHigh: 75.0,
  CPUUsageCritical: 85.0,
  MemoryUsageHigh: 80.0,
  MemoryUsageCritical: 90.0,
  DiskUsageHigh: 85.0,
  DiskUsageCritical: 95.0,
  Receivers: ["ops@company.com"],
  Subject: "ES Health Alert"
}

// System automatically:
// - Creates configuration in MySQL
// - Starts monitoring scheduler
// - Collects metrics to TimescaleDB
// - Sends alerts on threshold breach
```

### Query Device History

```
GET /api/v1/History/GetData/webserver?days=7

Returns:
[
  {
    date: "2024-10-29",
    logname: "webserver",
    name: "web-server-01",
    status: "online",
    time: "14:30:00",
    response_time: 245,
    data_count: 1523
  },
  ...
]
```

---

## Database Connection Details

### MySQL (GORM)
```
Host: 10.99.1.133
Port: 3306
Database: logdetect
User: runner
Password: (from config.yml)
Pool: Max Idle=10, Max Open=100
Lifetime: 1 hour
```

### TimescaleDB (PostgreSQL)
```
Host: 10.99.1.213
Port: 5432
Database: monitoring
User: logdetect
Password: (from config.yml)
Pool: Max Idle=10, Max Open=100
Timezone: Asia/Taipei
```

### Elasticsearch
```
URL: https://10.99.1.213:9200
Username: elastic
Password: (from config.yml)
SSL: InsecureSkipVerify=true (dev only!)
Indices: logstash-*, custom patterns
```

---

## JWT Token Details

- **Algorithm**: HS256 (HMAC-SHA256)
- **Expiration**: 24 hours
- **Secret**: From JWT_SECRET environment variable
- **Claims**: user_id, username, role_id, exp, iat
- **Format**: Bearer {token} in Authorization header

---

## Permission Model

```
Resource: device, target, receiver, indices, elasticsearch, user
Action: create, read, update, delete

Example:
- Permission: "device.create" (create devices)
- Permission: "target.update" (update targets)
- Permission: "elasticsearch.read" (view ES monitors)

Default Roles:
- admin: All permissions
- user: Read-only access
- operator: CRUD on assigned resources
```

---

## Startup Sequence

1. Load environment configuration (config.yml)
2. Initialize MySQL connection (GORM)
3. Initialize TimescaleDB connection (SQL)
4. Create database schema (AutoMigrate)
5. Initialize BatchWriter (for time-series data)
6. Initialize Elasticsearch client
7. Create default RBAC roles & permissions
8. Load cron scheduler
9. Initialize ES monitor scheduler
10. Bootstrap all monitoring targets
11. Start HTTP server on port 8006
12. Ready to accept requests

---

## Debugging Tips

### Enable Debug Logging
```go
// In main.go
log.Logrecord_no_rotate("DEBUG", "message here")
```

### Check Running Cron Jobs
```sql
-- View registered cron jobs
SELECT * FROM cron_lists;

-- View monitoring history
SELECT * FROM history ORDER BY created_at DESC LIMIT 10;

-- View ES monitoring data
SELECT * FROM elasticsearch_monitors;
```

### Common Issues

**Issue**: Devices not detected
- Check Elasticsearch connection
- Verify index pattern matches
- Check time range calculation in detect.go
- Review CloudWatch/logs for ES query errors

**Issue**: Emails not sending
- Verify SMTP credentials in config.yml
- Check email service logs
- Verify receiver email addresses

**Issue**: Cron jobs not running
- Check global.Crontab is initialized
- Verify targets are enabled
- Check cron_lists table for registered jobs

**Issue**: Performance degradation
- Monitor batch_writer flush times
- Check TimescaleDB insert performance
- Review ES query performance
- Check database connection pool utilization

---

## File Statistics

| Category | Count | Examples |
|----------|-------|----------|
| Controllers | 9 | auth.go, device.go, target.go, ... |
| Services | 22 | detect.go, center.go, batch_writer.go, ... |
| Entities | 20+ | User, Role, Device, Target, Index, History, ... |
| API Routes | 60+ | All CRUD operations + dashboard + ES monitoring |
| Database Tables | 16+ | users, targets, history, elasticsearch_monitors, ... |
| Configuration Sections | 8 | database, timescale, es, email, server, cors, sso |

---

## Important Directories

| Path | Purpose |
|------|---------|
| `/controller` | HTTP request handlers |
| `/services` | Business logic & operations |
| `/entities` | Data model definitions |
| `/clients` | Database & external service connections |
| `/middleware` | Authentication & authorization |
| `/router` | Route definitions & configuration |
| `/models` | Response & data structures |
| `/global` | Global application state |
| `/structs` | Configuration structures |
| `/log_record` | Log file storage location |
| `/specs` | Documentation (this file included) |
| `/docs` | Swagger API documentation |

---

## Technology Versions

```
Go: 1.24.0
Gin: v1.9.1
GORM: v1.25.7
Elasticsearch: v8
PostgreSQL/TimescaleDB: 12+
MySQL: 5.7+
JWT: github.com/golang-jwt/jwt/v5
Cron: github.com/robfig/cron/v3
```

---

## External Dependencies

- **Gin-gonic**: Web framework with middleware
- **GORM**: ORM for MySQL
- **lib/pq**: PostgreSQL driver for TimescaleDB
- **golang-jwt**: JWT token handling
- **bcrypt**: Password hashing
- **robfig/cron**: Job scheduling
- **elastic/go-elasticsearch**: ES client
- **swaggo**: Swagger documentation
- **natefinch/lumberjack**: Log file rotation

---

This quick reference covers the essential information needed to understand and work with the Log Detect Backend system.
