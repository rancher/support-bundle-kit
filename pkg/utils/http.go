package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	// Errors happened during request.
	Errors []string `json:"errors,omitempty"`
}

func httpResponseBody(obj interface{}) []byte {
	respBody, err := json.Marshal(obj)
	if err != nil {
		return []byte(`{\"errors\":[\"Failed to parse response body\"]}`)
	}
	return respBody
}

func HttpResponseOKWithBody(rw http.ResponseWriter, obj interface{}) {
	rw.Header().Set("Content-type", "application/json")
	_, _ = rw.Write(httpResponseBody(obj))
}

func HttpResponseStatus(rw http.ResponseWriter, statusCode int) {
	rw.WriteHeader(statusCode)
}

func HttpResponseError(rw http.ResponseWriter, statusCode int, err error) {
	logrus.Error(err)
	HttpResponseErrorMsg(rw, statusCode, err.Error())
}

func HttpResponseErrorMsg(rw http.ResponseWriter, statusCode int, errMsg string) {
	rw.WriteHeader(statusCode)
	_, _ = rw.Write(httpResponseBody(ErrorResponse{Errors: []string{errMsg}}))
}

// HttpGetDispositionFilename parses value of "Content-Disposition" header
// e.g., extract "abc.zip" from "attachment; filename=abc.zip"
func HttpGetDispositionFilename(disposition string) (string, error) {
	err := fmt.Errorf("unexpected disposition value: %s", disposition)

	if disposition == "" {
		return "", err
	}

	var attachement bool
	var filename string

	for _, param := range strings.Split(disposition, ";") {
		p := strings.TrimSpace(param)
		if p == "attachment" {
			attachement = true
		}

		if strings.HasPrefix(p, "filename=") {
			filename = strings.Trim(strings.SplitN(p, "filename=", 2)[1], "\"")
		}
	}
	if attachement && filename != "" {
		return filename, nil
	}
	return "", err
}
