package main

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/khanek/GoProgrammingBlueprints/Chapter1/trace"
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients
	forward chan []byte
	// join is a channel for clients wishing to join the room
	join chan *client
	// leave is a channel for clients wishing to leave the room
	leave chan *client
	// clients holds all current clients in this room
	clients map[*client]bool
	// tracer will receive trace information of activity
	// in the room.
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) clientJoin(c *client) {
	r.clients[c] = true
	r.tracer.Trace("New client joined")
}

func (r *room) clientLeave(c *client) {
	delete(r.clients, c)
	c.exit()
	r.tracer.Trace("Client left")
}

func (r *room) newMsg(msg []byte) {
	r.tracer.Trace("Message received: ", string(msg))
	for client := range r.clients {
		select {
		case client.send <- msg:
			// send the messages
			r.tracer.Trace(" -- sent to client")
		default:
			// failed to send
			r.clientLeave(client)
			r.tracer.Trace(" -- failed to send, cleaned up client")
		}
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clientJoin(client)
		case client := <-r.leave:
			r.clientLeave(client)
		case msg := <-r.forward:
			r.newMsg(msg)
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	socket, err := upgrader.Upgrade(w, request, nil)
	if err != nil {
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
