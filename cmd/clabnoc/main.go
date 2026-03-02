package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/tjst-t/clabnoc/internal/api"
	"github.com/tjst-t/clabnoc/internal/capture"
	"github.com/tjst-t/clabnoc/internal/docker"
	"github.com/tjst-t/clabnoc/internal/network"
)

var version = "dev"

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dev := flag.Bool("dev", false, "development mode (CORS permissive)")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Println("clabnoc", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if *dev {
		slog.Info("development mode enabled")
	}

	dockerClient, err := docker.NewRealClient()
	if err != nil {
		slog.Error("failed to create Docker client", "error", err)
		os.Exit(1)
	}

	execFn := func(ctx context.Context, containerID string, cmd []string) (string, error) {
		return docker.ExecCommand(ctx, dockerClient, containerID, cmd)
	}
	faultManager := network.NewFaultManager(network.NewDockerFaultOperator(execFn))

	pidProvider := capture.NewDockerPIDProvider(dockerClient)
	vethResolver := capture.NewProcVethResolver(pidProvider)
	captureManager := capture.NewCaptureManager(&capture.HostCaptureExecutor{}, "/tmp/clabnoc/captures")

	server := &api.Server{
		Docker:         dockerClient,
		FaultManager:   faultManager,
		CaptureManager: captureManager,
		VethResolver:   vethResolver,
	}

	router := api.NewRouter(server)

	slog.Info("starting clabnoc", "addr", *addr, "version", version)
	if err := http.ListenAndServe(*addr, router); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
