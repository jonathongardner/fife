package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func NewReverseProxy(ctx context.Context, cfg config) error {
	server := &http.Server{
		Addr:    cfg.BindHost,
		Handler: reverseMuxProxy(cfg),
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

func reverseMuxProxy(cfg config) *http.ServeMux {
	var defProx *httputil.ReverseProxy
	// Define the target backend server URL
	proxies := make(map[string]*httputil.ReverseProxy)
	for _, svr := range cfg.Services {
		logrus.WithFields(
			logrus.Fields{
				"name": svr.name,
				"host": svr.host.String(),
			},
		).Info("Added proxy")
		proxies[svr.name] = httputil.NewSingleHostReverseProxy(svr.host)
		if svr.deflt {
			defProx = proxies[svr.name]
		}
	}

	// Create a new ReverseProxy instance
	logrus.Infof("Creating server with %d proxies", len(proxies))
	mux := http.NewServeMux()
	// Define the handler function for the reverse proxy
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := strings.SplitN(r.Host, ":", 2)[0]
		logrus.WithFields(logrus.Fields{
			"originalHost": r.Host,
			"host":         host,
			"path":         r.URL.Path,
			"remoteAdd":    r.RemoteAddr,
		}).Info("request")

		proxy, ok := proxies[host]
		if !ok {
			proxy = defProx
		}

		proxy.ServeHTTP(w, r)
	})
	return mux
}
