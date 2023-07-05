package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// provides a wrapper around zap.Logger that can be used as a http middleware

type loggerMiddleware struct {
	l *zap.Logger
}

// NewLoggerMiddleware instantiates a middleware function that logs all requests
// using the provided logger
func NewLoggerMiddleware(l *zap.Logger) func(next http.Handler) http.Handler {
	return loggerMiddleware{l.Named("middleware")}.Handler
}

func (z loggerMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		latency := time.Since(start)

		var requestID string
		if reqID := r.Context().Value(middleware.RequestIDKey); reqID != nil {
			requestID = reqID.(string)
		}
		z.l.Info("request completed",
			// request metadata
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("method", r.Method),
			zap.String("user-agent", r.UserAgent()),

			// response metadata
			zap.Int("status", ww.Status()),
			zap.Duration("took", latency),

			// additional metadata
			zap.String("real-ip", r.RemoteAddr),
			zap.String("request-id", requestID))
	},
	)
}
