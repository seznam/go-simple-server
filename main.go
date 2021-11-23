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

	sloClasses = []string{
		"high",
		"low",
	}
	defaultSloDomain = "go-simple-server-domain"
	sloApp           = "go-simple-server"
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
		log.Infof("%s %v %d %s slo-class=%s slo-result=%s", r.Method, r.RequestURI, o.statusCode, duration, w.Header().Get("slo-class"), w.Header().Get("slo-result"))
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "OK")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// add slo domain and slo app header
	sloDomain := os.Getenv("SLO_DOMAIN")
	if sloDomain == "" {
		sloDomain = defaultSloDomain
	}
	w.Header().Add("slo-app", sloApp)
	w.Header().Add("slo-domain", sloDomain)

	// sleep for random duration
	wait := rand.Int63n(1000)
	time.Sleep(time.Duration(wait) * time.Millisecond)

	// pick random class
	class := sloClasses[rand.Intn(len(sloClasses))]
	w.Header().Add("slo-class", class)

	// return random status code:
	// - 200 - 50%
	// - 404 - 20%
	// - 500 - 30%
	result := rand.Intn(10)
	if result < 5 {
		w.Header().Add("slo-result", "ok")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "OK")
	} else if result < 7 {
		w.Header().Add("slo-result", "ok")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "Not found")
	} else {
		w.Header().Add("slo-result", "fail")
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
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, PadLevelText: true})
	log.SetOutput(os.Stdout)

	prometheus.MustRegister(requestDuration)
	rand.Seed(time.Now().Unix())
}
