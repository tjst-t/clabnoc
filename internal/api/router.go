package api

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tjst-t/clabnoc/internal/docker"
	"github.com/tjst-t/clabnoc/internal/frontend"
	"github.com/tjst-t/clabnoc/internal/network"
)

// Server holds the HTTP router and dependencies
type Server struct {
	router       *chi.Mux
	discoverer   *docker.Discoverer
	docker       docker.DockerClient
	faultManager *network.FaultManager
	faultState   *network.FaultState
}

// NewServer creates a new API server
func NewServer(dockerClient docker.DockerClient) *Server {
	s := &Server{
		router:       chi.NewRouter(),
		discoverer:   docker.NewDiscoverer(dockerClient),
		docker:       dockerClient,
		faultManager: network.NewFaultManager(),
		faultState:   network.NewFaultState(),
	}
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(corsMiddleware)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/projects", s.handleGetProjects)
		r.Get("/projects/{name}/topology", s.handleGetTopology)
		r.Get("/projects/{name}/nodes", s.handleGetNodes)
		r.Get("/projects/{name}/nodes/{node}", s.handleGetNode)
		r.Post("/projects/{name}/nodes/{node}/action", s.handleNodeAction)
		r.Get("/projects/{name}/nodes/{node}/exec", s.handleExecWS)
		r.Get("/projects/{name}/nodes/{node}/ssh", s.handleSSHWS)
		r.Get("/projects/{name}/links", s.handleGetLinks)
		r.Get("/projects/{name}/links/{id}", s.handleGetLink)
		r.Post("/projects/{name}/links/{id}/fault", s.handleLinkFault)
		r.Get("/events", s.handleEventsWS)
	})

	// Serve SPA from embedded filesystem
	s.router.Handle("/*", http.HandlerFunc(s.serveSPA))

	return s
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// serveSPA serves the embedded React SPA
func (s *Server) serveSPA(w http.ResponseWriter, r *http.Request) {
	distFS, err := fs.Sub(frontend.FS, "dist")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fileServer := http.FileServer(http.FS(distFS))

	// Try to serve the requested file; fall back to index.html for SPA routing
	_, statErr := fs.Stat(distFS, r.URL.Path[1:])
	if r.URL.Path != "/" && statErr == nil {
		fileServer.ServeHTTP(w, r)
	} else {
		// Serve index.html for all unknown paths (SPA client-side routing)
		data, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
