package internal

import (
	"bytes"
	"net/http"
	"net/url"
)

type CanonicalRequest struct{}

func (CanonicalRequest) Create(endpoint *url.URL, payload []byte) (*http.Request, error) {

	req, err := http.NewRequest(
		"POST",
		endpoint.String(),
		bytes.NewReader(payload),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json, text/javascript")
	req.Header.Set("content-encoding", "amz-1.0")
	req.Header.Set("content-type", "application/json; charset=UTF-8")
	req.Header.Set("host", endpoint.Hostname())

	req.ContentLength = -1
	req.Header.Del("Content-Length")
	req.Header.Del("content-length")

	return req, nil
}
