package main

type Hub struct {
	//
	Clients map[*Client]bool

	//
	Register chan *Client

	//
	Unregister chan *Client

	//
	Broadcast chan []byte
}

// - Create Hub struct
func newHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
	}
}

// Run websocket server, accepting requests from clients, it will run infinitely
func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.Register:
			hub.Clients[client] = true
		case client := <-hub.Unregister:
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				close(client.send)
			}
		case message := <-hub.Broadcast:
			for client := range hub.Clients {
				client.send <- message
			}
		}
	}

}
