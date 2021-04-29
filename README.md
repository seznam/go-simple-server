# GO simple server

Simple go server with healthcheck, metrics and logging listening on port `:8080`

## Build
Requires Go version 1.16.

```
go build
```

## Run
```
$ ./go-simple-server
INFO[2021-04-29T12:01:25+02:00] Server started
```

## Endpoints
* `/` - returns `200 OK` or `500 internal server error`. There is 2:1 chance to get `200`.
* `/liveness` - used for kubernetes liveness check - returns `200 OK`
* `/readinsss` - used for kubernetes readinsss check - returns `200 OK`
* `/metrics` - used for prometheus metrics scraping
