package middlewares

import (
	"net/http"

	"github.com/squarefactory/cloud-burster/logger"
	"go.uber.org/zap"
)

type RoundTripper struct {
	http.RoundTripper
}

func (lrt *RoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	logger.I.Debug("http",
		zap.String("method", request.Method),
		zap.String("uri", request.RequestURI),
		zap.String("proto", request.Proto),
		zap.String("remote_addr", request.RemoteAddr),
		zap.String("host", request.Host),
		zap.Any("headers", request.Header),
	)

	return lrt.RoundTripper.RoundTrip(request)
}
