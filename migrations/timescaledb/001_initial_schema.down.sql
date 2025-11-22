-- Rollback TimescaleDB initial schema
-- Version: 001

DROP TABLE IF EXISTS es_alerts;
DROP TABLE IF EXISTS es_metrics;
