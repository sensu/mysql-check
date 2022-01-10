package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/prometheus/common/expfmt"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	CustomTLS
	Servers         []string
	CustomTLSConfig *tls.Config
}

var (
	cfg = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "mysql-check",
			Short:    "Check for producing observability metrics from mysql databases",
			Keyspace: "sensu.io/plugins/mysql-check/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "servers",
			Argument:  "servers",
			Shorthand: "s",
			Env:       "SERVERS",
			Usage:     "A list of one or more server connection URLs in DNS Format",
			Value:     &cfg.Servers,
		}, {
			Path:     "tls-ca",
			Argument: "tls-ca",
			Env:      "TLS_CA",
			Usage:    "Path to a ca.pem file for custom TLS config",
			Value:    &cfg.TLSCA,
		}, {
			Path:     "tls-cert",
			Argument: "tls-cert",
			Env:      "TLS_CERT",
			Usage:    "Path to a cert.pem file for custom TLS config",
			Value:    &cfg.TLSCert,
		}, {
			Path:     "tls-key",
			Argument: "tls-key",
			Env:      "TLS_KEY",
			Usage:    "Path to a key.pem file for custom TLS config",
			Value:    &cfg.TLSKey,
		}, {
			Path:     "insecure-skip-verify",
			Argument: "insecure-skip-verify",
			Env:      "INSECURE_SKIP_VERIFY",
			Usage:    "If true, use TLS but skip chain & host verification for custom TLS config",
			Value:    &cfg.InsecureSkipVerify,
		},
	}
)

func main() {
	useStdin := false
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Printf("Error check stdin: %v\n", err)
		panic(err)
	}
	//Check the Mode bitmask for Named Pipe to indicate stdin is connected
	if fi.Mode()&os.ModeNamedPipe != 0 {
		log.Println("using stdin")
		useStdin = true
	}

	check := sensu.NewGoCheck(&cfg.PluginConfig, options, checkArgs, executeCheck, useStdin)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	if len(cfg.Servers) == 0 {
		return sensu.CheckStateCritical, fmt.Errorf("expected at least one server")
	}
	custom, err := cfg.TLSConfig()
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("invalid TLS Config: %v", err)
	}
	cfg.CustomTLSConfig = custom
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	if cfg.CustomTLSConfig != nil {
		_ = mysql.RegisterTLSConfig("custom", cfg.CustomTLSConfig)
	}

	metricFamilies, err := GatherMetrics(cfg.Servers)
	if err != nil {
		fmt.Println(err.Error())
		return sensu.CheckStateCritical, nil
	}

	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, family := range metricFamilies {
		if err := enc.Encode(family); err != nil {
			fmt.Printf("failed to encode metrics to prometheus exposition format: %v\n", err)
			return sensu.CheckStateCritical, nil
		}
	}
	fmt.Print(buf.String())

	return sensu.CheckStateOK, nil
}
