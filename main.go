package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// Message represents the structure of incoming WebSocket messages
type Message struct {
	ChannelID            int     `json:"channel_hash"`
	ControlSubtype       string  `json:"control_subtype"`
	DestHash             string  `json:"dest_hash"`
	Direction            string  `json:"direction"`
	DisplayCombinedPath  string  `json:"display_combined_path"`
	ChannelName          string  `json:"group_channel_name"`
	Message              string  `json:"group_message_text"`
	MessageHash          string  `json:"hash"`
	MessageID            int     `json:"id"`
	Observer             string  `json:"observer"`
	ObserverId           string  `json:"observer_id"`
	ObserverPubkeyPrefix string  `json:"observer_pubkey_prefix"`
	PacketPath           string  `json:"packet_path"`
	PacketSummary        string  `json:"packet_summary"`
	PathReturnedHops     string  `json:"path_returned_hops"`
	PayloadType          int     `json:"payload_type"`
	PropagationId        int     `json:"propagation_id"`
	RegionCode           string  `json:"region_code"`
	RouteType            int     `json:"route_type"`
	Rssi                 float64 `json:"rssi"`
	SenderName           string  `json:"sender_name"`
	Snr                  float64 `json:"snr"`
	SrcHash              string  `json:"src_hash"`
	TsIngest             string  `json:"ts_ingest"`
	TsMqtt               string  `json:"ts_mqtt"`
	TsObserver           string  `json:"ts_observer"`
}

func main() {
	headers := http.Header{}

	headers.Set("User-Agent", "Mozilla/5.0")

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(
		"wss://www.meshcoretel.ru/ws/packets?region_code=MQF",
		headers,
	)

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal("config error:", err)
	}

	sender := NewTelegramSender(cfg)
	pingAggregator := NewPingAggregator(cfg, sender)
	publicAggregator := NewPublicAggregator(cfg, sender)

	if err != nil {
		log.Fatal("ws connect error:", err)
	}
	defer conn.Close()

	// 5. Start loop
	err = readLoop(conn, publicAggregator, pingAggregator)
	if err != nil {
		log.Fatal("read loop error:", err)
	}
}
