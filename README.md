# Exchange rate API

`GET /rate` - get USD to UAH exchange rate

`POST /subscribe` - subscribe to exchange rate update (send application/x-www-form-urlencoded email address)

`POST /unsubscribe` - delete exchange rate subscription (send application/x-www-form-urlencoded email address)


## Running application

### Local
Create set of local configuration files (use `config/*.pruduction.yml` as an example) with development values. Then create `docker-compose.local.yml` to add configation specific directives to core `docker-compose.yml` file.

`docker compose -f docker-compose.yml -f docker-compose.local.yml --build`

### Production
Edit set of configuration files (`config/*.pruduction.yml`), put there your SMTP credentials and so on. 
Edit `docker-compose.production.yml` to change environment variables values for DB and RabbitMQ.

`docker compose -f docker-compose.yml -f docker-compose.production.yml --build`

## Metrics

Application (each service at :8080/metrics in Prometheus format) exposes different metrics such as:

- total_email_sent
- customers_created_total{success=true|false}
- requests_total{method, path, status}
- total_subscribers{success=true|false}
- total_unsubscribers{success=true|false}

And other go_* and process_* metrics

To create metrics dashboards and view graphical representation go to localhost:3000 (Grafana)

## Alerts

- total_subscribers{success=true} != customers_created_total{success=true} (Error in interservice communication, SAGA does not work)
- Anomaly rising of total_unsubscribers metric
- go_max_fd - go_total_fd < 100 (Small amount of open free file descriptors, connections leaks)
- total_email_sent < total_subscribers - total_unsubsribers (Not all emails were sent)


## Architecture

![alt text](https://raw.githubusercontent.com/GenesisEducationKyiv/software-engineering-school-4-0-fdemchenko/568a67efbfa5e8ab819cf4f53e3599ed348f7792/docs/architecture.png)

## Tests

To run unit tests for whole applicaton run `go test -short ./...`
Run `go test ./...` to run all tests, including integration ones (require docker installed on your system)

## TODO

- [x] Graceful server shutdown
- [ ] SSL encryption
- [x] Parallel emails sending




