package ws

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Connection struct {
	id      ID
	hub     *Hub
	socket  *websocket.Conn
	send    chan []byte
	destroy Handler
}

func NewConnection(id ID, hub *Hub, socket *websocket.Conn, destroy Handler) *Connection {
	return &Connection{
		id:      id,
		hub:     hub,
		socket:  socket,
		send:    make(chan []byte, 32),
		destroy: destroy,
	}
}

func (c Connection) ID() ID {
	return c.id
}

func (c Connection) Send(to ID, cmd Command) {
	if to == c.ID() {
		c.send <- cmd.Marshal()
		return
	}

	c.hub.relay <- Packet{To: to, Message: cmd.Marshal()}
}

func (c Connection) Broadcast(cmd Command) {
	c.hub.broadcast <- Packet{Message: cmd.Marshal()}
}

func (c *Connection) daemon() {
	go c.read()
	go c.write()
}

func (c *Connection) read() {
	defer func() {
		c.hub.log <- fmt.Sprintf("connection %s, leave read", c.id)
		c.hub.unregister <- c
		if c.destroy != nil {
			c.destroy(c, nil)
		}
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.log <- fmt.Sprintf("connection %s, failed to read message: %s", c.id, err)
			} else {
				c.hub.log <- fmt.Sprintf("connection %s, close message: %s", c.id, err)
			}
			break
		}

		var command Command
		if err := command.Unmarshal(message); err != nil {
			c.hub.log <- fmt.Sprintf("connection %s receive(%s), failed to unmarshal message: %s", c.id, string(message), err)
			continue
		}

		handler, ok := c.hub.handlers[command.Name]
		if !ok {
			c.hub.log <- fmt.Sprintf("connection %s recieve(%s), unknown API: %s", c.id, string(message), command.Name)
			continue
		}

		go handler(c, command.Body)
	}
}

func (c *Connection) write() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		c.hub.log <- fmt.Sprintf("connection %s, leave write", c.id)
		ticker.Stop()
		c.socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.socket.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.hub.log <- fmt.Sprintf("connection %s, failed to get message", c.id)
				// The hub closed the channel.
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.socket.NextWriter(websocket.TextMessage)
			if err != nil {
				c.hub.log <- fmt.Sprintf("connection %s, failed to get next writer: %s", c.id, err)
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				c.hub.log <- fmt.Sprintf("connection %s, failed to close writer: %s", c.id, err)
				return
			}
		case <-ticker.C:
			c.socket.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.log <- fmt.Sprintf("connection %s, failed to write ping message: %s", c.id, err)
				return
			}
		}
	}
}
