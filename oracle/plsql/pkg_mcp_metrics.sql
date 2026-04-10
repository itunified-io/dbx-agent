-- pkg_mcp_metrics.sql
-- PL/SQL package for server-side metric collection into MCP_METRIC_STAGING.
-- The dbx-agent reads from this staging table via SQL when direct V$ access is restricted.

CREATE OR REPLACE PACKAGE DBX_MONITOR.PKG_MCP_METRICS AS
    PROCEDURE collect_session_metrics;
    PROCEDURE collect_wait_events;
    PROCEDURE collect_tablespace_metrics;
    PROCEDURE collect_sga_pga_metrics;
    PROCEDURE collect_redo_undo_metrics;
    PROCEDURE collect_blocking_sessions;
    PROCEDURE collect_all;
    PROCEDURE purge_staging(p_retention_hours IN NUMBER DEFAULT 24);
END PKG_MCP_METRICS;
/

CREATE OR REPLACE PACKAGE BODY DBX_MONITOR.PKG_MCP_METRICS AS

    PROCEDURE collect_session_metrics IS
    BEGIN
        INSERT INTO MCP_METRIC_STAGING (metric_name, metric_value, labels)
        SELECT 'oracle_session_active',
               COUNT(*),
               '{"wait_class":"' || NVL(wait_class, 'CPU') || '"}'
        FROM v$session
        WHERE type = 'USER' AND status = 'ACTIVE'
        GROUP BY wait_class;
    END;

    PROCEDURE collect_wait_events IS
    BEGIN
        INSERT INTO MCP_METRIC_STAGING (metric_name, metric_value, labels)
        SELECT 'oracle_wait_event_time_ms',
               time_waited_micro / 1000,
               '{"event":"' || event || '","wait_class":"' || wait_class || '"}'
        FROM v$system_event
        WHERE wait_class NOT IN ('Idle')
        ORDER BY time_waited_micro DESC
        FETCH FIRST 20 ROWS ONLY;
    END;

    PROCEDURE collect_tablespace_metrics IS
    BEGIN
        INSERT INTO MCP_METRIC_STAGING (metric_name, metric_value, labels)
        SELECT 'oracle_tablespace_used_pct',
               ROUND(used_percent, 2),
               '{"tablespace":"' || tablespace_name || '"}'
        FROM dba_tablespace_usage_metrics;
    END;

    PROCEDURE collect_sga_pga_metrics IS
    BEGIN
        INSERT INTO MCP_METRIC_STAGING (metric_name, metric_value, labels)
        SELECT 'oracle_sga_total_bytes', SUM(value), '{}'
        FROM v$sga;

        INSERT INTO MCP_METRIC_STAGING (metric_name, metric_value, labels)
        SELECT 'oracle_pga_allocated_bytes', value, '{}'
        FROM v$pgastat
        WHERE name = 'total PGA allocated';
    END;

    PROCEDURE collect_redo_undo_metrics IS
    BEGIN
        INSERT INTO MCP_METRIC_STAGING (metric_name, metric_value, labels)
        SELECT 'oracle_redo_switches_last_hour', COUNT(*), '{}'
        FROM v$log_history
        WHERE first_time > SYSDATE - 1/24;

        INSERT INTO MCP_METRIC_STAGING (metric_name, metric_value, labels)
        SELECT 'oracle_undo_retention_seconds', value, '{}'
        FROM v$parameter
        WHERE name = 'undo_retention';
    END;

    PROCEDURE collect_blocking_sessions IS
    BEGIN
        INSERT INTO MCP_METRIC_STAGING (metric_name, metric_value, labels)
        SELECT 'oracle_blocking_chain_count',
               COUNT(DISTINCT blocking_session),
               '{}'
        FROM v$session
        WHERE blocking_session IS NOT NULL;
    END;

    PROCEDURE collect_all IS
    BEGIN
        FOR rec IN (SELECT check_name, enabled FROM MCP_METRIC_CONFIG) LOOP
            IF rec.enabled = 1 THEN
                CASE rec.check_name
                    WHEN 'session_metrics'    THEN collect_session_metrics;
                    WHEN 'wait_events'        THEN collect_wait_events;
                    WHEN 'tablespace_metrics' THEN collect_tablespace_metrics;
                    WHEN 'sga_pga_metrics'    THEN collect_sga_pga_metrics;
                    WHEN 'redo_undo_metrics'  THEN collect_redo_undo_metrics;
                    WHEN 'blocking_sessions'  THEN collect_blocking_sessions;
                    ELSE NULL;
                END CASE;
            END IF;
        END LOOP;
        COMMIT;
    END;

    PROCEDURE purge_staging(p_retention_hours IN NUMBER DEFAULT 24) IS
    BEGIN
        DELETE FROM MCP_METRIC_STAGING
        WHERE collected_at < SYSTIMESTAMP - NUMTODSINTERVAL(p_retention_hours, 'HOUR');
        COMMIT;
    END;

END PKG_MCP_METRICS;
/
