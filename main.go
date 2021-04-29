package main

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "Time (in seconds) spent serving HTTP requests.",
		Buckets: prometheus.DefBuckets,
	}, []string{"app", "method", "endpoint", "status_code"})
)

type responseObserver struct {
	http.ResponseWriter
	statusCode int
}

func (o *responseObserver) Write(p []byte) (n int, err error) {
	if o.statusCode == 0 {
		o.statusCode = http.StatusOK
	}

	n, err = o.ResponseWriter.Write(p)
	return n, err
}

func (o *responseObserver) WriteHeader(status int) {
	o.statusCode = status
	o.ResponseWriter.WriteHeader(status)
}

func middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := &responseObserver{ResponseWriter: w}

		start := time.Now()
		handler.ServeHTTP(o, r)
		duration := time.Since(start)

		requestDuration.WithLabelValues("go-simple-server", r.Method, r.RequestURI, strconv.Itoa(o.statusCode)).Observe(float64(duration.Seconds()))
		log.Infof("%s %v %d %s", r.Method, r.RequestURI, o.statusCode, duration)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "OK")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	result := rand.Intn(3)
	if result != 0 {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "OK")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "Internal server error")
	}
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/liveness", healthHandler)
	router.HandleFunc("/readiness", healthHandler)
	router.HandleFunc("/", indexHandler)
	router.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))

	log.Info("Server started")
	http.ListenAndServe(":8080", middleware(router))
}

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)

	prometheus.MustRegister(requestDuration)
}
