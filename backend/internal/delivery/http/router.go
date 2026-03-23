// Package delivery/http: wire-up tất cả dependencies theo Clean Architecture.
// Đây là "Composition Root" – nơi duy nhất tạo concrete implementations và inject vào.
package delivery

import (
	"fmt"
	"net/http"
	"os"

	"fusion/internal/platform/utils"
	"fusion/internal/repository"
	"fusion/internal/usecase"
)

// Router đăng ký tất cả HTTP routes theo Clean Architecture.
// Được gọi từ server.go thay vì http.HandleFunc toàn cục.
type Router struct {
	mux     *http.ServeMux
	handler *DashboardHandler
}

// NewRouter khởi tạo Router với Dependency Injection đầy đủ.
// Đây là nơi duy nhất các concrete repository được tạo ra.
func NewRouter() *Router {
	// ── Repository layer (Infrastructure) ─────────────────────────────────────
	vmRepo := repository.NewVMDashboardRepo()        // talks to VictoriaMetrics
	entityRepo := repository.NewPostgresEntityRepo() // talks to PostgreSQL

	// ── Usecase layer (Business Logic) ────────────────────────────────────────
	dashUC := usecase.NewDashboardUsecase(vmRepo)

	// ── Delivery layer (HTTP Adapters) ────────────────────────────────────────
	dashHandler := NewDashboardHandler(dashUC, entityRepo)

	mux := http.NewServeMux()
	r := &Router{mux: mux, handler: dashHandler}
	r.registerRoutes()
	return r
}

// registerRoutes maps URL patterns → handlers.
func (r *Router) registerRoutes() {
	// Auth
	r.mux.HandleFunc("/api/auth/login", HandleAuthLogin)
	r.mux.HandleFunc("/api/auth/refresh", HandleAuthRefresh)
	r.mux.HandleFunc("/api/auth/change-password", HandleAuthChangePassword)

	// Dashboard & Streaming
	r.mux.HandleFunc("/api/production-monthly", r.handler.HandleMonthlyProduction)
	r.mux.HandleFunc("/api/inverter/dc-power", r.handler.HandleInverterDCPower)
	r.mux.HandleFunc("/api/rename", r.handler.HandleRename)

	// Utilities (stateless, no DI needed)
	r.mux.HandleFunc("/healthz", HandleHealthz)
}

// ServeHTTP delegates to the internal mux.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// StartCleanServer starts a parallel HTTP server on port 5040,
// exposing only the Clean Architecture routes.
// Legacy routes on port 5039 (server.go) remain untouched.
// This allows gradual migration without breaking existing integrations.
func StartCleanServer() {
	router := NewRouter()
	port := ":5040"
	fmt.Printf("[CA] Clean Architecture Server listening on %s\n", port)
	if err := http.ListenAndServe(port, corsHandler(router)); err != nil {
		utils.LogError("[CA] Server failed: %v", err)
		os.Exit(1)
	}
}

// corsHandler wraps a handler with CORS headers.
func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
