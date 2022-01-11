# mysql-check
Sensu check/metrics collector for MySQL


## Test Suite

The current test suite depends on a running mysql instance listening for tcp connections locally on the default port.

Use the docker-compose.yml stack to stand this up.

`docker-compose up -d`

`go test ./...`

