package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jonathongardner/fife/app"
	"github.com/jonathongardner/fife/logger"
	"github.com/jonathongardner/fife/wol"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

//go:embed static
var staticFiles embed.FS

func NewReverseProxy(ctx context.Context, cfg config) error {
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("couldnt create static fs: %w", err)
	}

	server := &http.Server{
		Addr:    cfg.BindHost,
		Handler: reverseMuxProxy(cfg, subFS),
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		// TODO: handle server closed and ignore?
		if err := server.ListenAndServe(); err != nil {
			return fmt.Errorf("error in listen %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done() // Block until the context is done
		logrus.Infof("Initiating server shutdown...")

		// Create a shutdown context with a timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("error shutingdown server %w", err)
		}
		logrus.Infof("Server shutdown...")
		return nil
	})

	return g.Wait()
}

func reverseMuxProxy(cfg config, subFS fs.FS) *mux.Router {
	logrus.Infof("Creating server with %d proxies", len(cfg.Services))
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	// Create an info route
	if cfg.InfoHost != "" {
		logrus.WithFields(
			logrus.Fields{
				"route": cfg.InfoHost,
			},
		).Info("Adding info route")

		r.Host(cfg.InfoHost).Path("/api/v1/version").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(app.Version)); err != nil {
				http.Error(w, "Error writing response", http.StatusInternalServerError)
				return
			}

		})

		r.Host(cfg.InfoHost).Path("/api/v1/services").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(cfg); err != nil {
				http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
				return
			}

		})

		r.Host(cfg.InfoHost).Path("/api/v1/services/{id}/wol").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			for _, s := range cfg.Services {
				if s.id == id && s.wolInfo != nil {
					err := s.wolInfo.WakeUp(r.Context())
					if err != nil {
						http.Error(w, "Error waking up", http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					if err := json.NewEncoder(w).Encode(s); err != nil {
						http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
						return
					}
					return
				}
			}

			http.Error(w, "Service not found", http.StatusNotFound)
		})
		r.Host(cfg.InfoHost).Handler(http.FileServer(http.FS(subFS)))
	}

	// Create proxy route
	for _, s := range cfg.Services {
		logrus.WithFields(
			logrus.Fields{
				"route":    s.proxyOn.host,
				"proxy-to": s.proxyToUrl.String(),
			},
		).Info("Added proxy route")

		r.Host(s.proxyOn.host).PathPrefix("/").Handler(wol.Middleware(s.wolInfo, s.ProxyTo()))
		// http.StripPrefix(path, s.ProxyTo()
	}

	// Fallback function
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logger.ContextLogger(r.Context())
		logger.Infof("unknown route")

		w.WriteHeader(418)
		_, err := w.Write([]byte("Im a teapot... sorry"))
		if err != nil {
			logger.WithError(err).Warn("couldnt write response")
		}
	})

	return r
}

// logger := logrus.WithFields(logrus.Fields{
// 	"originalHost": r.Host,
// 	"host":         host,
// 	"path":         r.URL.Path,
// 	"remoteAdd":    r.RemoteAddr,
// })
// logger.Info("request")
