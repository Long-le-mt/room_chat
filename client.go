package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//  method upgrades the HTTP server connection to the WebSocket protocol
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// - Client representer the websocket client at the server
//
// - Client is reposible for keeping connection, user info, websocket connection,..
type Client struct {
	// The websocket connection
	Conn *websocket.Conn

	// Keep a reference to the hub for each Client
	Hub *Hub
}

// - New client has connected, initialize it and listening for comming messages
// - Create Client struct
func newClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
	}
}

// ServerWs handle websocket request from clients request
func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error at server ws:", err)
		return
	}

	// new client connect to websocket
	client := newClient(conn)

	log.Printf("new client has joined: %s", client.Conn.RemoteAddr().String())
}
