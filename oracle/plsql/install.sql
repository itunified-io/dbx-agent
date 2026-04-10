-- install.sql
-- Master installation script for dbx-agent Oracle PL/SQL companion.
-- Run as SYS or DBA with CREATE USER privileges.
--
-- Usage: sqlplus / as sysdba @install.sql

PROMPT ========================================
PROMPT dbx-agent PL/SQL Companion Installation
PROMPT ========================================

PROMPT [1/4] Creating DBX_MONITOR schema and staging table...
@@mcp_metric_staging.sql

PROMPT [2/4] Creating metric configuration table...
@@mcp_metric_config.sql

PROMPT [3/4] Creating PKG_MCP_METRICS package...
@@pkg_mcp_metrics.sql

PROMPT [4/4] Creating DBMS_SCHEDULER jobs...
@@mcp_agent_job.sql

PROMPT ========================================
PROMPT Installation complete.
PROMPT Verify: SELECT job_name, state FROM dba_scheduler_jobs WHERE owner = 'DBX_MONITOR';
PROMPT ========================================
