package util

import (
	"bytes"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func LogResponseBody(response *http.Response) {
	response.Body = logBody(response.Body)
}

func LogRequestBody(request *http.Request) {
	request.Body = logBody(request.Body)
}

func logBody(rd io.ReadCloser) io.ReadCloser {
	bodyb, _ := io.ReadAll(rd)
	log.Infof("%s", string(bodyb))
	return io.NopCloser(bytes.NewBuffer(bodyb))
}
