package wol

import (
	"net/http"
)

func Middleware(info *Info, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if info != nil {
			// ignore error cause its logged and it wol didnt work server wont be awake anyways
			// so user will get error than
			info.WakeUp(r.Context())
		}
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
