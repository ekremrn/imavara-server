package observe

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

var counters = struct {
	sync.Mutex
	values map[string]float64
}{values: map[string]float64{}}

// Metric names and labels come only from bounded server-side constants.
func Add(metric string, amount float64) {
	counters.Lock()
	counters.values["imavara_"+metric] += amount
	counters.Unlock()
}
func Start(service string) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", service))
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{Dsn: dsn, ServerName: service, SendDefaultPII: false}); err != nil {
			slog.Warn("Sentry initialization failed")
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintln(w, "imavara_service_ready 1")
		counters.Lock()
		defer counters.Unlock()
		for k, v := range counters.values {
			fmt.Fprintf(w, "%s %g\n", k, v)
		}
	})
	go func() {
		if err := http.ListenAndServe(":9091", mux); err != nil {
			slog.Error("metrics endpoint stopped", "error", err)
		}
	}()
}
func Error(err error) {
	if err != nil && os.Getenv("SENTRY_DSN") != "" {
		sentry.CaptureException(err)
	}
}
func Duration(metric string, start time.Time) {
	Add(metric+"_seconds_sum", time.Since(start).Seconds())
	Add(metric+"_seconds_count", 1)
}
