package ws

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// Maximum message size allowed from peer.
	MaxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ID string

func (id ID) String() string {
	return string(id)
}

func IDFromString(id string) ID {
	return ID(id)
}

func NewID() ID {
	return ID(uuid.New().String())
}

type Command struct {
	Name string `json:"name"`
	Body []byte `json:"body"`
}

func NewCommand(name string, body []byte) Command {
	return Command{
		Name: name,
		Body: body,
	}
}

func (c Command) Marshal() []byte {
	data, _ := json.Marshal(c)
	return data
}

func (c *Command) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}

type Packet struct {
	To      ID
	Message []byte
}

type Handler func(*Connection, []byte)
