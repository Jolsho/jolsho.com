package servers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type ServerState struct {
	visitors map[string]*rate.Limiter
	visit_mux *sync.RWMutex

	connected map[string]*Room
	connected_mux *sync.RWMutex

	allowedIps map[string]bool

	streams map[string]*Stream
	streams_mux *sync.RWMutex
}

type Room struct {
	conns map[string]*websocket.Conn
	mux *sync.RWMutex
}

type Stream struct {
	Name 	string
	IsLive 	bool
	Title 	string
	Viewers uint32
	mux *sync.RWMutex
}

func NewServerState() *ServerState {
	server := &ServerState{
		allowedIps: make(map[string]bool),

		visitors: make(map[string]*rate.Limiter),
		visit_mux: &sync.RWMutex{},

		connected: make(map[string]*Room),
		connected_mux: &sync.RWMutex{},

		streams: make(map[string]*Stream),
		streams_mux: &sync.RWMutex{},
	}
	server.allowedIps["172.18.0.1"] = true
	room := &Room {
		conns: make(map[string]*websocket.Conn),
		mux: &sync.RWMutex{},
	}
	server.connected["jolsho"] = room
	server.streams["jolsho"] = &Stream{
		Name: "jolsho",
		IsLive: false,
		Title: "Will be back tomorrow.",
		Viewers: 0,
	}
	return server
}

func (s *ServerState) IsLive(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	if stream, exists := s.streams[room]; exists {
		payload, err := json.Marshal(stream)
		if err == nil {
			w.WriteHeader(http.StatusOK)
			w.Write(payload)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
}

// Message represents the incoming and outgoing chat messages
type Message struct {
	Code		int 	`json:"code"`
	Text      	string 	`json:"text"`
	Timestamp 	string 	`json:"timestamp,omitempty"`
}

// WebSocket upgrader with some sensible defaults
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now; tighten this in production
		return true
	},
}

func (s *ServerState) HandleChat(w http.ResponseWriter, r *http.Request) {

	room_name := r.URL.Query().Get("room")
	s.connected_mux.RLock()
	room ,exists := s.connected[room_name]
	if !exists {
		http.Error(w, "Room doesnt exist", http.StatusBadRequest)
		return
	}
	s.connected_mux.RUnlock()

	// Upgrade the connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()
	s.connected_mux.Lock()
	room.mux.Lock()
	room.conns[r.RemoteAddr] = conn
	room.mux.Unlock()
	// Create a rate limiter for this connection: 5 messages per 10 seconds
	limiter := rate.NewLimiter(rate.Every(2*time.Second), 5)

	for {

		// Read message from client
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			room.mux.Lock()
			delete(room.conns, r.RemoteAddr)
			log.Println("Read error:", err)
			break
		}

		// Check rate limit
		if !limiter.Allow() {
			log.Printf("Rate limit exceeded for %s\n", r.RemoteAddr)
			continue // drop message
		}

		// Parse incoming JSON
		var msg Message
		if err := json.Unmarshal(msgData, &msg); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		// Add timestamp if missing
		if msg.Timestamp == "" {
			msg.Timestamp = time.Now().Format("15:04")
		}

		// Echo the message back to clients
		out, _ := json.Marshal(msg)
		for ip, c := range room.conns {
			if err := c.WriteMessage(websocket.TextMessage, out); err != nil {
				log.Println("Write error:", err)
				delete(room.conns, ip)
				continue
			}
		}
	}
}


