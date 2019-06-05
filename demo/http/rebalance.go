package http

import (
	"bytes"
	"log"
	"net/http"
)

type rebalanceHandler struct {
	*Server
}

func (h *rebalanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	go h.rebalance()
}

func (h *rebalanceHandler) rebalance() {
	s := h.NewScanner()
	defer s.Close()
	c := &http.Client{}
	for s.Scan() {
		k := s.Key()
		addr, e := h.ShouldProcess(k)
		if !e {
			r, _ := http.NewRequest(http.MethodPut, "http://"+addr+":14350/cache/"+k,
				bytes.NewBuffer(s.Value()))
			resp, err := c.Do(r)
			log.Println(resp, err)
			h.Del(k)
		}
	}
}

func (s *Server) rebalanceHandler() http.Handler {
	return &rebalanceHandler{s}
}
