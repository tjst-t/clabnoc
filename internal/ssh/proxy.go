package ssh

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
	gossh "golang.org/x/crypto/ssh"
)

// Config holds SSH connection parameters
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	// If Password is empty, try common defaults
}

// ProxyWebSocket connects to the target via SSH and bridges stdin/stdout to a WebSocket
func ProxyWebSocket(ctx context.Context, cfg Config, conn *websocket.Conn) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	authMethods := []gossh.AuthMethod{}
	if cfg.Password != "" {
		authMethods = append(authMethods, gossh.Password(cfg.Password))
	}
	// Try keyboard-interactive with empty password as fallback
	authMethods = append(authMethods, gossh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) ([]string, error) {
		answers := make([]string, len(questions))
		return answers, nil
	}))

	sshConfig := &gossh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec // Personal VM, acceptable
		Timeout:         10 * time.Second,
	}

	client, err := gossh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("ssh dial %s: %w", addr, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("ssh new session: %w", err)
	}
	defer session.Close()

	// Set up PTY
	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 80, 24, modes); err != nil {
		return fmt.Errorf("request pty: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("start shell: %w", err)
	}

	done := make(chan error, 2)

	// Copy WebSocket -> SSH stdin
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				done <- err
				return
			}
			if _, err := stdin.Write(data); err != nil {
				done <- err
				return
			}
		}
	}()

	// Copy SSH stdout -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		merged := io.MultiReader(stdout, stderr)
		for {
			n, err := merged.Read(buf)
			if n > 0 {
				if err2 := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err2 != nil {
					done <- err2
					return
				}
			}
			if err != nil {
				done <- err
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
