package main

import (
	"github.com/gorilla/websocket"
)

type client struct {
	// socket is the web socket for this client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan []byte
	// room is the room this client is chatting in.
	room *room
}

func (c *client) exit() {
	c.socket.Close()
}

func (c *client) read() {
	defer c.exit()
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
}

func (c *client) write() {
	defer c.exit()
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}
