package main

// Room looks like struct of hub, room should be able to register clients, unregister clients, and boardcast to client
type Room struct {
	// - Name of room
	name string

	// - Set clients registed, connected to websocket server
	clients map[*Client]bool

	// - Register requests from server
	register chan *Client

	// - Unregister requests from server
	unregister chan *Client

	// - Channel to send broadcast from clients to the server
	broadcast chan []byte
}

// - Create a new room
func newRoom(name string) *Room {
	return &Room{
		name:       name,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Runroom runs our room, accepting various request
func (room *Room) runRoom() {
	for {
		select {

		case client := <-room.register:
			room.registerClientInRoom(client)

		case client := <-room.unregister:
			room.unregisterClientInRoom(client)

		case message := <-room.broadcast:
			room.broadcastToClientsInRoom(message)
		}
	}
}

func (room *Room) registerClientInRoom(client *Client) {
	// room.notifyClientJoined(client)
	room.clients[client] = true
}

func (room *Room) unregisterClientInRoom(client *Client) {
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
		close(client.send)
	}
}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	for client := range room.clients {
		client.send <- message
	}
}

func (room *Room) getName() string {
	return room.getName()
}