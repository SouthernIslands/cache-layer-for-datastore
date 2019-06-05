package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type statusHandler struct {
	*Server
}

func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("Map now is", h.GetMap())

	b, e := json.Marshal(h.GetMap())
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c, e := json.Marshal(map[string][]byte{"k1": {'v', '1'}, "k2": {'v', '2'}})
	d := append(c, b...)
	w.Write(d)
}

func (s *Server) statusHandler() http.Handler {
	return &statusHandler{s}
}
