package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
)

// StartSessionListener binds the given port and starts an HTTP server
// that routes WebSocket upgrades to WsProxyHandler and other HTTP requests
// to HttpProxyHandler. If port is 0, the OS assigns a free port.
// Returns the actual bound port and any error.
func StartSessionListener(session *Session, port int) (int, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("failed to bind port %d: %w", port, err)
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port

	session.mu.Lock()
	session.listener = listener
	session.Port = actualPort
	session.mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a WebSocket upgrade request
		if isWebSocketUpgrade(r) {
			HandleWsProxy(session, w, r)
			return
		}
		// Otherwise treat as HTTP proxy
		HandleHttpProxy(session, w, r)
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("session %s listener on port %d closed: %v", session.ID, actualPort, err)
		}
	}()

	// Store server reference so we can shut it down later
	session.mu.Lock()
	session.Server = server
	session.mu.Unlock()

	return actualPort, nil
}

// StopSessionListener gracefully shuts down the per-session HTTP server and closes the listener.
func StopSessionListener(session *Session) {
	session.mu.Lock()
	server := session.Server
	session.Server = nil
	session.mu.Unlock()

	if server != nil {
		server.Shutdown(context.Background())
	}
}

func isWebSocketUpgrade(r *http.Request) bool {
	for _, v := range r.Header["Upgrade"] {
		if v == "websocket" {
			return true
		}
	}
	return false
}
