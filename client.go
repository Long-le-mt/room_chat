package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
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
	conn *websocket.Conn

	// Keep a reference to the hub for each Client
	hub *Hub

	// Buffered channel of outbound messages.
	send chan []byte

	// Keeping track of the rooms clients joins
	rooms map[*Room]bool
}

// - New client has connected, initialize it and listening for comming messages
// - Create Client struct
func newClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		conn:  conn,
		hub:   hub,
		send:  make(chan []byte),
		rooms: make(map[*Room]bool),
	}
}

// ServerWs handle websocket request from clients request
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error at server ws:", err)
		return
	}

	// new client connect to websocket
	client := newClient(conn, hub)
	hub.register <- client

	go client.write()
	go client.read()

	log.Printf("new client has joined: %s", client.conn.RemoteAddr().String())
}

func (client *Client) read() {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// client.hub.broadcast <- message
		client.handleNewMessage(message)
	}
}

func (client *Client) write() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The Hub closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	// unregister with server
	client.hub.unregister <- client

	// unregister client with room
	for room := range client.rooms {
		room.unregister <- client
	}

	// close channel send message of client
	close(client.send)

	// close connection with webserver
	client.conn.Close()
}

// To handler message of client
func (client *Client) handleNewMessage(jsonMessage []byte) {

	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
	}

	// Attach the client object as the sender of the messsage.
	message.Sender = client

	switch message.Action {
	case SEND_MESSAGE_ACTION:
		// The send-message action, this will send messages to a specific room now.
		// Which room wil depend on the message Target
		roomName := message.Target
		// Use the ChatServer method to find the room, and if found, broadcast!
		if room := client.hub.findRoomByName(roomName); room != nil {
			room.broadcast <- &message
		}
	// We delegate the join and leave actions.
	case JOIN_ROOM_ACTION:
		client.handleJoinRoomMessage(message)

	case LEAVE_ROOM_ACTION:
		client.handleLeaveRoomMessage(message)
	}
}

// Handle message join room of client
//
// Find room and add client to room
func (client *Client) handleJoinRoomMessage(message Message) {
	roomName := message.Target

	room := client.hub.findRoomByName(roomName)
	if room == nil {
		room = client.hub.createRoom(roomName)
	}

	client.rooms[room] = true

	room.register <- client
}

// Handler message leave room of client
//
// Find room the user want to leave
func (client *Client) handleLeaveRoomMessage(message Message) {
	room := client.hub.findRoomByName(message.Target)
	if _, ok := client.rooms[room]; ok {
		// delete room from set room of client
		delete(client.rooms, room)
	}

	room.unregister <- client
}
