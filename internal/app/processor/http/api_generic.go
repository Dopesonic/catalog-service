package rprocessor

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	rhandler "github.com/Dopesonic/catalog-service/internal/app/handler/http"
)

func vGenericRegHealthCheck(r *mux.Router, h rhandler.Health) {
	reg(r, http.MethodGet, "/health", http.HandlerFunc(h.LastCheck))
	reg(r, http.MethodGet, "/metrics", promhttp.Handler())
}

func vGenericRegProf(r *mux.Router) {
	r.HandleFunc("/debug/pprof", pprofIndex)
	r.HandleFunc("/debug/pprof/", pprof.Index)

	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	for _, name := range []string{"allocs", "block", "goroutine", "heap", "mutex", "threadcreate"} {
		r.Handle("/debug/pprof/"+name, pprof.Handler(name))
	}
}

func pprofIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
}

func handlerNotFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
