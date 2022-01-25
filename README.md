# mysql-check
Sensu check/metrics collector for MySQL


## Usage

### Help Text Output

```
Usage:
  mysql-check [flags]
  mysql-check [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -h, --help                   help for mysql-check
      --insecure-skip-verify   If true, use TLS but skip chain & host verification for custom TLS config
  -s, --servers strings        A list of one or more server connection URLs in DNS Format
      --tls-ca string          Path to a ca.pem file for custom TLS config
      --tls-cert string        Path to a cert.pem file for custom TLS config
      --tls-key string         Path to a key.pem file for custom TLS config

Use "mysql-check [command] --help" for more information about a command.
```

### Configuration

The servers argument is a comma delimited list of server connection strings in [DSN Format][1]. Example: `sensu@tcp(localhost:3306)/mysql?timeout=3s`

In order to take advantage of advanced TLS configuration provided by the `--tls-*` configuration flags, set the `tls` parameter in the server connection string to `custom`.

Example: `mysql-check --servers "sensu@tcp(10.0.0.10:3306)/mydb?tls=custom" --tls-ca=/opt/tls/ca.pem --tls-cert=/opt/tls/client.pem --tls-key=/opt/tls/client-key.pem`

## Test Suite

The current test suite depends on a running mysql instance listening for tcp connections locally on the default port.

Use the docker-compose.yml stack to stand this up.

`docker-compose up -d`

`go test ./...`

[1]: https://github.com/go-sql-driver/mysql#dsn-data-source-name

