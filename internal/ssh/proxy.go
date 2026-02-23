package ssh

import (
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

// Proxy bridges a WebSocket connection to an SSH session.
type Proxy struct {
	target   string
	user     string
	password string
}

// NewProxy creates a new SSH proxy.
func NewProxy(target, user, password string) *Proxy {
	return &Proxy{target: target, user: user, password: password}
}

// Bridge connects a WebSocket to an SSH session.
func (p *Proxy) Bridge(ws *websocket.Conn) error {
	config := &ssh.ClientConfig{
		User: p.user,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if p.password != "" {
		config.Auth = append(config.Auth, ssh.Password(p.password))
	}
	config.Auth = append(config.Auth, ssh.KeyboardInteractive(
		func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = p.password
			}
			return answers, nil
		},
	))

	client, err := ssh.Dial("tcp", p.target, config)
	if err != nil {
		return fmt.Errorf("SSH dial to %s: %w", p.target, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("SSH session: %w", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", 24, 80, modes); err != nil {
		return fmt.Errorf("SSH request pty: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("SSH stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("SSH stdout pipe: %w", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("SSH shell: %w", err)
	}

	var wg sync.WaitGroup

	// SSH stdout -> WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				if werr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					slog.Debug("ws write error in SSH proxy", "error", werr)
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					slog.Debug("SSH stdout read error", "error", err)
				}
				ws.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
		}
	}()

	// WebSocket -> SSH stdin
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					slog.Debug("ws read error in SSH proxy", "error", err)
				}
				stdin.Close()
				return
			}
			if _, werr := stdin.Write(msg); werr != nil {
				slog.Debug("SSH stdin write error", "error", werr)
				return
			}
		}
	}()

	wg.Wait()
	return nil
}
