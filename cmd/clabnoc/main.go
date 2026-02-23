package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/tjst-t/clabnoc/internal/api"
	"github.com/tjst-t/clabnoc/internal/docker"
)

var version = "dev"

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dev := flag.Bool("dev", false, "dev mode (CORS relaxed)")
	flag.Parse()

	slog.Info("starting clabnoc", "version", version, "addr", *addr)

	dockerClient, err := docker.NewClient()
	if err != nil {
		slog.Error("failed to create Docker client", "err", err)
		os.Exit(1)
	}

	server := api.NewServer(dockerClient)

	slog.Info("listening", "addr", *addr)
	if err := http.ListenAndServe(*addr, server); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}

	_ = dev
}
