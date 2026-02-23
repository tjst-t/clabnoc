package ssh_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tjst-t/clabnoc/internal/ssh"
)

func TestSSHConfig_DefaultPort(t *testing.T) {
	cfg := ssh.Config{Host: "192.168.1.1", Port: 22, User: "admin"}
	assert.Equal(t, "192.168.1.1", cfg.Host)
	assert.Equal(t, 22, cfg.Port)
	assert.Equal(t, "admin", cfg.User)
}

func TestSSHConfig_WithPassword(t *testing.T) {
	cfg := ssh.Config{Host: "10.0.0.1", Port: 22, User: "root", Password: "secret"}
	assert.Equal(t, "secret", cfg.Password)
}

func TestSSHConfig_EmptyPassword(t *testing.T) {
	cfg := ssh.Config{Host: "10.0.0.1", Port: 22, User: "admin"}
	assert.Empty(t, cfg.Password)
}

func TestSSHConfig_CustomPort(t *testing.T) {
	cfg := ssh.Config{Host: "10.0.0.1", Port: 2222, User: "admin"}
	assert.Equal(t, 2222, cfg.Port)
}

func TestSSHConfig_ZeroValues(t *testing.T) {
	cfg := ssh.Config{}
	assert.Empty(t, cfg.Host)
	assert.Equal(t, 0, cfg.Port)
	assert.Empty(t, cfg.User)
	assert.Empty(t, cfg.Password)
}
