package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type EventMsg struct {
	Level   string `json:"level"`
	Message string `json:"msg"`
}

var wsClient websocket.Conn

func NotifyConnect() {
	wsClient, _, err := websocket.DefaultDialer.Dial(WS_NOTIFY_URL, nil)
	if err != nil {
		log.Fatal("Unable to connect to notify WS server", err)
	}
	defer wsClient.Close()

	log.Println("WS notifier connected")
}

func SendEvent(level, message string) {
	wsClient.WriteJSON(EventMsg{
		Level:   level,
		Message: message,
	})
}
