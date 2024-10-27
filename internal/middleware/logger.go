package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var Log *zap.SugaredLogger

func Initialize() {
	zl, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize zap logger")
	}
	Log = zl.Sugar()
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseStatus int
	responseSize   int
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(data)
	lrw.responseSize += size
	return size, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.ResponseWriter.WriteHeader(statusCode)
	lrw.responseStatus = statusCode
}

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Log.Infow(
			"request",
			"uri", r.RequestURI,
			"method", r.Method,
		)

		lrw := loggingResponseWriter{ResponseWriter: w}

		start := time.Now()
		h.ServeHTTP(&lrw, r) // Вызов метода ServeHTTP для хендлера
		duration := time.Since(start)

		Log.Infow(
			"response",
			"size", lrw.responseSize,
			"status", lrw.responseStatus,
			"duration", duration,
		)
	})
}
