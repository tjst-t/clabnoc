package api

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tjst-t/clabnoc/internal/capture"
	"github.com/tjst-t/clabnoc/internal/docker"
	"github.com/tjst-t/clabnoc/internal/frontend"
	"github.com/tjst-t/clabnoc/internal/network"
)

// Server holds the API server dependencies.
type Server struct {
	Docker         docker.DockerClient
	FaultManager   *network.FaultManager
	CaptureManager *capture.CaptureManager
	VethResolver   capture.VethResolver
}

// NewRouter creates the HTTP router with all API routes.
func NewRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/projects", s.listProjects)
		r.Get("/projects/{name}/topology", s.getTopology)
		r.Get("/projects/{name}/nodes", s.listNodes)
		r.Get("/projects/{name}/nodes/{node}", s.getNode)
		r.Post("/projects/{name}/nodes/{node}/action", s.nodeAction)
		r.Get("/projects/{name}/nodes/{node}/ssh-credentials", s.getSSHCredentials)
		r.Get("/projects/{name}/nodes/{node}/exec", s.execTerminal)
		r.Get("/projects/{name}/nodes/{node}/ssh", s.sshTerminal)
		r.Get("/projects/{name}/links", s.listLinks)
		r.Get("/projects/{name}/links/{id}", s.getLink)
		r.Post("/projects/{name}/links/{id}/fault", s.injectFault)
		r.Post("/projects/{name}/links/{id}/capture", s.captureAction)
		r.Get("/projects/{name}/links/{id}/capture/download", s.captureDownload)
		r.Get("/projects/{name}/stats", s.stats)
		r.Get("/bpf-presets", s.listBPFPresets)
		r.Get("/events", s.events)
	})

	// noVNC reverse proxy (outside /api/v1 so relative paths in noVNC work)
	r.HandleFunc("/proxy/{name}/{node}/*", s.proxyHandler)

	// Serve frontend SPA
	distFS, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		slog.Error("failed to create sub filesystem for frontend", "error", err)
	} else {
		fileServer := http.FileServer(http.FS(distFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file directly; if not found, serve index.html for SPA routing
			f, err := distFS.Open(r.URL.Path[1:]) // strip leading /
			if err != nil {
				// Serve index.html for SPA client-side routing
				r.URL.Path = "/"
			} else {
				f.Close()
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}
