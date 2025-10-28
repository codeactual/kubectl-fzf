package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"

	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
)

type FzfHttpServer struct {
	Port        int
	ResourceHit int
	//LastModifiedHit int

	stores      []*store.Store
	storeConfig *store.StoreConfig
}

type routeResourceFunc func(http.ResponseWriter, *http.Request, resources.ResourceType)

func curryResourceRoute(f routeResourceFunc, resourceType resources.ResourceType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r, resourceType)
	}
}

func (f *FzfHttpServer) readinessRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Ok"))
	case http.MethodHead:
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (f *FzfHttpServer) statsRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	stats := store.GetStatsFromStores(f.stores)
	log.Debugf("Sending stats: %v", stats)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Errorf("unable to encode stats response: %v", err)
	}
}

func (f *FzfHttpServer) resourcesRoute(w http.ResponseWriter, r *http.Request, resourceType resources.ResourceType) {
	switch r.Method {
	case http.MethodGet:
		f.ResourceHit++
	case http.MethodHead:
	// handled below without incrementing hits
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if resourceType == resources.ResourceTypeUnknown {
		http.Error(w, "Resource type unknown", http.StatusBadRequest)
		return
	}
	if !f.storeConfig.FileStoreExists(resourceType) {
		http.Error(w, fmt.Sprintf("resource file for %s not found", resourceType), http.StatusNotFound)
		return
	}
	filePath := f.storeConfig.GetResourceStorePath(resourceType)
	log.Debugf("Serving file %s", filePath)
	http.ServeFile(w, r, filePath)
}

func (f *FzfHttpServer) setupRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/readiness", f.readinessRoute)
	mux.HandleFunc("/stats", f.statsRoute)

	for r := resources.ResourceTypeApiResource; r < resources.ResourceTypeUnknown; r++ {
		path := fmt.Sprintf("/k8s/resources/%s", r.String())
		mux.HandleFunc(path, curryResourceRoute(f.resourcesRoute, r))
	}

	skipLogs := map[string]struct{}{
		"/health": {},
	}
	return recoveryMiddleware(loggingMiddleware(skipLogs, mux))
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func loggingMiddleware(skipPaths map[string]struct{}, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := newLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)
		if _, ok := skipPaths[r.URL.Path]; ok {
			return
		}
		duration := time.Since(start)
		log.Infof("%s %s %d %s %dB", r.Method, r.URL.Path, lrw.status, duration, lrw.bytes)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Errorf("panic recovered while handling %s %s: %v", r.Method, r.URL.Path, rec)
				log.Debugf("stacktrace: %s", string(debug.Stack()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func startHttpServer(ctx context.Context, listener net.Listener, srv *http.Server) {
	go func() {
		log.Infof("Starting http server on %s", srv.Addr)
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error listening: %s", err)
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}
	log.Info("Exiting http server")
}

func StartHttpServer(ctx context.Context, h *HttpServerConfigCli, storeConfig *store.StoreConfig, stores []*store.Store) (*FzfHttpServer, error) {
	if h.ListenAddress == "" {
		return nil, nil
	}
	listener, err := net.Listen("tcp", h.ListenAddress)
	if err != nil {
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	f := FzfHttpServer{
		Port:        port,
		stores:      stores,
		storeConfig: storeConfig,
	}
	router := f.setupRouter()
	srv := &http.Server{
		Addr:    h.ListenAddress,
		Handler: router,
	}
	go startHttpServer(ctx, listener, srv)
	return &f, nil
}
