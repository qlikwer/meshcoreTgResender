package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

func readLoop(
	conn *websocket.Conn,
	publicAgg *PublicAggregator,
	pingAgg *PingAggregator,
) error {

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.ChannelName {

		case "#ping":
			pingAgg.Add(msg)

		case "Public", "Test":
			publicAgg.Add(msg)

		default:
			// можно логировать, но не слать в Telegram сразу
			log.Printf("unknown channel %s: %s", msg.ChannelName, msg.Message)
		}
	}
}
