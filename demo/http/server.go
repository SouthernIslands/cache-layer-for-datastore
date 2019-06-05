package http

import (
	"demo/cache"
	"demo/cluster"
	"net/http"
)

type Server struct {
	cache.Cache
	cluster.Node
}

func (s *Server) Listen() {
	http.Handle("/cache/", s.cacheHandler())
	http.Handle("/status", s.statusHandler())
	http.Handle("/cluster", s.clusterHandler())
	http.Handle("/rebalance", s.rebalanceHandler())
	http.ListenAndServe(":14350", nil)
}

func New(c cache.Cache, n cluster.Node) *Server {
	return &Server{c, n}
}
