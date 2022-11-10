package main

type Hub struct {
	// - Set clients registed, connected to websocket server
	clients map[*Client]bool

	// - Register requests from server
	register chan *Client

	// - Unregister requests from server
	unregister chan *Client

	// - Channel to send broadcast from clients to the server
	broadcast chan []byte
}

// - Create Hub struct
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Run websocket server, accepting requests from clients, it will run infinitely
func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.registerClient(client)
		case client := <-hub.unregister:
			hub.unregisterClient(client)
		case message := <-hub.broadcast:
			hub.broadcastToClients(message)
		}
	}
}

// Add client to set clients
func (hub *Hub) registerClient(client *Client) {
	hub.clients[client] = true
}

// Remove client from set clients and close channel send message of it,
func (hub *Hub) unregisterClient(client *Client) {
	if _, ok := hub.clients[client]; ok {
		delete(hub.clients, client)
		close(client.send)
	}
}

// Send broadcast to clients were registered, connected to websocket server
func (hub *Hub) broadcastToClients(message []byte) {
	for client := range hub.clients {
		client.send <- message
	}
}
