# Exchange rate API

`GET /rate` - get USD to UAH exchange rate

`POST /subscribe` - subscribe to exchange rate update (send application/x-www-form-urlencoded email address)


## Running application

Edit `docker-compose.yml` file and put your SMTP credentials, to be able to send updates to subscribers

To start http server and PostgreSQL service run: - `docker compose up`

## Tests

To run unit tests for whole applicaton run `go test -short ./...`
Run `go test ./...` to run all tests, including integration ones (require docker installed on your system)

## TODO

- [ ] Graceful server shutdown
- [ ] SSL encryption
- [ ] Parallel emails sending




