package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	dto "github.com/prometheus/client_model/go"
)

var serverNameLabel = "server"

// getFamilyDefinitions definitions for metric families produced by the check.
func getFamilyDefinitions() map[string]*dto.MetricFamily {
	sp := func(s string) *string { return &s }

	return map[string]*dto.MetricFamily{
		"uptime": {
			Name: sp("uptime"),
			Help: sp("The number of seconds that the server has been up"),
			Type: dto.MetricType_COUNTER.Enum(),
		},
		"connection_errors_internal": {
			Name: sp("connection_errors_internal"),
			Help: sp("The number of connections refused due to internal errors in the server, such as failure to start a new thread or an out-of-memory condition."),
			Type: dto.MetricType_COUNTER.Enum(),
		},
		"connection_errors_max_connections": {
			Name: sp("connection_errors_max_connections"),
			Help: sp("The number of connections refused because the server max_connections limit was reached."),
			Type: dto.MetricType_COUNTER.Enum(),
		},
		"slow_queries": {
			Name: sp("slow_queries"),
			Help: sp("The number of queries that have taken more than long_query_time seconds. "),
			Type: dto.MetricType_COUNTER.Enum(),
		},
		"queries": {
			Name: sp("queries"),
			Help: sp("The number of statements executed by the server."),
			Type: dto.MetricType_COUNTER.Enum(),
		},
		"innodb_data_fsyncs": {
			Name: sp("innodb_data_fsyncs"),
			Help: sp("The number of fsync() operations so far."),
			Type: dto.MetricType_COUNTER.Enum(),
		},
		"innodb_row_lock_waits": {
			Name: sp("innodb_row_lock_waits"),
			Help: sp("The number of times operations on InnoDB tables had to wait for a row lock."),
			Type: dto.MetricType_COUNTER.Enum(),
		},
		"innodb_row_lock_current_waits": {
			Name: sp("innodb_row_lock_current_waits"),
			Help: sp("The number of row locks currently being waited for by operations on InnoDB tables."),
			Type: dto.MetricType_GAUGE.Enum(),
		},
		"table_locks_waited": {
			Name: sp("table_locks_waited"),
			Help: sp("The number of times that a request for a table lock could not be granted immediately and a wait was needed."),
			Type: dto.MetricType_COUNTER.Enum(),
		},
	}
}

func GatherMetrics(servers []string) ([]*dto.MetricFamily, error) {
	metrics := make([]MetricFamilyMap, 0)
	for _, server := range cfg.Servers {
		serverCfg, err := mysql.ParseDSN(server)
		if err != nil {
			return nil, fmt.Errorf("error parsing server dsn: %v", err)
		}
		if serverCfg.Timeout == 0 {
			serverCfg.Timeout = time.Second * 5
		}
		serverName := serverCfg.Addr
		serverLabel := &dto.LabelPair{Name: &serverNameLabel, Value: &serverName}
		db, err := sql.Open("mysql", serverCfg.FormatDSN())
		if err != nil {
			fmt.Printf("error opening connection: %v\n", err)
			return nil, fmt.Errorf("error opening connection: %v", err)
		}
		results, err := fromServerStatusVars(db)
		if err != nil {
			fmt.Printf("error collecting metrics for server: %s db:%s: %v\n", serverCfg.Addr, serverCfg.DBName, err)
			return nil, fmt.Errorf("error collecting metrics for server: %s db:%s: %v", serverCfg.Addr, serverCfg.DBName, err)
		}
		tagAll(results, []*dto.LabelPair{serverLabel})
		metrics = append(metrics, results)

	}
	metricFamilies := getFamilyDefinitions()

	for _, group := range metrics {
		for groupName, groupPoints := range group {
			family, ok := metricFamilies[groupName]
			if !ok {
				return nil, fmt.Errorf("unexpected error coalescing metrics. unspecified metric point %s", groupName)
			}
			family.Metric = append(family.Metric, groupPoints...)
		}
	}

	results := make([]*dto.MetricFamily, 0, len(metricFamilies))
	for _, family := range metricFamilies {
		results = append(results, family)
	}
	return results, nil
}

// Producer - documentation only type until sources other than mysql server status are implemented
type Producer func(*sql.DB) (MetricFamilyMap, error)

func fromServerStatusVars(db *sql.DB) (MetricFamilyMap, error) {
	rows, err := db.Query("SHOW GLOBAL STATUS;")
	if err != nil {
		return nil, fmt.Errorf("error getting server status: %v", err)
	}
	defer rows.Close()
	metrics := make(MetricFamilyMap)
	for rows.Next() {
		var key string
		var val sql.RawBytes
		if err = rows.Scan(&key, &val); err != nil {
			return nil, fmt.Errorf("error scanning server status results: %v", err)
		}
		switch rowKey := strings.ToLower(key); rowKey {
		case "uptime":
			fallthrough
		case "connection_errors_internal":
			fallthrough
		case "connection_errors_max_connections":
			fallthrough
		case "slow_queries":
			fallthrough
		case "queries":
			fallthrough
		case "innodb_data_fsyncs":
			fallthrough
		case "innodb_row_lock_waits":
			fallthrough
		case "table_locks_waited":
			i, err := strconv.ParseInt(string(val), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("could not read server status value as integer %s: %s", key, string(val))
			}
			fVal := float64(i)
			metrics[rowKey] = []*dto.Metric{{Counter: &dto.Counter{Value: &fVal}}}
		case "innodb_row_lock_current_waits":
			i, err := strconv.ParseInt(string(val), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("could not read server status value as integer %s: %s", key, string(val))
			}
			fVal := float64(i)
			metrics[rowKey] = []*dto.Metric{{Gauge: &dto.Gauge{Value: &fVal}}}
		default:
			// ignore
		}
	}
	return metrics, nil
}

func tagAll(families map[string][]*dto.Metric, tags []*dto.LabelPair) {
	for _, family := range families {
		for _, metric := range family {
			metric.Label = append(metric.Label, tags...)
		}
	}
}

// MetricFamilyMap is a partial representation of a prometheus metric family
// a map from metric family name -> list of metric observations
type MetricFamilyMap map[string][]*dto.Metric
