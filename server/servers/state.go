package servers

import (
	"net/http"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type ServerState struct {
	visitors map[string]*rate.Limiter
	connected map[string]map[string]*websocket.Conn
	allowedIps map[string]bool
	streams map[string]*Stream
}

type Stream struct {
	Name 	string
	IsLive 	bool
	Title 	string
	Viewers uint32
}

func NewServerState() *ServerState {
	server := &ServerState{
		visitors: make(map[string]*rate.Limiter),
		connected: make(map[string]map[string]*websocket.Conn),
		allowedIps: make(map[string]bool),
		streams: make(map[string]*Stream),
	}
	server.allowedIps["172.18.0.1"] = true
	server.connected["jolsho"] = make(map[string]*websocket.Conn)
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

	room := r.URL.Query().Get("room")
	if _,exists := s.connected[room]; !exists {
		http.Error(w, "Room doesnt exist", http.StatusBadRequest)
		return
	}

	// Upgrade the connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()
	s.connected[room][r.RemoteAddr] = conn

	for {

		// Read message from client
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			delete(s.connected[room], r.RemoteAddr)
			log.Println("Read error:", err)
			break
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
		for ip, c := range s.connected[room] {
			if err := c.WriteMessage(websocket.TextMessage, out); err != nil {
				log.Println("Write error:", err)
				delete(s.connected[room], ip)
				continue
			}
		}
	}
}


