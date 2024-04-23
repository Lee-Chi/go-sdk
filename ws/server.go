package ws

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		id = uuid.New().String()
	}

	// Register our new client
	client := &Client{
		id:     id,
		hub:    hub,
		socket: ws,
		send:   make(chan []byte, 20),
	}

	hub.register <- client

	client.Run()
}
