package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types/container"
	"github.com/gorilla/websocket"
)

// ExecWebSocket runs a command in a container and bridges stdin/stdout to a WebSocket
func ExecWebSocket(ctx context.Context, client DockerClient, containerID string, cmd []string, conn *websocket.Conn) error {
	// Create exec
	execResp, err := client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	})
	if err != nil {
		return fmt.Errorf("create exec: %w", err)
	}

	// Attach to exec
	hijacked, err := client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return fmt.Errorf("attach exec: %w", err)
	}
	defer hijacked.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 2)

	// Copy stdout -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}
			n, err := hijacked.Reader.Read(buf)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					errCh <- fmt.Errorf("write to websocket: %w", werr)
					return
				}
			}
			if err != nil {
				if err == io.EOF {
					errCh <- nil
				} else {
					errCh <- fmt.Errorf("read from container: %w", err)
				}
				return
			}
		}
	}()

	// Copy WebSocket -> stdin
	go func() {
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					errCh <- nil
				} else {
					errCh <- fmt.Errorf("read from websocket: %w", err)
				}
				return
			}
			if _, werr := hijacked.Conn.Write(msg); werr != nil {
				errCh <- fmt.Errorf("write to container stdin: %w", werr)
				return
			}
		}
	}()

	// Wait for first goroutine to finish
	err = <-errCh
	if err != nil {
		slog.Debug("exec websocket ended", "err", err)
	}
	return nil
}
