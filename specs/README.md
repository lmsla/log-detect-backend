# Log Detect Backend - Complete Specification Documentation

Welcome to the comprehensive analysis of the Log Detect Backend codebase. This directory contains complete documentation generated through very thorough analysis of the entire codebase.

## Documentation Files

### 1. CODEBASE_ANALYSIS.md (40 KB, 1334 lines)
**Comprehensive technical analysis of the entire codebase**

This is the most detailed document covering:
- Executive summary and overall architecture
- Complete project structure and technology stack
- All 22 services with detailed function descriptions
- 9 controllers with all endpoints
- Database schema (MySQL and TimescaleDB tables)
- Complete API endpoint reference (60+ routes)
- Authentication & authorization mechanisms
- External integrations (Elasticsearch, email, SSO)
- Business logic workflows
- Design patterns used
- Security considerations
- Scalability & performance notes
- Configuration file formats

**Use this for**: Deep understanding of how the system works, architectural decisions, data models, and integration points.

---

### 2. ARCHITECTURE_DIAGRAMS.md (38 KB, 703 lines)
**Visual architecture and data flow diagrams**

This document contains:
- System architecture overview diagram
- Layered architecture visualization
- Device monitoring detection flow
- Elasticsearch monitoring flow
- Authentication & authorization flow
- Batch processing & data consistency flow
- Component dependency graph
- Configuration structure diagrams
- Request/response flow examples
- Deployment architecture
- Database connection topology

**Use this for**: Understanding how components interact, visualizing data flows, and seeing the big picture of system design.

---

### 3. QUICK_REFERENCE.md (13 KB, 517 lines)
**Quick lookup guide for developers and operators**

This reference guide includes:
- Project overview
- Key features at a glance
- Core data models
- Important file locations
- API endpoints quick list
- Workflow examples (device monitoring, ES monitoring, batch writing)
- Key services & responsibilities
- Common operations
- Database connection details
- JWT token details
- Permission model
- Startup sequence
- Debugging tips
- File statistics
- Technology versions

**Use this for**: Quick lookups, API integration, debugging, and onboarding new team members.

---

## Quick Navigation

### For Different Audiences

**System Architects**
- Start with: ARCHITECTURE_DIAGRAMS.md
- Then read: CODEBASE_ANALYSIS.md sections 1-3, 12

**Backend Developers**
- Start with: QUICK_REFERENCE.md
- Then read: CODEBASE_ANALYSIS.md sections 2-6, 8
- Reference: ARCHITECTURE_DIAGRAMS.md for flows

**DevOps/Operations**
- Start with: QUICK_REFERENCE.md (Startup Sequence, Database Connection Details)
- Then read: CODEBASE_ANALYSIS.md sections 1, 7
- Reference: ARCHITECTURE_DIAGRAMS.md (Deployment Architecture)

**API Consumers**
- Start with: QUICK_REFERENCE.md (API Endpoints Quick List)
- Then read: CODEBASE_ANALYSIS.md section 5 (API Endpoints & Routing)
- Reference: ARCHITECTURE_DIAGRAMS.md (Request/Response Flow Example)

**QA/Testers**
- Start with: QUICK_REFERENCE.md (Workflow Examples, Common Operations)
- Then read: CODEBASE_ANALYSIS.md sections 3, 6, 8
- Reference: ARCHITECTURE_DIAGRAMS.md (Data Flow Diagrams)

---

## Key Statistics

| Metric | Value |
|--------|-------|
| Total Documentation Lines | 2,554 |
| Total Documentation Size | 91 KB |
| Services Documented | 22 |
| Controllers Documented | 9 |
| Data Models | 20+ |
| API Endpoints | 60+ |
| Database Tables | 16+ |
| External Integrations | 3 (ES, MySQL, TimescaleDB) |

---

## System Overview

**Log Detect Backend** is a comprehensive monitoring system built in Go that:

1. **Monitors Device Health**: Queries Elasticsearch for device status, compares against database, and alerts on changes
2. **Monitors Elasticsearch Health**: Tracks ES cluster health (CPU, memory, disk, shards)
3. **Manages Users & Access**: RBAC with JWT authentication
4. **Stores Time-Series Data**: Efficient batch writing to TimescaleDB
5. **Sends Alerts**: HTML-formatted email notifications via SMTP
6. **Provides Analytics**: Dashboard with statistics, trends, and timeline visualizations
7. **Exposes REST API**: 60+ endpoints for management and monitoring

**Technology Stack**:
- Language: Go 1.24
- Web Framework: Gin
- Databases: MySQL (config) + TimescaleDB (metrics) + Elasticsearch (logs)
- Scheduler: robfig/cron
- Auth: JWT + bcrypt + RBAC
- Email: SMTP
- Docs: Swagger

---

## Core Workflows

### Device Monitoring Workflow
1. Admin creates Target (email recipients) and Index (ES pattern)
2. System registers cron job automatically
3. On schedule: Detect() queries ES, compares with DB
4. Creates History records and email alerts
5. Stores metrics via BatchWriter to TimescaleDB

### Elasticsearch Health Monitoring Workflow
1. Admin creates ElasticsearchMonitor with check interval
2. System starts periodic health checks
3. Collects CPU, memory, disk, shard data
4. Compares against configurable thresholds
5. Creates alerts and stores metrics

### Batch Writing Workflow
1. Services generate History or ESMetric records
2. BatchWriter queues them in memory
3. Flush when: size >= 50 records OR 5 seconds elapsed
4. Atomic insert to TimescaleDB via prepared statement
5. Log success count and continue

---

## Database Architecture

### MySQL (logdetect database)
- **Configuration**: Users, Roles, Permissions, Devices, Targets, Indices
- **History**: Monitoring results, alerts, mail logs, cron job tracking
- **Purpose**: System configuration and operational state

### TimescaleDB (monitoring database)
- **device_metrics**: Time-series device monitoring data (hypertable)
- **es_metrics**: Time-series Elasticsearch metrics (hypertable)
- **alert_history**: Alert records with timestamps
- **Purpose**: Efficient storage and querying of time-series data

### Elasticsearch
- **Indices**: logstash-*, custom patterns
- **Purpose**: Log storage and device discovery
- **Integration**: Device monitoring queries logs for hosts

---

## API Structure

- **Public**: POST /auth/login
- **Protected**: All /api/v1/* routes require JWT + optional permission checks
- **Middleware Chain**: AuthMiddleware → PermissionMiddleware
- **Documentation**: Swagger available at /swagger/

**Route Groups**:
- /auth - User authentication (8 endpoints)
- /Target - Monitoring targets (4 endpoints)
- /Device - Device management (6 endpoints)
- /Indices - Index patterns (7 endpoints)
- /Receiver - Email receivers (4 endpoints)
- /History - Historical data (2 endpoints)
- /dashboard - Analytics (10+ endpoints)
- /elasticsearch - ES monitoring (13 endpoints)

---

## Authentication & Authorization

**JWT Token Flow**:
1. User logs in with credentials
2. System validates against bcrypt hash in MySQL
3. Generates HS256-signed JWT (24-hour expiration)
4. Client includes Bearer token in Authorization header
5. Middleware validates signature and expiration
6. PermissionMiddleware checks resource.action permission

**RBAC Model**:
- User → Role (1-to-1)
- Role → Permissions (many-to-many)
- Permission = Resource + Action (e.g., "device.create")

**Default Roles**:
- admin: All permissions
- user: Read-only
- operator: CRUD on assigned resources

---

## Configuration Files

### config.yml
Main configuration with sections for:
- database (MySQL connection)
- timescale (TimescaleDB connection)
- batch_writer (size: 50, interval: 5s)
- es (Elasticsearch connection)
- email (SMTP configuration)
- server (port: 8006, mode: debug/release)
- cors (allowed origins)
- sso (Keycloak integration)

### setting.yml
Monitoring configuration with:
- targets: List of monitoring configurations
  - receiver: Email recipients
  - subject: Alert subject
  - indices: ES index patterns and frequency

---

## Getting Started

### Understanding the System
1. Read: QUICK_REFERENCE.md (Project Overview, Key Features)
2. Study: ARCHITECTURE_DIAGRAMS.md (System Architecture, Data Flows)
3. Dive Deep: CODEBASE_ANALYSIS.md (Complete details)

### Setting Up Monitoring
1. Configure database connections in config.yml
2. Create Indices (ES patterns) via API or UI
3. Create Targets (email recipients) via API or UI
4. Link Indices to Targets
5. Enable targets and wait for cron scheduling

### Troubleshooting
1. Check: QUICK_REFERENCE.md (Debugging Tips, Common Issues)
2. Monitor: MySQL (history, cron_lists) and logs
3. Verify: Elasticsearch connectivity, SMTP settings

---

## Important Concepts

**Device Monitoring**: System queries Elasticsearch for devices matching a pattern, compares with database Device list, detects changes (new/offline/online), and sends alerts

**Batch Writing**: Memory accumulation of records flushed in batches to reduce database overhead and improve performance

**Cron Jobs**: Dynamic registration of periodic monitoring tasks, stored in cron_lists table, managed by robfig/cron scheduler

**Time-Series Data**: Efficient storage in TimescaleDB using hypertables for device metrics and ES metrics with automatic time-based partitioning

**RBAC**: Multi-level access control with User → Role → Permissions model, enforced at API endpoint level

---

## Documentation Maintenance

These documents were generated through comprehensive static code analysis on: **October 29, 2024**

The analysis covered:
- All 22 service files
- All 9 controller files
- All entity and model definitions
- Router and middleware configuration
- Client/database connections
- Configuration file formats
- Database schema definitions
- API endpoint routing

---

## Document Location

All documentation is stored in: `/specs/`

- CODEBASE_ANALYSIS.md
- ARCHITECTURE_DIAGRAMS.md
- QUICK_REFERENCE.md
- README.md (this file)

---

## Related Documentation

See also:
- `/docs/` - Swagger API documentation (generated)
- README_AUTH.md - Authentication details
- DATA_MANAGEMENT_GUIDE.md - Data archival and management
- TROUBLESHOOTING.md - Common issues and solutions
- `go.mod` - Dependency versions

---

This documentation provides a complete reference for understanding the Log Detect Backend system architecture, components, and functionality. Use the table of contents above to find the information most relevant to your needs.
