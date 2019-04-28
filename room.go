package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	//
	foward chan []byte
	// clients que desean unirse al room
	join chan *client
	// clients que desean abandonar el room
	leave chan *client
	// Todos los clientes en el room
	clients map[*client]bool
}

// Crear un nuevo room

func newRoom() *room {
	return &room{
		foward:  make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

// El ciclo for se ejecuta hasta que el programa termine
// cuando un usuario entra al cuarto se almacena en el map
// y se setea su valor a true
// cuando sale del cuarto se elimina el usuario del map

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.foward:
			for client := range r.clients {
				client.send <- msg
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: messageBufferSize}

// Salir del room se deja como defer para que se ejecute cuando
// sea necesario

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("Serve HTTP:", err)
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
