package api

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/tjst-t/clabnoc/internal/docker"
)

// proxyHandler proxies HTTP and WebSocket requests to a node's management IP.
// Route: /proxy/{name}/{node}/*
func (s *Server) proxyHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")

	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("proxy: failed to get topology", "project", name, "error", err)
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	var mgmtIP string
	for _, n := range topo.Nodes {
		if n.Name == nodeName {
			mgmtIP = n.MgmtIPv4
			break
		}
	}
	if mgmtIP == "" {
		http.Error(w, "node not found or no management IP", http.StatusNotFound)
		return
	}

	// Get the remaining path after /proxy/{name}/{node}
	proxyPath := chi.URLParam(r, "*")
	proxyPath = strings.TrimPrefix(proxyPath, "/")

	// Store auth credentials from HTTP requests for WebSocket reuse.
	// Browsers don't send Basic Auth on WebSocket upgrades, so the proxy
	// remembers credentials from the initial HTTP page load.
	proxyKey := name + "/" + nodeName
	if auth := r.Header.Get("Authorization"); auth != "" {
		proxyAuthCache.Store(proxyKey, auth)
	}

	if websocket.IsWebSocketUpgrade(r) {
		s.proxyWebSocket(w, r, mgmtIP, proxyPath, proxyKey)
		return
	}
	s.proxyHTTP(w, r, mgmtIP, proxyPath)
}

// proxyAuthCache stores Authorization headers from HTTP requests,
// keyed by "project/node", so WebSocket connections can reuse them.
var proxyAuthCache sync.Map

// schemeCache caches whether a host:443 speaks TLS or plain HTTP.
var schemeCache sync.Map

// detectScheme probes host:443 to determine if it speaks TLS.
func detectScheme(host string) string {
	if v, ok := schemeCache.Load(host); ok {
		return v.(string)
	}

	scheme := "http"
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 2 * time.Second},
		"tcp", host+":443",
		&tls.Config{InsecureSkipVerify: true},
	)
	if err == nil {
		conn.Close()
		scheme = "https"
	}

	slog.Info("proxy: detected backend scheme", "host", host, "scheme", scheme)
	schemeCache.Store(host, scheme)
	return scheme
}

func (s *Server) proxyHTTP(w http.ResponseWriter, r *http.Request, mgmtIP, path string) {
	scheme := detectScheme(mgmtIP)
	// Always use port 443 — qemu-bmc serves on :443 regardless of TLS
	hostPort := mgmtIP + ":443"
	target, _ := url.Parse(fmt.Sprintf("%s://%s", scheme, hostPort))

	var transport http.RoundTripper
	if scheme == "https" {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		transport = http.DefaultTransport
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = "/" + path
			req.Host = mgmtIP
		},
		Transport: transport,
	}
	proxy.ServeHTTP(w, r)
}

func (s *Server) proxyWebSocket(w http.ResponseWriter, r *http.Request, mgmtIP, path, proxyKey string) {
	scheme := detectScheme(mgmtIP)

	var wsScheme string
	dialer := websocket.Dialer{
		Subprotocols: websocket.Subprotocols(r),
	}
	if scheme == "https" {
		wsScheme = "wss"
		// Force HTTP/1.1 — WebSocket Upgrade does not work over HTTP/2
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"http/1.1"},
		}
	} else {
		wsScheme = "ws"
	}

	backendURL := fmt.Sprintf("%s://%s:443/%s", wsScheme, mgmtIP, path)

	// Build headers for backend connection
	reqHeader := http.Header{}

	// Use auth from the request if present, otherwise fall back to cached auth
	// from a prior HTTP request (browsers don't send Basic Auth on WebSocket)
	if auth := r.Header.Get("Authorization"); auth != "" {
		reqHeader.Set("Authorization", auth)
	} else if cached, ok := proxyAuthCache.Load(proxyKey); ok {
		reqHeader.Set("Authorization", cached.(string))
	}

	backendConn, backendResp, err := dialer.Dial(backendURL, reqHeader)
	if err != nil {
		status := 0
		if backendResp != nil {
			status = backendResp.StatusCode
		}
		slog.Error("proxy: failed to connect to backend WebSocket",
			"url", backendURL, "error", err, "backend_status", status)
		if status == http.StatusUnauthorized {
			w.Header().Set("WWW-Authenticate", `Basic realm="noVNC"`)
			http.Error(w, "authentication required", http.StatusUnauthorized)
		} else {
			http.Error(w, "failed to connect to backend", http.StatusBadGateway)
		}
		return
	}
	defer backendConn.Close()

	// Upgrade client connection, passing the negotiated subprotocol
	respHeader := http.Header{}
	if proto := backendResp.Header.Get("Sec-WebSocket-Protocol"); proto != "" {
		respHeader.Set("Sec-WebSocket-Protocol", proto)
	}

	clientConn, err := upgrader.Upgrade(w, r, respHeader)
	if err != nil {
		slog.Error("proxy: failed to upgrade client WebSocket", "error", err)
		return
	}
	defer clientConn.Close()

	slog.Info("proxy: WS relay started",
		"remote", r.RemoteAddr, "backend", backendURL)

	// Relay messages bidirectionally
	errc := make(chan error, 2)

	go func() {
		for {
			msgType, msg, err := backendConn.ReadMessage()
			if err != nil {
				errc <- err
				return
			}
			if err := clientConn.WriteMessage(msgType, msg); err != nil {
				errc <- err
				return
			}
		}
	}()

	go func() {
		for {
			msgType, msg, err := clientConn.ReadMessage()
			if err != nil {
				errc <- err
				return
			}
			if err := backendConn.WriteMessage(msgType, msg); err != nil {
				errc <- err
				return
			}
		}
	}()

	err = <-errc
	slog.Info("proxy: WS relay ended", "remote", r.RemoteAddr, "error", err)
}
