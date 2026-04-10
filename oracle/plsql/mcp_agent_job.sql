-- mcp_agent_job.sql
-- DBMS_SCHEDULER job that runs PKG_MCP_METRICS.collect_all every 15 seconds.
-- Also creates a daily purge job (24h retention).

BEGIN
    DBMS_SCHEDULER.CREATE_JOB(
        job_name        => 'DBX_MONITOR.MCP_METRIC_COLLECTOR',
        job_type        => 'PLSQL_BLOCK',
        job_action      => 'BEGIN DBX_MONITOR.PKG_MCP_METRICS.collect_all; END;',
        start_date      => SYSTIMESTAMP,
        repeat_interval => 'FREQ=SECONDLY;INTERVAL=15',
        enabled         => TRUE,
        comments        => 'dbx-agent: collect metrics into staging table every 15s'
    );
END;
/

BEGIN
    DBMS_SCHEDULER.CREATE_JOB(
        job_name        => 'DBX_MONITOR.MCP_STAGING_PURGE',
        job_type        => 'PLSQL_BLOCK',
        job_action      => 'BEGIN DBX_MONITOR.PKG_MCP_METRICS.purge_staging(24); END;',
        start_date      => SYSTIMESTAMP,
        repeat_interval => 'FREQ=HOURLY;INTERVAL=1',
        enabled         => TRUE,
        comments        => 'dbx-agent: purge staging table entries older than 24h'
    );
END;
/
