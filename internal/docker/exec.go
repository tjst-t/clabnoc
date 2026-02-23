package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/gorilla/websocket"
)

// ExecSession manages a docker exec session bridged to a WebSocket.
type ExecSession struct {
	cli         DockerClient
	containerID string
	cmd         string
}

// NewExecSession creates a new docker exec session.
func NewExecSession(cli DockerClient, containerID, cmd string) *ExecSession {
	if cmd == "" {
		cmd = "/bin/bash"
	}
	return &ExecSession{cli: cli, containerID: containerID, cmd: cmd}
}

// Bridge bridges a WebSocket connection to a docker exec session.
func (s *ExecSession) Bridge(ctx context.Context, ws *websocket.Conn) error {
	execConfig := container.ExecOptions{
		Cmd:          []string{s.cmd},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	resp, err := s.cli.ContainerExecCreate(ctx, s.containerID, execConfig)
	if err != nil {
		return fmt.Errorf("exec create: %w", err)
	}

	hijack, err := s.cli.ContainerExecAttach(ctx, resp.ID, container.ExecStartOptions{Tty: true})
	if err != nil {
		return fmt.Errorf("exec attach: %w", err)
	}
	defer hijack.Close()

	var wg sync.WaitGroup

	// Docker -> WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := hijack.Reader.Read(buf)
			if n > 0 {
				if werr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					slog.Debug("ws write error", "error", werr)
					return
				}
			}
			if err != nil {
				if err != io.EOF && !isClosedError(err) {
					slog.Debug("docker read error", "error", err)
				}
				ws.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
		}
	}()

	// WebSocket -> Docker
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) && !isClosedError(err) {
					slog.Debug("ws read error", "error", err)
				}
				hijack.Close()
				return
			}
			if _, werr := hijack.Conn.Write(msg); werr != nil {
				slog.Debug("docker write error", "error", werr)
				return
			}
		}
	}()

	wg.Wait()
	return nil
}

func isClosedError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*net.OpError); ok {
		return true
	}
	return false
}
