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

		log.Printf("raw message: %s", message)

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("unmarshal error: %v (raw: %s)", err, message)
			continue
		}

		log.Printf("parsed message: %+v", msg)

		switch msg.ChannelName {

		case "#ping", "Пинги":
			pingAgg.Add(msg)

		case "Public", "Test":
			publicAgg.Add(msg)

		default:
			// можно логировать, но не слать в Telegram сразу
			log.Printf("unknown channel %q: %q", msg.ChannelName, msg.Message)
		}
	}
}
