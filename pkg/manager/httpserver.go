package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/rancher/support-bundle-kit/pkg/utils"
)

type HttpServer struct {
	context context.Context
	manager *SupportBundleManager
}

func (s *HttpServer) getStatus(w http.ResponseWriter, req *http.Request) {
	s.manager.status.RLock()
	defer s.manager.status.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(s.manager.status.ManagerStatus)
	if err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, err)
		return
	}
}

func (s *HttpServer) getBundle(w http.ResponseWriter, req *http.Request) {
	bundleFile := s.manager.getBundlefile()
	f, err := os.Open(bundleFile)
	if err != nil {
		utils.HttpResponseError(w, http.StatusNotFound, fmt.Errorf("fail to open bundle file: %v", err))
		return
	}
	defer func() {
		_ = f.Close()
	}()

	fstat, err := f.Stat()
	if err != nil {
		utils.HttpResponseError(w, http.StatusNotFound, fmt.Errorf("fail to stat bundle file: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Length", strconv.FormatInt(fstat.Size(), 10))
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(bundleFile))
	if _, err := io.Copy(w, f); err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, err)
		return
	}
}

func (s *HttpServer) createNodeBundle(w http.ResponseWriter, req *http.Request) {
	node := mux.Vars(req)["nodeName"]
	if node == "" {
		utils.HttpResponseError(w, http.StatusBadRequest, errors.New("empty node name"))
		return
	}

	logrus.Debugf("Handle create node bundle for %s", node)
	nodesDir := filepath.Join(s.manager.getWorkingDir(), "nodes")
	err := os.MkdirAll(nodesDir, os.FileMode(0775))
	if err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, fmt.Errorf("fail to create directory %s: %v", nodesDir, err))
		return
	}

	nodeBundle := filepath.Join(nodesDir, node+".zip")
	f, err := os.Create(nodeBundle)
	if err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, fmt.Errorf("fail to create file %s: %s", nodeBundle, err))
		return
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = io.Copy(f, req.Body)
	if err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, err)
		return
	}

	err = s.manager.verifyNodeBundle(nodeBundle)
	if err != nil {
		utils.HttpResponseError(w, http.StatusBadRequest, fmt.Errorf("fail to verify file %s: %v", nodeBundle, err))
		return
	}
	s.manager.completeNode(node)
	utils.HttpResponseStatus(w, http.StatusCreated)
}

func (s *HttpServer) Run(m *SupportBundleManager) {
	defaultTimeout := 24 * time.Hour

	r := mux.NewRouter()
	r.UseEncodedPath()

	r.Path("/status").Methods("GET").HandlerFunc(s.getStatus)
	r.Path("/bundle").Methods("GET").HandlerFunc(s.getBundle)
	r.Path("/nodes/{nodeName}").Methods("POST").HandlerFunc(s.createNodeBundle)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    defaultTimeout,
		WriteTimeout:   defaultTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	_ = server.ListenAndServe()
}
