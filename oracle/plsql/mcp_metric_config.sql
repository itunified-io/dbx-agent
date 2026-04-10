-- mcp_metric_config.sql
-- Configuration table for enabling/disabling metric checks and setting intervals.

CREATE TABLE DBX_MONITOR.MCP_METRIC_CONFIG (
    check_name    VARCHAR2(64)  PRIMARY KEY,
    enabled       NUMBER(1)     DEFAULT 1 NOT NULL,
    interval_sec  NUMBER        DEFAULT 60 NOT NULL
);

INSERT INTO DBX_MONITOR.MCP_METRIC_CONFIG VALUES ('session_metrics', 1, 15);
INSERT INTO DBX_MONITOR.MCP_METRIC_CONFIG VALUES ('wait_events', 1, 15);
INSERT INTO DBX_MONITOR.MCP_METRIC_CONFIG VALUES ('tablespace_metrics', 1, 60);
INSERT INTO DBX_MONITOR.MCP_METRIC_CONFIG VALUES ('sga_pga_metrics', 1, 30);
INSERT INTO DBX_MONITOR.MCP_METRIC_CONFIG VALUES ('redo_undo_metrics', 1, 30);
INSERT INTO DBX_MONITOR.MCP_METRIC_CONFIG VALUES ('blocking_sessions', 1, 15);
COMMIT;
