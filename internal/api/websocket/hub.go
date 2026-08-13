package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	upgrader  websocket.Upgrader
}

var (
	instance *Hub
	once     sync.Once
)

// GetHub returns the singleton instance of the WebSocket Hub
func GetHub() *Hub {
	once.Do(func() {
		instance = &Hub{
			clients: make(map[*websocket.Conn]bool),
			upgrader: websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true // Allow all origins for developers/local environments
				},
			},
		}
	})
	return instance
}

// HandleConnection handles WebSocket handshakes and client state management
func (h *Hub) HandleConnection(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS ERROR] Failed to upgrade connection: %v", err)
		return
	}

	h.clientsMu.Lock()
	h.clients[conn] = true
	h.clientsMu.Unlock()

	log.Printf("[WS INFO] Client connected: %s. Active clients: %d", conn.RemoteAddr(), len(h.clients))

	// Keep connection alive, listen for close events, and clean up connection state
	go func() {
		defer func() {
			conn.Close()
			h.clientsMu.Lock()
			delete(h.clients, conn)
			h.clientsMu.Unlock()
			log.Printf("[WS INFO] Client disconnected. Active clients: %d", len(h.clients))
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// Broadcast broadcasts a message of a specific type to all connected clients
func (h *Hub) Broadcast(msgType string, payload interface{}) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	if len(h.clients) == 0 {
		return
	}

	data := map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("[WS ERROR] Failed to marshal broadcast payload: %v", err)
		return
	}

	log.Printf("[WS INFO] Broadcasting message of type '%s' to %d client(s)", msgType, len(h.clients))

	for conn := range h.clients {
		err := conn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			log.Printf("[WS WARNING] Failed to send message to client %s: %v", conn.RemoteAddr(), err)
		}
	}
}
