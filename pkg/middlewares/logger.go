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
		zap.String("host", request.Host),
		zap.String("path", request.URL.Path),
		zap.String("proto", request.Proto),
		zap.String("remote_addr", request.RemoteAddr),
		zap.Any("headers", request.Header),
	)

	return lrt.RoundTripper.RoundTrip(request)
}
