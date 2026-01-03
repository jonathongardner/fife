package server

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jonathongardner/fife/logger"
	log "github.com/sirupsen/logrus"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := log.WithFields(log.Fields{
			"host":      r.Host,
			"path":      r.URL.Path,
			"method":    r.Method,
			"requestId": uuid.New(),
		})

		lgr.Info("Request")
		rww := &responseWriterWrapper{ResponseWriter: w}
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rww, r.WithContext(logger.NewContextLogger(r.Context(), lgr)))
		lgr.WithField("status", rww.statusCode).Info("Response")
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	Error      error
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
