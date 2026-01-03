package wol

import (
	"net/http"

	"github.com/jonathongardner/fife/logger"
	"github.com/sirupsen/logrus"
)

func Middleware(info *Info, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if info != nil {
			if info.ShouldWakeUp() {
				logger.ContextLogger(r.Context()).WithFields(logrus.Fields{
					"error": info.WakeUp(),
				}).Info("Sent WOL packet.")
			}
		}
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
