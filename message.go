package main

import (
	"encoding/json"
	"fmt"
	"log"
)

const (
	SEND_MESSAGE_ACTION = "send-message"
	JOIN_ROOM_ACTION    = "join-room"
	LEAVE_ROOM_ACTION   = "leave-room"
)

// - Message struct to handle different kinds of message like join room, leave room and send message
type Message struct {
	// - Action of message: join, leave, send-message, ...
	Action string `json:"action"`

	// - Content of message
	Message string `json:"message"`

	// - Target of message, e.g: a room
	Target string `json:"target"`

	// - client sending the message
	Sender *Client `json:"sender"`
}

func (msg *Message) encode() []byte {
	json, err := json.Marshal(msg)
	if err != nil {
		log.Println("Err when encode message:", err)
	}

	fmt.Println("json", json)
	return json
}
