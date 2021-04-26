package manager

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type HttpServer struct {
	context context.Context
}

func (s *HttpServer) Run(m *SupportBundleManager) {
	defaultTimeout := 30 * time.Second

	r := mux.NewRouter()
	r.UseEncodedPath()

	r.Path("/bundle").Methods("GET").HandlerFunc(m.getBundle)
	r.Path("/nodes/{nodeName}").Methods("POST").HandlerFunc(m.createNodeBundle)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    defaultTimeout,
		WriteTimeout:   defaultTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	_ = server.ListenAndServe()
}
