# Log Detect Backend - Comprehensive Codebase Analysis

## Executive Summary

**Log Detect Backend** is a Go-based monitoring and logging system built with Gin framework that integrates Elasticsearch, MySQL, and TimescaleDB for centralized log aggregation, device monitoring, and anomaly detection. The system provides a REST API for managing monitoring targets, devices, and indices with a comprehensive scheduler for periodic health checks and alert generation.

---

## 1. OVERALL ARCHITECTURE & PROJECT STRUCTURE

### Directory Structure
```
log-detect-backend/
├── main.go                 # Application entry point
├── config.yml             # Configuration file
├── setting.yml            # Monitoring targets and receivers configuration
├── go.mod / go.sum        # Go dependencies
├── clients/               # External service clients
│   ├── mysql.go           # MySQL database connection
│   ├── es.go              # Elasticsearch client
│   └── timescale.go       # TimescaleDB connection
├── entities/              # Domain models (User, Role, Target, Index, etc.)
├── models/                # Response and data structures
├── services/              # Business logic layer (22 services)
├── controller/            # HTTP request handlers (9 controllers)
├── router/                # Route definitions
├── middleware/            # Authentication and authorization
├── global/                # Global variables and configuration state
├── structs/               # Configuration structures
├── utils/                 # Utility functions (CORS configuration)
├── handler/               # Error handling
├── log/                   # Logging system
└── docs/                  # Swagger API documentation
```

### Technology Stack
- **Language**: Go 1.24.0
- **Web Framework**: Gin-gonic (v1.9.1)
- **Primary Database**: MySQL (GORM ORM)
- **Time Series DB**: TimescaleDB (PostgreSQL extension)
- **Search Engine**: Elasticsearch v8
- **Job Scheduler**: Robfig Cron v3
- **Authentication**: JWT (golang-jwt)
- **Password Hashing**: bcrypt
- **Documentation**: Swagger (swaggo)
- **CORS**: Gin-contrib/cors

---

## 2. MAIN COMPONENTS

### 2.1 Clients Layer (Connection Management)

#### mysql.go
- Establishes MySQL connection using GORM
- Implements connection pooling and lifecycle management
- Configuration: host, port, user, password, database name
- Supports automatic reconnection retry logic

#### es.go
- Elasticsearch v8 client initialization
- SSL/TLS configuration with InsecureSkipVerify option
- Supports username/password authentication
- Connection validation via cluster info endpoint

#### timescale.go
- PostgreSQL driver initialization for TimescaleDB
- Connection pooling configuration
- Time zone setting: Asia/Taipei
- Supports batch inserts for time-series data

---

### 2.2 Entities (Domain Models)

#### User & Authentication Entities
```go
type User struct {
  ID       uint
  Username string (unique index)
  Email    string (unique index)
  Password string (bcrypt hashed)
  Role     Role (foreign key)
  RoleID   uint
  IsActive bool
}

type Role struct {
  ID          uint
  Name        string (unique index)
  Description string
  Permissions []Permission (many-to-many)
}

type Permission struct {
  ID          uint
  Name        string (unique index)
  Resource    string  // e.g., "device", "target", "elasticsearch"
  Action      string  // "create", "read", "update", "delete"
  Description string
}
```

#### Monitoring Entities
```go
type Target struct {
  ID      int
  Subject string
  To      []string (email receivers)
  Enable  bool
  Indices []Index (many-to-many via indices_targets)
}

type Index struct {
  ID          int
  Pattern     string      // ES index pattern
  DeviceGroup string
  Logname     string
  Period      string      // "minutes", "hours"
  Unit        int
  Field       string      // field for aggregation
}

type Device struct {
  ID          int
  DeviceGroup string
  Name        string
}

type Receiver struct {
  ID   int
  Name []string (JSON serialized)
}
```

#### History & Metrics Entities
```go
type History struct {
  Logname     string
  DeviceGroup string
  Name        string        // device name
  Status      string        // "online", "offline", "warning", "error"
  Lost        string        // "true"/"false"
  LostNum     int
  Date        string        // YYYY-MM-DD
  Time        string        // HH:MM:SS
  DateTime    string        // YYYY-MM-DD HH:MM:SS
  Timestamp   int64         // Unix timestamp
  Period      string
  Unit        int
  ResponseTime int64        // milliseconds
  DataCount   int64
  ErrorMsg    string
  ErrorCode   string
  Metadata    string        // JSON
}

type HistoryArchive struct {
  // Same fields as History (for archival)
}

type HistoryDailyStats struct {
  Date            string
  Logname         string
  DeviceGroup     string
  TotalChecks     int64
  OnlineCount     int64
  OfflineCount    int64
  WarningCount    int64
  ErrorCount      int64
  UptimeRate      float64
  AvgResponseTime float64
}

type AlertHistory struct {
  Logname     string
  DeviceGroup string
  DeviceName  string
  AlertType   string        // "offline", "error", "warning"
  Severity    string        // "low", "medium", "high", "critical"
  Message     string
  Status      string        // "active", "resolved", "acknowledged"
  ResolvedAt  *int64
  ResolvedBy  string
}
```

#### Elasticsearch Monitoring Entities
```go
type ElasticsearchMonitor struct {
  ID                  int
  Name                string
  Host                string
  Port                int           // default: 9200
  Username            string
  Password            string
  EnableAuth          bool
  CheckType           string        // "health,performance"
  Interval            int           // seconds
  EnableMonitor       bool
  Receivers           []string
  Subject             string
  Description         string
  
  // Alert Thresholds
  CPUUsageHigh        *float64
  CPUUsageCritical    *float64
  MemoryUsageHigh     *float64
  MemoryUsageCritical *float64
  DiskUsageHigh       *float64
  DiskUsageCritical   *float64
  ResponseTimeHigh    *int64
  ResponseTimeCritical *int64
  UnassignedShardsThreshold *int
  AlertThreshold      string        // JSON backup
  AlertDedupeWindow   int           // seconds (default: 300)
}

type ESMetric struct {
  Time               time.Time
  MonitorID          int
  Status             string        // "online", "offline", "warning", "error"
  ClusterName        string
  ClusterStatus      string        // "green", "yellow", "red"
  ResponseTime       int64         // milliseconds
  CPUUsage           float64       // percentage
  MemoryUsage        float64       // percentage
  DiskUsage          float64       // percentage
  NodeCount          int
  DataNodeCount      int
  QueryLatency       int64
  IndexingRate       float64
  SearchRate         float64
  TotalIndices       int
  TotalDocuments     int64
  TotalSizeBytes     int64
  ActiveShards       int
  RelocatingShards   int
  UnassignedShards   int
  ErrorMessage       string
  WarningMessage     string
  Metadata           string        // JSON
}

type ESAlert struct {
  Time            time.Time
  MonitorID       int
  AlertType       string        // "health", "performance", "capacity"
  Severity        string        // "critical", "high", "medium", "low"
  Message         string
  Status          string        // "active", "resolved", "acknowledged"
  ClusterName     string
  ThresholdValue  *float64
  ActualValue     *float64
  ResolvedAt      *time.Time
  ResolvedBy      string
  ResolutionNote  string
  AcknowledgedAt  *time.Time
  AcknowledgedBy  string
  Metadata        string
}
```

#### Dashboard & Statistics Entities
```go
type HistoryStatistics struct {
  Date            string
  Logname         string
  DeviceGroup     string
  TotalChecks     int64
  OnlineCount     int64
  OfflineCount    int64
  WarningCount    int64
  ErrorCount      int64
  UptimeRate      float64
  AvgResponseTime int64
}

type DashboardData struct {
  TotalTargets   int64
  ActiveTargets  int64
  TotalDevices   int64
  OnlineDevices  int64
  OfflineDevices int64
  UptimeRate     float64
  ActiveAlerts   int64
  LastUpdateTime string
}
```

---

### 2.3 Services Layer (Business Logic - 22 Services)

#### Authentication Service (auth.go)
- **HashPassword()**: bcrypt hashing with default cost
- **CheckPassword()**: Password verification
- **GenerateJWT()**: JWT token generation (24-hour expiration)
- **ValidateJWT()**: Token validation and claims extraction
- **Login()**: User authentication
- **Register()**: User registration with permission checking
- **CreateDefaultRolesAndPermissions()**: Bootstrap default RBAC structure
- **CreateDefaultAdmin()**: Create default admin user
- **CheckPermission()**: Permission verification for resources
- **GetUserByID()**: User retrieval

#### Detection Service (detect.go)
- **Detect()**: Core monitoring function
  - Queries Elasticsearch for devices in given time range
  - Compares ES results with database device list
  - Detects new/removed/offline devices
  - Generates history records
  - Sends email alerts
  - Handles device auto-creation

#### Center Scheduler Service (center.go)
- **LoadCrontab()**: Initializes cron scheduler
- **ExecuteCrontab()**: Registers cron jobs for targets
- **Control_center()**: Bootstrap all active targets on startup
- **Control_center_by_TargetID()**: Register single target
- **Control_center_by_IndiceID()**: Register single index

#### History Service (history.go)
- **CreateHistory()**: Insert monitoring results
- **CreateMailHistory()**: Log email sending events
- **GetHistoryDataByDeviceName()**: Query device history
- **GetHistoryDataByDeviceName_TS()**: TimescaleDB variant
- **GenerateTimeArray()**: Generate time slots for analysis

#### Device Service (device.go)
- **CreateDevice()**: Bulk device creation
- **UpdateDevice()**: Device update
- **DeleteDevice()**: Device deletion
- **GetDevicesDataByGroupName()**: Get devices by group
- **GetDeviceGroup()**: Get all device groups with counts

#### Target Service (target.go)
- **GetAllTargets()**: Fetch all targets with indices
- **CreateTarget()**: Create monitoring target
- **UpdateTarget()**: Update target configuration
- **DeleteTarget()**: Delete target and associated cron jobs
- **GetTargetByID()**: Retrieve single target

#### Indices Service (indices.go)
- **CreateIndices()**: Create ES index configuration
- **UpdateIndices()**: Update index mapping
- **DeleteIndices()**: Delete index configuration
- **GetIndicesDataByLogname()**: Find index by log name
- **GetAllIndices()**: List all indices

#### Receiver Service (receiver.go)
- **GetAllReceivers()**: Fetch all receivers
- **CreateReceiver()**: Create email receiver
- **UpdateReceiver()**: Update receiver
- **DeleteReceiver()**: Delete receiver

#### Mail Service (mail.go)
- **Mail4()**: Send HTML email with table formatting
- **Support for**: SMTP authentication, custom templates
- **Configuration**: Gmail SMTP, custom host/port

#### Elasticsearch Monitor Service (es_monitor_service.go)
- **CreateESMonitor()**: Create ES health monitor
- **UpdateESMonitor()**: Update monitor config
- **GetAllESMonitors()**: List all monitors
- **GetESMonitorByID()**: Get monitor details
- **DeleteESMonitor()**: Delete monitor

#### ES Scheduler Service (es_scheduler.go)
- **InitESScheduler()**: Initialize singleton scheduler
- **LoadAllMonitors()**: Load enabled monitors on startup
- **StartMonitor()**: Start periodic health check
- **RestartMonitor()**: Restart with new configuration
- **StopMonitor()**: Stop monitoring

#### Elasticsearch Query Service (es_query.go)
- **SearchRequest()**: Execute ES DSL queries
  - Aggregation on specified field
  - Date range filtering with @timestamp
  - Support for complex filtering

#### ES Insert Service (es_insert.go)
- Data insertion into Elasticsearch indices

#### Batch Writer Service (batch_writer.go)
- **NewBatchWriter()**: Initialize batch writer with configurable size/interval
- **AddHistory()**: Add records to batch queue
- **flushDeviceMetrics()**: Bulk insert to TimescaleDB device_metrics
- **flushESMetrics()**: Bulk insert to TimescaleDB es_metrics
- **flushBatch()**: Periodic flush triggered by ticker
- **Stop()**: Graceful shutdown with final flush

**Features**:
- Supports multiple history types (Device/ES metrics)
- Configurable batch size (default: 50)
- Configurable flush interval (default: 5s)
- Transaction support for data consistency
- Prepared statements for performance

#### Setting Service (setting.go)
- Load YAML configuration from setting.yml
- Parse targets and receivers from YAML

#### SQL Table Service (sqltable.go)
- **CreateTable()**: Database schema migration using GORM AutoMigrate
- Creates tables for: Users, Roles, Permissions, Devices, Receivers, Indices, Targets, History, Archives, Stats, MailHistory, AlertHistory, CronList, ESMonitor

#### Tools Service (tools.go)
- **ListCompare()**: Compare device lists (added/removed/intersection)

#### Dashboard Service
- **GetDashboardData()**: Aggregate system statistics
- **GetHistoryStatistics()**: Historical data aggregation
- **GetTrendData()**: Trend analysis over time
- **GetGroupStatistics()**: Device group statistics
- **GetDeviceStatusOverview()**: Device status summary
- **GetDeviceTimeline()**: Device status timeline visualization
- **GetRecentAlerts()**: Recent alert list
- **CreateAlert()**: Create alert record
- **UpdateAlertStatus()**: Update alert status

---

### 2.4 Controller Layer (HTTP Handlers)

#### Authentication Controller (auth.go)
- `POST /auth/login` - User login
- `POST /api/v1/auth/register` - Register new user
- `GET /api/v1/auth/profile` - Get user profile
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `GET /api/v1/auth/users` - List users
- `GET /api/v1/auth/users/:id` - Get user by ID
- `PUT /api/v1/auth/users/:id` - Update user
- `DELETE /api/v1/auth/users/:id` - Delete user

#### Target Controller (target.go)
- `GET /api/v1/Target/GetAll` - List all targets
- `POST /api/v1/Target/Create` - Create target
- `PUT /api/v1/Target/Update` - Update target
- `DELETE /api/v1/Target/Delete/:id` - Delete target

#### Device Controller (device.go)
- `GET /api/v1/Device/GetAll` - List devices
- `POST /api/v1/Device/Create` - Create device
- `PUT /api/v1/Device/Update` - Update device
- `DELETE /api/v1/Device/Delete/:id` - Delete device
- `GET /api/v1/Device/count` - Get device counts by group
- `GET /api/v1/Device/GetGroup` - Get device groups

#### Indices Controller (indices.go)
- `GET /api/v1/Indices/GetAll` - List indices
- `POST /api/v1/Indices/Create` - Create index
- `PUT /api/v1/Indices/Update` - Update index
- `DELETE /api/v1/Indices/Delete/:id` - Delete index
- `GET /api/v1/Indices/GetLogname` - Get log names
- `GET /api/v1/Indices/GetIndicesByLogname/:logname` - Get by logname
- `GET /api/v1/Indices/GetIndicesByTargetID/:id` - Get by target

#### Receiver Controller (receiver.go)
- `GET /api/v1/Receiver/GetAll` - List receivers
- `POST /api/v1/Receiver/Create` - Create receiver
- `PUT /api/v1/Receiver/Update` - Update receiver
- `DELETE /api/v1/Receiver/Delete/:id` - Delete receiver

#### History Controller (history.go)
- `GET /api/v1/History/GetData/:logname` - Get history by logname
- `GET /api/v1/History/GetLognameData` - Get all logname data

#### Dashboard Controller (dashboard.go)
- `GET /api/v1/dashboard/overview` - System overview
- `GET /api/v1/dashboard/statistics` - History statistics
- `GET /api/v1/dashboard/trends` - Trend data
- `GET /api/v1/dashboard/groups/statistics` - Group statistics
- `GET /api/v1/dashboard/devices/status` - Device status overview
- `GET /api/v1/dashboard/devices/:device_name/timeline` - Device timeline
- `GET /api/v1/dashboard/alerts/recent` - Recent alerts
- `POST /api/v1/dashboard/alerts` - Create alert
- `PUT /api/v1/dashboard/alerts/:id/status` - Update alert status

#### Elasticsearch Controller (elasticsearch.go)
- `GET /api/v1/elasticsearch/monitors` - List monitors
- `GET /api/v1/elasticsearch/monitors/:id` - Get monitor by ID
- `POST /api/v1/elasticsearch/monitors` - Create monitor
- `PUT /api/v1/elasticsearch/monitors` - Update monitor
- `DELETE /api/v1/elasticsearch/monitors/:id` - Delete monitor
- `POST /api/v1/elasticsearch/monitors/:id/test` - Test connection
- `POST /api/v1/elasticsearch/monitors/:id/toggle` - Enable/disable monitor
- `GET /api/v1/elasticsearch/status` - Get all monitors status
- `GET /api/v1/elasticsearch/status/:id/history` - Monitor history
- `GET /api/v1/elasticsearch/statistics` - ES statistics
- `GET /api/v1/elasticsearch/alerts` - List alerts
- `GET /api/v1/elasticsearch/alerts/:monitor_id` - Get alert by ID
- `POST /api/v1/elasticsearch/alerts/:monitor_id/resolve` - Resolve alert
- `PUT /api/v1/elasticsearch/alerts/:monitor_id/acknowledge` - Acknowledge alert

#### Settings Controller (setting.go)
- `GET /api/v1/get-sso-url` - Get SSO configuration
- `GET /api/v1/user/get-server-menu` - Get menu structure
- `GET /api/v1/get-server-module` - Get available modules

---

### 2.5 Router Configuration

**Route Groups & Middleware Chain**:

1. **Public Routes** (No Auth)
   - `GET /healthcheck` - Health check endpoint
   - `POST /auth/login` - Public login

2. **Protected API Routes** (/api/v1)
   - Applied middlewares:
     - `OptionalAuthMiddleware()` - Backward compatibility
     - `AuthMiddleware()` - On protected resources
     - `PermissionMiddleware(resource, action)` - RBAC enforcement

3. **Resource-based Route Groups**:
   - `/Receiver` - Email receivers (permissions: target.*)
   - `/Target` - Monitoring targets (permissions: target.*)
   - `/Device` - Managed devices (permissions: device.*)
   - `/Indices` - ES indices (permissions: indices.*)
   - `/History` - Monitoring history (public read)
   - `/dashboard` - Visualization endpoints
   - `/elasticsearch` - ES monitoring (permissions: elasticsearch.*)
   - `/admin/data` - Data management (admin only)

---

### 2.6 Middleware (Security & Authorization)

#### Authentication Middleware (middleware/auth.go)
- **AuthMiddleware()**: Validates JWT tokens
  - Requires "Bearer {token}" format
  - Sets user context (user_id, username, role_id)
  - Returns 401 on invalid/missing token

- **OptionalAuthMiddleware()**: Optional authentication
  - Validates token if present
  - Does not abort on missing token
  - Enables backward compatibility

- **PermissionMiddleware(resource, action)**: RBAC enforcement
  - Checks user has permission for resource action
  - Returns 403 if insufficient permissions

- **RequireRole(...roles)**: Role-based access control
  - Accepts multiple allowed roles
  - Used for admin/special routes

- **AdminOnly()**: Convenience for admin-only routes
  - Shorthand for RequireRole("admin")

- **GetCurrentUser()**: Retrieves user from context
  - Fetches full user object from database
  - Used in handlers for user-specific logic

---

### 2.7 Global Configuration & State

```go
type GlobalState {
  EnvConfig      *EnviromentModel  // Loaded from config.yml
  Elasticsearch  *elasticsearch.Client
  TargetStruct   *TargetStruct     // From setting.yml
  Mysql          *gorm.DB          // MySQL connection
  Crontab        *cron.Cron        // Scheduler instance
  TimescaleDB    *sql.DB           // TimescaleDB connection
  BatchWriter    BatchWriterType   // Batch insert service
}
```

---

## 3. KEY FUNCTIONALITIES & FEATURES

### 3.1 Core Monitoring Workflow

1. **Configuration Loading** (main.go)
   - Load environment config (config.yml)
   - Load monitoring targets (setting.yml)
   - Initialize database connections
   - Create database schema

2. **Scheduler Initialization** (center.go)
   - Cron job registration for each target/index pair
   - Cron expression: Minutes: `*/{unit} * * * *`, Hours: `0 */{unit} * * *`
   - Entry IDs stored in CronList table for management

3. **Periodic Detection** (detect.go)
   - Execute on schedule
   - Query Elasticsearch for devices in time range
   - Compare with device database
   - Detect new/offline devices
   - Store results in History table
   - Generate alerts for missing devices

4. **Alert Generation & Delivery**
   - Email notifications via SMTP
   - HTML-formatted tables with offline device lists
   - Receiver list from Target configuration

### 3.2 Elasticsearch Monitoring

**Independent Monitoring System**:

1. **Monitor Configuration** (ElasticsearchMonitor entity)
   - Separate scheduler (ESMonitorScheduler)
   - Configurable check intervals (10-3600 seconds)
   - Multiple check types: health, performance, capacity, availability

2. **Health Checks** (ESMonitorService)
   - Cluster health status (green/yellow/red)
   - CPU/Memory/Disk usage
   - Node count and shard distribution
   - Query/Index latency
   - Response time tracking

3. **Alert Thresholds**
   - CPU: High (75%), Critical (85%)
   - Memory: High (80%), Critical (90%)
   - Disk: High (85%), Critical (95%)
   - Response Time: High (3000ms), Critical (10000ms)
   - Unassigned Shards: 1+

4. **Data Storage**
   - Metrics: TimescaleDB (es_metrics table)
   - Alerts: TimescaleDB (alert_history table)
   - Configuration: MySQL (elasticsearch_monitors table)

5. **Alert Management**
   - Status: active, resolved, acknowledged
   - Deduplication window (default: 5 minutes)
   - Resolution tracking

### 3.3 User Authentication & Authorization

**Multi-level RBAC**:

```
User
  └─ Role (1-to-1)
      └─ Permissions (many-to-many)
```

**Permission Structure**:
```
Resource: device, target, receiver, indices, elasticsearch, user
Action: create, read, update, delete
Name: {resource}.{action}
```

**Default Roles**:
- admin: Full permissions
- user: Read access to all resources
- operator: Create/Update permissions on assigned resources

### 3.4 Data Pipeline

**MySQL → History → TimescaleDB Batch Pipeline**:

1. **Write Path**:
   - Detect() generates History records
   - BatchWriter accumulates records in memory
   - Flush on: batch full (50 items) OR timer (5s)
   - TimescaleDB insert via prepared statements
   - Transaction-wrapped for consistency

2. **TimescaleDB Tables**:
   - `device_metrics`: Device monitoring data (hypertable)
   - `es_metrics`: Elasticsearch metrics (hypertable)
   - `alert_history`: Alert records

3. **Query Path**:
   - History queries from MySQL for real-time data
   - TimescaleDB for historical/aggregated data
   - Dashboard services aggregate across both sources

### 3.5 Data Archival & Aggregation

**History Management** (dashboard.go):
- **CleanOldHistory()**: Delete history older than threshold
- **ArchiveOldHistory()**: Move to history_archive table
- **CreateDailyAggregates()**: Generate HistoryDailyStats
- **GetStorageStats()**: Storage usage monitoring

---

## 4. DATABASE SCHEMA

### MySQL Tables

| Table | Purpose | Key Fields |
|-------|---------|-----------|
| users | User accounts | id, username (unique), email (unique), password, role_id, is_active |
| roles | User roles | id, name (unique), description |
| permissions | System permissions | id, name (unique), resource, action |
| role_permissions | RBAC mapping | role_id, permission_id |
| devices | Monitored devices | id, device_group, name |
| targets | Monitoring targets | id, subject, to (JSON), enable |
| indices | ES index configs | id, pattern, device_group, logname, period, unit, field |
| indices_targets | Target-Index mapping | target_id, index_id |
| receivers | Email receivers | id, name (JSON) |
| history | Monitoring results | id, logname, device_group, name, status, lost, date, time, response_time, data_count, metadata (JSON) |
| history_archive | Archived history | same as history |
| history_daily_stats | Daily aggregates | date, logname, device_group, total_checks, online_count, uptime_rate, avg_response_time |
| mail_history | Email logs | id, date, time, logname, sended |
| alert_history | Alert records | id, logname, device_group, device_name, alert_type, severity, status, resolved_at, resolved_by |
| cron_lists | Scheduler mapping | entry_id, target_id, index_id |
| elasticsearch_monitors | ES monitor config | id, name, host, port, check_type, interval, enable_monitor, threshold fields, alert_threshold (JSON) |

### TimescaleDB Tables

| Table | Purpose | Key Fields |
|-------|---------|-----------|
| device_metrics | Device monitoring (hypertable) | time (partitioned), device_id, device_group, logname, status, lost, response_time, data_count, error_msg |
| es_metrics | ES metrics (hypertable) | time (partitioned), monitor_id, cpu_usage, memory_usage, disk_usage, node_count, active_shards, unassigned_shards |
| alert_history | Alert records | time, monitor_id, alert_type, severity, status |

---

## 5. API ENDPOINTS & ROUTING

### Authentication Endpoints
```
POST    /auth/login                         # Login
POST    /api/v1/auth/register              # Register user
GET     /api/v1/auth/profile               # Get profile
POST    /api/v1/auth/refresh               # Refresh token
GET     /api/v1/auth/users                 # List users
GET     /api/v1/auth/users/:id             # Get user
PUT     /api/v1/auth/users/:id             # Update user
DELETE  /api/v1/auth/users/:id             # Delete user
```

### Target Management
```
GET     /api/v1/Target/GetAll              # List targets
POST    /api/v1/Target/Create              # Create target
PUT     /api/v1/Target/Update              # Update target
DELETE  /api/v1/Target/Delete/:id          # Delete target
```

### Device Management
```
GET     /api/v1/Device/GetAll              # List devices
POST    /api/v1/Device/Create              # Create device
PUT     /api/v1/Device/Update              # Update device
DELETE  /api/v1/Device/Delete/:id          # Delete device
GET     /api/v1/Device/count               # Count by group
GET     /api/v1/Device/GetGroup            # List groups
```

### Index Management
```
GET     /api/v1/Indices/GetAll             # List indices
POST    /api/v1/Indices/Create             # Create index
PUT     /api/v1/Indices/Update             # Update index
DELETE  /api/v1/Indices/Delete/:id         # Delete index
GET     /api/v1/Indices/GetLogname         # Get log names
GET     /api/v1/Indices/GetIndicesByLogname/:logname  # By logname
GET     /api/v1/Indices/GetIndicesByTargetID/:id      # By target
```

### Receiver Management
```
GET     /api/v1/Receiver/GetAll            # List receivers
POST    /api/v1/Receiver/Create            # Create receiver
PUT     /api/v1/Receiver/Update            # Update receiver
DELETE  /api/v1/Receiver/Delete/:id        # Delete receiver
```

### History & Monitoring
```
GET     /api/v1/History/GetData/:logname   # Get by logname
GET     /api/v1/History/GetLognameData     # Get all data
```

### Dashboard & Visualization
```
GET     /api/v1/dashboard/overview         # System overview
GET     /api/v1/dashboard/statistics       # Statistics
GET     /api/v1/dashboard/trends           # Trends
GET     /api/v1/dashboard/groups/statistics # Group stats
GET     /api/v1/dashboard/devices/status   # Device status
GET     /api/v1/dashboard/devices/:device_name/timeline # Timeline
GET     /api/v1/dashboard/alerts/recent    # Recent alerts
POST    /api/v1/dashboard/alerts           # Create alert
PUT     /api/v1/dashboard/alerts/:id/status # Update alert
```

### Elasticsearch Monitoring
```
GET     /api/v1/elasticsearch/monitors     # List monitors
GET     /api/v1/elasticsearch/monitors/:id # Get monitor
POST    /api/v1/elasticsearch/monitors     # Create monitor
PUT     /api/v1/elasticsearch/monitors     # Update monitor
DELETE  /api/v1/elasticsearch/monitors/:id # Delete monitor
POST    /api/v1/elasticsearch/monitors/:id/test      # Test connection
POST    /api/v1/elasticsearch/monitors/:id/toggle    # Enable/disable
GET     /api/v1/elasticsearch/status       # All status
GET     /api/v1/elasticsearch/status/:id/history     # Monitor history
GET     /api/v1/elasticsearch/statistics   # Statistics
GET     /api/v1/elasticsearch/alerts       # List alerts
GET     /api/v1/elasticsearch/alerts/:monitor_id     # Get alert
POST    /api/v1/elasticsearch/alerts/:monitor_id/resolve      # Resolve
PUT     /api/v1/elasticsearch/alerts/:monitor_id/acknowledge  # Acknowledge
```

### Admin Data Management
```
DELETE  /api/v1/admin/data/clean-history          # Clean old history
POST    /api/v1/admin/data/archive-history        # Archive history
POST    /api/v1/admin/data/create-aggregates      # Create aggregates
GET     /api/v1/admin/data/storage-stats          # Storage stats
```

### Configuration
```
GET     /api/v1/get-sso-url                # SSO configuration
GET     /api/v1/user/get-server-menu       # Menu structure
GET     /api/v1/get-server-module          # Modules
GET     /healthcheck                       # Health check
```

---

## 6. AUTHENTICATION & AUTHORIZATION

### JWT Authentication Flow

1. **Login** (POST /auth/login)
   ```
   Request: { username, password }
   Response: { token, user }
   Token: HS256 signed, 24-hour expiration
   Claims: user_id, username, role_id, exp, iat
   ```

2. **Token Validation** (AuthMiddleware)
   ```
   Header: Authorization: Bearer {token}
   Validates: Signature, Expiration, Format
   Sets Context: user_id, username, role_id
   ```

3. **Permission Checking** (PermissionMiddleware)
   ```
   Validates: User has resource.action permission
   Returns: 403 Forbidden if insufficient
   ```

### RBAC Implementation

```
User → Role → Permissions (many-to-many)
       ↓
   Loaded on login
   
Middleware checks:
1. User authenticated (valid JWT)
2. User active
3. User has required role/permission
```

### JWT Configuration

- **Secret Key**: From environment variable JWT_SECRET
- **Default**: "your-secret-key-change-in-production"
- **Expiration**: 24 hours
- **Algorithm**: HS256 (HMAC-SHA256)
- **Hash Function**: bcrypt for passwords

---

## 7. EXTERNAL INTEGRATIONS

### 7.1 Elasticsearch Integration

**Query Capabilities**:
- DSL query construction
- Aggregation on configurable fields
- Date range filtering with @timestamp
- Support for complex boolean queries

**Monitoring**:
- Cluster health status
- Node and shard information
- Index statistics
- Performance metrics

**Query Example**:
```json
{
  "aggs": {
    "2": {
      "terms": {
        "field": "host.keyword",
        "size": 100
      }
    }
  },
  "query": {
    "bool": {
      "filter": [
        {
          "range": {
            "@timestamp": {
              "gte": "2024-01-01T00:00",
              "lte": "2024-01-01T23:59"
            }
          }
        }
      ]
    }
  }
}
```

### 7.2 Email Integration

**SMTP Configuration**:
- Provider: Gmail (configurable)
- Host: smtp.gmail.com (configurable)
- Port: 587 (configurable)
- Auth: Username/Password
- TLS: Supported

**Email Features**:
- HTML templates
- Table formatting for results
- Multiple recipients
- Subject customization

**From config.yml**:
```yaml
email:
  user: "rabot6201@gmail.com"
  password: "fuhbrbezkwpcuzsv"
  sender: "rabot6201@gmail.com"
  host: "smtp.gmail.com"
  port: "587"
  auth: true
```

### 7.3 SSO Integration (Keycloak)

**Configuration** (config.yml):
```yaml
sso:
  url: "https://10.99.1.133:8443"
  realm: "master"
  username: "admin"
  password: "1qaz2wsx"
  client_id: "log-detect"
  admin_role: "log-detect-admin"
  user_role: "log-detect-op"
```

**Purpose**: External identity provider integration (not fully implemented in code)

---

## 8. BUSINESS LOGIC & CORE FEATURES

### 8.1 Device Discovery & Management

**Auto-Discovery** (detect.go):
1. Query ES for devices in configured pattern/field
2. If no devices in DB for group: Create all ES devices
3. If devices exist: Compare ES results with DB
4. **Added** devices: Create and insert into DB
5. **Removed** devices: Generate alert
6. **Duplicates**: Consolidate and cleanup

**Device Groups**:
- Logical grouping of monitored devices
- Used in monitoring queries
- Tracked in Device entity

### 8.2 Target & Index Management

**Target**:
- Monitoring configuration container
- Contains email receivers
- Can be enabled/disabled
- Multiple indices per target

**Index**:
- ES index pattern definition
- Aggregation field specification
- Monitoring schedule (period/unit)
- Links to targets via many-to-many

**Workflow**:
1. Create Indices (ES patterns)
2. Create Targets (receiver groups)
3. Link Indices → Targets
4. System generates cron jobs
5. Detect() executes on schedule

### 8.3 Cron Job Management

**Registration** (center.go):
```
For each Target with Indices:
  For each Index:
    Generate cron expression
    Register with robfig/cron/v3
    Store EntryID in CronList
```

**Cron Expressions**:
- Minutes: `*/{unit} * * * *`
- Hours: `0 */{unit} * * *`

**Lifecycle**:
- Create: On target creation
- Update: Remove old → Create new
- Delete: Remove from cron + CronList
- Restart: Full rebuild on system start

### 8.4 Alert Generation

**Trigger Events**:
1. Device offline (missing from ES results)
2. ES monitor threshold exceeded
3. Elasticsearch cluster issue

**Alert Properties**:
- Type: offline, error, warning, health, performance
- Severity: low, medium, high, critical
- Status: active, resolved, acknowledged
- Timestamp: When detected
- Device/Monitor info: Which resource affected

**Delivery**:
- Email to configured receivers
- HTML formatted
- Device list in table format

### 8.5 Time-Series Data Management

**Batch Writing** (batch_writer.go):
```
Add Record → Memory Buffer → Flush Decision
                                    ↓
                          Size (≥50) or Time (5s)
                                    ↓
                          Batch Insert to TimescaleDB
                                    ↓
                          Transaction + Prepared Stmt
```

**Advantages**:
- Reduced write overhead
- Better network utilization
- Transaction support
- Prepared statement protection

**Configuration** (config.yml):
```yaml
batch_writer:
  enabled: true
  batch_size: 50
  flush_interval: "5s"
```

### 8.6 Dashboard & Analytics

**Overview**:
- Total targets/devices
- Online/offline counts
- System-wide uptime rate
- Active alert count

**Statistics**:
- Per-logname breakdown
- Per-device-group breakdown
- Time-series data points

**Trends**:
- Uptime percentage over time
- Offline device count trends
- Error count trends
- Response time trends

**Device Timeline**:
- Status at each check interval
- Response time per check
- Data count per check
- Error details

---

## 9. CONFIGURATION FILES

### config.yml (Main Configuration)

```yaml
database:
  client: "mysql"
  host: "10.99.1.133"
  port: "3306"
  user: "runner"
  password: "1qaz2wsx"
  name: "logdetect"
  max_idle: 10
  max_open_conn: 100
  max_life_time: "1h"

timescale:
  host: "10.99.1.213"
  port: "5432"
  user: "logdetect"
  password: "your_secure_password"
  name: "monitoring"
  max_idle: 10
  max_open_conn: 100

batch_writer:
  enabled: true
  batch_size: 50
  flush_interval: "5s"

server:
  port: ":8006"
  mode: "debug"

es:
  url: "https://10.99.1.213:9200"
  sourceAccount: "elastic"
  sourcePassword: "a12345678"

email:
  user: "rabot6201@gmail.com"
  password: "fuhbrbezkwpcuzsv"
  sender: "rabot6201@gmail.com"
  host: "smtp.gmail.com"
  port: "587"
  auth: true

sso:
  url: "https://10.99.1.133:8443"
  realm: "master"
  client_id: "log-detect"
```

### setting.yml (Monitoring Configuration)

```yaml
targets:
  - receiver:
      - "russell.chen@bimap.co"
      - "rabot6201@gmail.com"
    subject: "log-detect test01"
    indices:
      - index: "logstash-log_detect*"
        logname: "waf"
        period: "minutes"
        unit: 60
        field: "host.keyword"
```

---

## 10. UTILITY & HELPER FUNCTIONS

### CORS Configuration (utils/cors.go)
```go
AllowOrigins: []string{"http://localhost:4200"}
AllowMethods: ["PUT", "POST", "GET", "DELETE"]
AllowHeaders: ["Content-Type", "Authorization", "realm"]
AllowCredentials: true
```

### Error Logging (handler/error.go)
- Writes to monthly-rotated log files
- Format: `apiError-{YYYYMM}.log`
- Logs: Method, URL, Error Message

### Tools & Helpers (services/tools.go)
- **ListCompare()**: Detects added/removed items between lists

---

## 11. INITIALIZATION & STARTUP FLOW

**main.go Startup Sequence**:

```
1. LoadEnvironment()           # Parse config.yml
   ↓
2. LoadDatabase()              # MySQL connection
   ↓
3. LoadTimescaleDB()           # PostgreSQL connection
   ↓
4. Initialize BatchWriter      # Batch insert service
   (if enabled in config)
   ↓
5. SetElkClient()              # Elasticsearch connection
   ↓
6. CreateTable()               # Database migration
   ↓
7. InitAuthService()           # RBAC initialization
   - CreateDefaultRolesAndPermissions()
   - CreateDefaultAdmin()
   ↓
8. LoadCrontab()               # Scheduler initialization
   ↓
9. InitESScheduler()           # ES monitor scheduler
   - LoadAllMonitors()         # Load from database
   ↓
10. Control_center()            # Bootstrap all targets
    - Register all cron jobs
    ↓
11. LoadRouter()                # HTTP routes
    ↓
12. Listen(:8006)               # Start server
```

---

## 12. KEY DESIGN PATTERNS

### Singleton Pattern
- GlobalESScheduler (ES monitoring)
- Global database connections

### Factory Pattern
- NewAuthService(), NewBatchWriter()

### Middleware Chain
- AuthMiddleware → PermissionMiddleware

### Repository Pattern (Partial)
- Service functions act as business logic
- Direct GORM queries in services

### Batch Processing Pattern
- BatchWriter accumulates and flushes periodically

### Scheduler Pattern
- Robfig Cron for detection scheduling
- Time-based ticker for batch flushing

---

## 13. DEPENDENCIES & IMPORTS

### External Libraries
```go
github.com/gin-gonic/gin              # Web framework
github.com/gin-contrib/cors           # CORS middleware
gorm.io/gorm                          # ORM
gorm.io/driver/mysql                  # MySQL driver
github.com/elastic/go-elasticsearch/v8 # ES client
github.com/lib/pq                     # PostgreSQL driver
github.com/golang-jwt/jwt/v5          # JWT
golang.org/x/crypto/bcrypt            # Password hashing
github.com/robfig/cron/v3             # Cron scheduler
github.com/spf13/viper                # Config management
github.com/swaggo/gin-swagger         # Swagger UI
github.com/natefinch/lumberjack       # Log rotation
```

---

## 14. SECURITY CONSIDERATIONS

### Strengths
- JWT-based authentication
- bcrypt password hashing
- RBAC with permission checking
- SSL/TLS for Elasticsearch
- Parameterized queries (GORM + prepared statements)

### Areas for Improvement
- JWT secret should never be hardcoded default
- SSL verification disabled for ES (InsecureSkipVerify)
- Email credentials in config file
- No rate limiting
- No input validation middleware
- No request signing

---

## 15. SCALABILITY & PERFORMANCE

### Current Optimizations
- Connection pooling (MySQL, TimescaleDB)
- Batch inserts to reduce I/O
- Prepared statements for ES queries
- Configurable batch size and flush interval
- Time-series database for metrics
- JSON metadata fields for extensibility

### Potential Bottlenecks
- Single Cron instance (scheduler)
- Synchronous email sending
- Memory-based batching (not persistent)
- No caching layer
- No query optimization hints

---

## 16. COMPLIANCE & GOVERNANCE

### Data Management
- History archival for old records
- Daily statistics aggregation
- Configurable retention policies
- Alert history tracking

### Audit Trail
- Mail history logging
- CronList for job tracking
- User audit via roles/permissions
- Alert resolution tracking

---

## CONCLUSION

Log Detect Backend is a comprehensive monitoring system with:
- **Multi-database architecture**: MySQL + TimescaleDB + Elasticsearch
- **Sophisticated scheduling**: Cron-based detection + ES monitor scheduler
- **Complete RBAC**: User/Role/Permission model
- **Rich analytics**: Dashboard, statistics, trends
- **Email alerting**: HTML-formatted notifications
- **Extensible design**: JSON metadata, pluggable checkers

The codebase follows layered architecture (controllers → services → entities → clients) with good separation of concerns. The batch writing system enables efficient time-series data storage, while the scheduler manages complex monitoring workflows.

