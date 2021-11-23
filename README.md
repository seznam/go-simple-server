# Exercise for a candidate to SRE role 

The purpose of this test scenario is to find out if the candidate is able to quickly study new
technologies and apply them in practice. Even though most of the configuration snippets can be found around the Internet, we also expect the candidate to understand how the infrastructure works and is able to explain choices he made during the implementation.

Discussion will take place on chosen solution and used technologies after task is completed.


## Assignment

1. Start an instance of Kubernetes (use  minikube or other similar project, don't use cloud provider's managed kubernetes).
2. Run the `go-simple-server` service (built from this repo) in the Kubernetes instance.
   * create Dockerfile and build container
   * push container image into some public registry (e.g. https://hub.docker.com/)
   * create kubernetes manifests and deploy them into your k8s cluster - Service rolling update must not lead to an outage.
3. Run Prometheus and Grafana in the same Kubernetes instance observing the running service.
4. Create a Grafana dashboard with information about the state of the `go-simple-server` service. It should show at least:
   1. Number of requests served by the service grouped by a status code (as a graph).
   2. Histogram of a response latency.



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
* `/` - returns `200 OK`, `404 Not Found`, `500 internal server error`. There is 5:2:3 chances to get 200, 404, 500 status code. Also there is random latency (0-1000 ms)
* `/liveness` - used for kubernetes liveness check - returns `200 OK`
* `/readiness` - used for kubernetes readinsss check - returns `200 OK`
* `/metrics` - used for prometheus metrics scraping
