package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	addr     string
	registry *prometheus.Registry
}

func NewServer(addr string, registry *prometheus.Registry) *Server {
	return &Server{
		addr:     addr,
		registry: registry,
	}
}

func (s *Server) Start() error {
	http.Handle("/metrics", promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}))
	return http.ListenAndServe(s.addr, nil)
}
