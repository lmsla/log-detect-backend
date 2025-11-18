# Log Detect Backend - Architecture & Component Interaction Diagram

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          CLIENT LAYER (Frontend)                         │
│                    Angular/React Dashboard                              │
│                                                                          │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
                        HTTP REST API (Gin)
                                 │
┌────────────────────────────────▼────────────────────────────────────────┐
│                     ROUTING & MIDDLEWARE LAYER                           │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Router (router/router.go)                                        │  │
│  │  - Route Groups: /auth, /Target, /Device, /Indices, /Receiver   │  │
│  │  - Middleware Chain: AuthMiddleware → PermissionMiddleware      │  │
│  │  - Swagger API Documentation                                    │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Middleware (middleware/auth.go)                                  │  │
│  │  - AuthMiddleware(): JWT token validation                        │  │
│  │  - OptionalAuthMiddleware(): Backward compatibility             │  │
│  │  - PermissionMiddleware(): RBAC enforcement                     │  │
│  │  - RequireRole(), AdminOnly(): Role-based access               │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                          │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────────────┐
│                      CONTROLLER LAYER (9 Controllers)                    │
│                                                                          │
│  auth.go          target.go        device.go        indices.go          │
│  ├─ Login         ├─ GetAllTargets  ├─ GetAllDevices ├─ GetAllIndices   │
│  ├─ Register      ├─ CreateTarget   ├─ CreateDevice  ├─ CreateIndices   │
│  ├─ GetProfile    ├─ UpdateTarget   ├─ UpdateDevice  ├─ UpdateIndices   │
│  └─ RefreshToken  └─ DeleteTarget   └─ DeleteDevice  └─ DeleteIndices   │
│                                                                          │
│  receiver.go      history.go       elasticsearch.go  dashboard.go       │
│  ├─ GetAllReceivers ├─ GetData      ├─ CreateMonitor  ├─ GetDashboard   │
│  ├─ CreateReceiver  ├─ GetLogname   ├─ GetStatus      ├─ GetStatistics  │
│  ├─ UpdateReceiver  └─ Monitoring   ├─ GetAlerts      └─ GetTrends      │
│  └─ DeleteReceiver                  └─ TestConnection                    │
│                                                                          │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────────────┐
│               BUSINESS LOGIC LAYER (22 Services)                         │
│                                                                          │
│  ┌─ auth.go (Authentication & RBAC)                                    │
│  │   - HashPassword(), CheckPassword()                                │
│  │   - GenerateJWT(), ValidateJWT()                                  │
│  │   - CreateDefaultRolesAndPermissions()                            │
│  │   - CheckPermission(), Login(), Register()                        │
│  │                                                                    │
│  ┌─ detect.go (Core Monitoring Logic)                                │
│  │   - Detect(): Query ES → Compare DB → Generate History            │
│  │   - Device auto-discovery and anomaly detection                   │
│  │   - Email alert generation for offline devices                    │
│  │                                                                    │
│  ┌─ center.go (Cron Job Scheduler)                                   │
│  │   - LoadCrontab(), ExecuteCrontab()                               │
│  │   - Control_center(): Bootstrap all targets                       │
│  │   - Dynamic cron job management                                   │
│  │                                                                    │
│  ┌─ device.go, target.go, indices.go, receiver.go (CRUD Services)   │
│  │   - Create, Read, Update, Delete operations                       │
│  │   - Database entity management                                    │
│  │                                                                    │
│  ┌─ history.go (Monitoring History)                                  │
│  │   - CreateHistory(), GetHistoryData()                             │
│  │   - TimescaleDB query support                                     │
│  │   - Historical data retrieval                                     │
│  │                                                                    │
│  ┌─ mail.go (Email Service)                                          │
│  │   - Mail4(): HTML email generation                                │
│  │   - SMTP client configuration                                     │
│  │   - Alert notification delivery                                   │
│  │                                                                    │
│  ┌─ batch_writer.go (Time-Series Data Pipeline)                      │
│  │   - AddHistory(): Batch queue accumulation                        │
│  │   - flushDeviceMetrics(): TimescaleDB bulk insert                │
│  │   - flushESMetrics(): ES metrics batch write                      │
│  │   - Periodic flushing (size or time-based)                        │
│  │                                                                    │
│  ┌─ es_monitor_service.go (ES Health Monitoring)                     │
│  │   - CreateESMonitor(), UpdateESMonitor()                          │
│  │   - GetESMonitorByID(), DeleteESMonitor()                         │
│  │   - Configuration persistence                                     │
│  │                                                                    │
│  ┌─ es_scheduler.go (ES Monitor Scheduling)                          │
│  │   - InitESScheduler(): Singleton initialization                   │
│  │   - LoadAllMonitors(): Startup loading                            │
│  │   - StartMonitor(), StopMonitor()                                 │
│  │   - Time-based health check triggering                            │
│  │                                                                    │
│  ┌─ es_query.go (Elasticsearch Queries)                              │
│  │   - SearchRequest(): DSL query execution                          │
│  │   - Aggregation on configurable fields                            │
│  │   - Date range filtering                                          │
│  │                                                                    │
│  ├─ sqltable.go (Schema Migration)                                   │
│  ├─ setting.go (YAML Config Loading)                                │
│  ├─ tools.go (Helper Functions)                                     │
│  └─ dashboard.go (Analytics & Visualization)                         │
│                                                                    │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────────────┐
│                  DATA MODEL LAYER (Entities & Models)                    │
│                                                                          │
│  Authentication Models       Monitoring Models        Analytics Models   │
│  ├─ User                    ├─ Target               ├─ History          │
│  ├─ Role                    ├─ Index                ├─ HistoryDailyStats│
│  ├─ Permission              ├─ Device               ├─ AlertHistory     │
│  └─ LoginRequest            ├─ Receiver             └─ DashboardData    │
│                             ├─ IndicesTargets                          │
│                             ├─ CronList                                │
│                             ├─ MailHistory                             │
│                             ├─ ElasticsearchMonitor                    │
│                             ├─ ESMetric                                │
│                             └─ ESAlert                                 │
│                                                                          │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────────────┐
│                    CLIENT LAYER (Database Connections)                   │
│                                                                          │
│  ┌──────────────────────┐  ┌──────────────────────┐  ┌───────────────┐  │
│  │   MySQL (GORM)       │  │  TimescaleDB (SQL)   │  │ Elasticsearch │  │
│  ├──────────────────────┤  ├──────────────────────┤  ├───────────────┤  │
│  │ - Database: logdetect│  │ - Database: monitoring│ │ - ES v8        │  │
│  │ - Connection pooling │  │ - Time-series data   │  │ - Health check │  │
│  │ - Transaction support│  │ - Hypertable storage │  │ - SSL/TLS      │  │
│  │ - AutoMigrate schema │  │ - Batch insert       │  │ - Auth support │  │
│  │                      │  │ - High compression   │  │ - DSL queries  │  │
│  │ Tables:              │  │                      │  │                │  │
│  │ ├─ users             │  │ Tables:              │  │ Indices:       │  │
│  │ ├─ roles             │  │ ├─ device_metrics    │  │ ├─ logs*       │  │
│  │ ├─ permissions       │  │ ├─ es_metrics        │  │ ├─ events*     │  │
│  │ ├─ devices           │  │ └─ alert_history     │  │ └─ monitoring* │  │
│  │ ├─ targets           │  │                      │  │                │  │
│  │ ├─ indices           │  │ Connection Pool:     │  │ Features:      │  │
│  │ ├─ history           │  │ ├─ Max Idle: 10      │  │ ├─ Aggregation │  │
│  │ ├─ alert_history     │  │ ├─ Max Open: 100     │  │ ├─ Filtering   │  │
│  │ └─ elasticsearch...  │  │ └─ Max Lifetime: 1h  │  │ └─ Faceting    │  │
│  │                      │  │                      │  │                │  │
│  │ Connection Pool:     │  │ Driver: pq           │  │ Client:        │  │
│  │ ├─ Max Idle: 10      │  │ (PostgreSQL)         │  │ elastic v8     │  │
│  │ ├─ Max Open: 100     │  │                      │  │                │  │
│  │ └─ Max Lifetime: 1h  │  │                      │  │                │  │
│  │                      │  │                      │  │                │  │
│  │ Driver: MySQL        │  │ Timezone:            │  │ URL:           │  │
│  │ (gorm/mysql)         │  │ Asia/Taipei          │  │ 10.99.1.213:9200 │
│  │                      │  │                      │  │                │  │
│  │ DSN: runner:pwd@     │  │ DSN: logdetect:pwd@  │  │ SSL:           │  │
│  │      10.99.1.133/... │  │      10.99.1.213/... │  │ InsecureSkip   │  │
│  │                      │  │                      │  │                │  │
│  └──────────────────────┘  └──────────────────────┘  └───────────────┘  │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘

                        ┌──────────────────────┐
                        ��   Global Config      │
                        ├──────────────────────┤
                        │ - EnvConfig          │
                        │ - Mysql (GORM)       │
                        │ - TimescaleDB        │
                        │ - Elasticsearch      │
                        │ - Crontab            │
                        │ - BatchWriter        │
                        └──────────────────────┘
```

---

## Data Flow Diagrams

### 1. Device Monitoring Detection Flow

```
┌─────────────────────────────┐
│   Cron Job Trigger          │
│   (Every N minutes/hours)   │
└────────────┬────────────────┘
             │
             ▼
┌─────────────────────────────┐
│  Detect() Function          │
├─────────────────────────────┤
│ 1. Calculate time range     │
│ 2. Query ES for devices     │
│    using aggregation        │
│ 3. Get device list from DB  │
└────────────┬────────────────┘
             │
      ┌──────┴──────┐
      │             │
      ▼             ▼
┌─────────────┐ ┌──────────────┐
│ ES Results  │ │ DB Devices   │
└──────┬──────┘ └────────┬─────┘
       │                │
       └────────┬───────┘
                ▼
        ┌──────────────────────┐
        │  ListCompare()       │
        │  Detect:             │
        │  - Added devices     │
        │  - Removed devices   │
        │  - Intersection      │
        └──────────┬───────────┘
                   │
      ┌────────────┼────────────┐
      │            │            │
      ▼            ▼            ▼
  ┌────────┐  ┌────────┐  ┌──────────┐
  │ Create │  │ Store  │  │  Send    │
  │ History│  │ History│  │ Alerts   │
  └───┬────┘  └──┬─────┘  └────┬─────┘
      │          │             │
      │          ▼             │
      │      ┌───────────────┐ │
      │      │   MySQL       │ │
      │      │  history      │ │
      │      │   table       │ │
      │      └────────┬──────┘ │
      │               │        │
      │               ▼        │
      │          ┌──────────┐  │
      │          │BatchWriter  │
      │          │  Queue   │  │
      │          └────┬─────┘  │
      │               │        │
      │        Flush (size/time)
      │               │        │
      │               ▼        │
      │          ┌──────────┐  │
      │          │TimescaleDB  │
      │          │device_metrics
      │          └──────────┘  │
      │                        │
      └────────────┬───────────┘
                   │
              ┌────▼────┐
              │  SMTP   │
              │  Client │
              └────┬────┘
                   │
                   ▼
            ┌─────────────┐
            │   Email     │
            │ Notification│
            └─────────────┘
```

### 2. Elasticsearch Monitoring Flow

```
┌────────────────────────────────┐
│  ESMonitorScheduler            │
│  (Per Monitor Interval)        │
└────────────┬───────────────────┘
             │
             ▼
┌────────────────────────────────┐
│  MonitorESCluster()            │
├────────────────────────────────┤
│ 1. Query ES /_cluster/health   │
│ 2. Get node stats              │
│ 3. Collect metrics:            │
│    - CPU, Memory, Disk usage   │
│    - Response time             │
│    - Shard status              │
└────────────┬───────────────────┘
             │
             ▼
┌────────────────────────────────┐
│  Threshold Checking            │
│  (ESAlert Logic)               │
├────────────────────────────────┤
│ Check against:                 │
│ - CPU High/Critical            │
│ - Memory High/Critical         │
│ - Disk High/Critical           │
│ - Response Time High/Critical  │
│ - Unassigned Shards            │
└────────────┬───────────────────┘
             │
      ┌──────┴──────┐
      │             │
    No            Alert
   Issue        Triggered
      │             │
      │             ▼
      │     ┌──────────────────┐
      │     │ Check Dedup Window
      │     │ (Default: 5 min) │
      │     └────────┬─────────┘
      │              │
      │              ▼
      │     ┌──────────────────┐
      │     │ Store ESAlert    │
      │     │ to TimescaleDB   │
      │     │ alert_history    │
      │     └────────┬─────────┘
      │              │
      │              ▼
      │     ┌──────────────────┐
      │     │  Send Notification
      │     │  (Email)         │
      │     └──────────────────┘
      │
      └─────────────┬─────────────┘
                    │
                    ▼
          ┌─────────────────────┐
          │  Store ESMetric     │
          │  to TimescaleDB     │
          │  es_metrics table   │
          │  (via BatchWriter)  │
          └─────────────────────┘
```

### 3. Authentication & Authorization Flow

```
┌──────────────────────┐
│  User Login Request  │
│  POST /auth/login    │
│  {username, password}│
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ AuthService.Login()  │
├──────────────────────┤
│ 1. Query User from DB│
│ 2. bcrypt.Compare()  │
│    password          │
└──────────┬───────────┘
           │
      ┌────┴─────┐
      │           │
   Valid      Invalid
   Pass        Pass
      │           │
      ▼           ▼
┌──────────┐  ┌──────────┐
│Generate  │  │ 401 Error│
│ JWT      │  │ Response │
└────┬─────┘  └──────────┘
     │
     ▼
┌──────────────────────────┐
│ JWT Token Created        │
│ HS256 Signed             │
│ 24-hour expiration       │
│ Claims: user_id,         │
│         username,        │
│         role_id,         │
│         exp, iat         │
└────┬─────────────────────┘
     │
     ▼
┌──────────────────────┐
│ Return LoginResponse │
│ {token, user}        │
└──────┬───────────────┘
       │
   [Client stores JWT in localStorage]
       │
       ▼
┌──────────────────────────────┐
│ Subsequent API Request       │
│ Authorization: Bearer {token}│
└──────────┬───────────────────┘
           │
           ▼
┌──────────────────────────────┐
│ AuthMiddleware checks:       │
│ 1. Extract token from header │
│ 2. jwt.Parse()               │
│ 3. Verify signature          │
│ 4. Check expiration          │
└──────────┬───────────────────┘
           │
      ┌────┴─────┐
      │           │
    Valid    Invalid/Expired
     │           │
     ▼           ▼
  ┌────┐    ┌───────────┐
  │ Set│    │ 401 Abort │
  │user│    │ Request   │
  │context   └───────────┘
  └──┬─┘
     │
     ▼
┌──────────────────────────────┐
│ PermissionMiddleware checks: │
│ 1. Get user from context     │
│ 2. Get user role             │
│ 3. Get role permissions      │
│ 4. Check resource.action     │
└──────────┬───────────────────┘
           │
      ┌────┴─────┐
      │           │
   Has        No
 Permission  Permission
      │           │
      ▼           ▼
  ┌────┐    ┌───────────┐
  │Next│    │ 403 Abort │
  │Handler   │ Request   │
  └────┘    └───────────┘
```

### 4. Data Consistency & Batch Processing

```
┌─────────────────────────┐
│ Monitoring Results      │
│ (History Entity)        │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ BatchWriter.AddHistory()│
└────────────┬────────────┘
             │
      ┌──────┴──────┐
      │             │
  Device        ES Metric
  History       (ESMetric)
      │             │
      ▼             ▼
┌──────────┐  ┌──────────┐
│ batch[]  │  │esBatch[] │
│ queue    │  │ queue    │
│ (max 50) │  │ (max 50) │
└────┬─────┘  └────┬─────┘
     │             │
     └──────┬──────┘
            │
      ┌─────▼─────┐
      │ Flush      │
      │ Decision   │
      └─────┬─────┐
            │ │
       Size │ │ Timer (5s)
      >=50  │ │
            │ │
            └─┴────┐
                   │
             ┌─────▼──────────┐
             │ Begin Txn      │
             │ TimescaleDB    │
             └────────┬───────┘
                      │
         ┌────────────┼────────────┐
         │            │            │
         ▼            ▼            ▼
    ┌────────┐   ┌────────┐   ┌────────┐
    │Insert  │   │Insert  │   │Insert  │
    │device_ │   │es_     │   │to MySQL│
    │metrics │   │metrics │   │history │
    │(batch) │   │(batch) │   │        │
    └───┬────┘   └───┬────┘   └───┬────┘
        │            │            │
        └────────────┼────────────┘
                     │
                     ▼
         ┌──────────────────────┐
         │ Commit Transaction   │
         │ (atomic)             │
         └──────────┬───────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
      Commit              Rollback (error)
        │                       │
        ▼                       ▼
   ┌────────┐           ┌──────────────┐
   │ Success│           │ Retry or Log │
   │ Log    │           │ Error        │
   └────────┘           └──────────────┘
```

---

## Component Dependency Graph

```
                    main.go (Startup)
                         │
         ┌───────────────┼───────────────┐
         │               │               │
    clients/      services/      router/
    └─ mysql.go   └─ sqltable.go   └─ router.go
    └─ es.go      └─ auth.go       └─ handlers
    └─ timescale  └─ detect.go
                  └─ center.go
                  └─ device.go
                  └─ target.go
                  └─ indices.go
                  └─ receiver.go
                  └─ history.go
                  └─ batch_writer.go
                  └─ es_monitor*.go
                  └─ mail.go
                  └─ dashboard.go
                  └─ tools.go
                       │
                 ┌─────┴─────┐
                 │           │
            entities/     models/
            ├─ User        ├─ Response
            ├─ Role        ├─ Common
            ├─ Permission  └─ targets.go
            ├─ Device
            ├─ Target
            ├─ Index
            ├─ Receiver
            ├─ History
            ├─ AlertHistory
            ├─ HistoryDailyStats
            └─ ElasticsearchMonitor*
                     │
                     ▼
        ┌─────────────────────────┐
        │   global/global.go      │
        │                         │
        │ Global State:           │
        │ - EnvConfig             │
        │ - Mysql (GORM)          │
        │ - TimescaleDB           │
        │ - Elasticsearch         │
        │ - Crontab              │
        │ - BatchWriter           │
        └─────────────────────────┘
```

---

## Configuration & Environment Setup

```
┌─────────────────────────────────────┐
│  config.yml (Main Configuration)    │
├─────────────────────────────────────┤
│  database:                          │
│    - MySQL connection details       │
│    - Connection pooling settings    │
│                                     │
│  timescale:                         │
│    - PostgreSQL/TimescaleDB details │
│    - Batch write settings           │
│                                     │
│  batch_writer:                      │
│    - Enabled: true/false            │
│    - Batch size: 50                 │
│    - Flush interval: 5s             │
│                                     │
│  es:                                │
│    - Elasticsearch URL              │
│    - Authentication credentials     │
│                                     │
│  email:                             │
│    - SMTP configuration             │
│    - Gmail settings                 │
│                                     │
│  server:                            │
│    - Port: 8006                     │
│    - Mode: debug/release            │
│                                     │
│  sso:                               │
│    - Keycloak integration (optional)│
│                                     │
│  cors:                              │
│    - Allow origins                  │
│    - Headers configuration          │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  setting.yml (Monitoring Config)    │
├─────────────────────────────────────┤
│  targets:                           │
│    - Receiver email lists           │
│    - Subject line templates         │
│    - Indices:                       │
│      - Index pattern                │
│      - Logname                      │
│      - Period (minutes/hours)       │
│      - Unit (frequency)             │
│      - Field (aggregation field)    │
└─────────────────────────────────────┘
```

---

## Request/Response Flow Example

### Create Monitoring Target

```
HTTP REQUEST
POST /api/v1/Target/Create
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "subject": "Production WebServer",
  "to": ["admin@company.com"],
  "enable": true,
  "indices": [
    {
      "id": 1,
      "pattern": "logstash-prod-*",
      "logname": "webserver",
      "period": "minutes",
      "unit": 5,
      "field": "host.keyword"
    }
  ]
}

↓ router/router.go routes to controller
↓ middleware/auth.go validates JWT
↓ middleware/auth.go validates permission (target.create)
↓ controller/target.go → CreateTarget handler

HTTP RESPONSE (201 Created)
{
  "id": 42,
  "subject": "Production WebServer",
  "to": ["admin@company.com"],
  "enable": true,
  "indices": [...],
  "created_at": 1698650000,
  "updated_at": 1698650000
}

SIDE EFFECTS:
1. MySQL: Insert into targets table
2. MySQL: Insert into indices_targets table
3. Crontab: Register cron job "*/5 * * * *"
4. MySQL: Insert into cron_lists table with entry_id
5. Services: Start detect() execution on schedule
```

---

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Production Environment                │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │  Go Application (log-detect-backend)           │    │
│  │  ├─ Binary: main                               │    │
│  │  ├─ Listening on: 0.0.0.0:8006                │    │
│  │  ├─ Running services:                          │    │
│  │  │  - Gin HTTP Server                          │    │
│  │  │  - Cron Job Scheduler                       │    │
│  │  │  - ES Monitor Scheduler                     │    │
│  │  │  - Batch Writer Service                     │    │
│  │  └─ Config files:                              │    │
│  │     - config.yml                               │    │
│  │     - setting.yml                              │    │
│  └────────────────────────────────────────────────┘    │
│         │                    │                │         │
│         ▼                    ▼                ▼         │
│  ┌─────────────┐     ┌─────────────┐  ┌──────────┐   │
│  │   MySQL     │     │TimescaleDB  │  │ ES       │   │
│  │  logdetect  │     │ monitoring  │  │ cluster  │   │
│  │             │     │             │  │          │   │
│  │ Config +    │     │ Metrics +   │  │ Logs +   │   │
│  │ History     │     │ Alerts      │  │ Events   │   │
│  └─────────────┘     └─────────────┘  └──────────┘   │
│       │                    │                │         │
└───────┼────────────────────┼────────────────┼─────────┘
        │                    │                │
        └────────────────────┼────────────────┘
                             │
                    ┌────────▼────────┐
                    │  SMTP Gateway   │
                    │  (Gmail/Custom) │
                    └────────┬────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │  Email Clients  │
                    └─────────────────┘
```

