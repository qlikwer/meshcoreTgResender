package main

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
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

// connect устанавливает WebSocket-соединение и держит его, переподключаясь
// при ошибке подключения или разрыве канала.
func connect(headers http.Header, publicAgg *PublicAggregator, pingAgg *PingAggregator) {
	for {
		dialer := websocket.DefaultDialer
		conn, resp, err := dialer.Dial(
			"wss://meshcoretel.ru/ws/packets?region_code=MQF",
			headers,
		)

		if err != nil {
			if resp != nil {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				log.Printf("Ошибка подключения: %v (статус: %s, Location: %s, тело: %s). Повтор через 10 секунд", err, resp.Status, resp.Header.Get("Location"), body)
			} else {
				log.Printf("Ошибка подключения: %v. Повтор через 10 секунд", err)
			}
			time.Sleep(10 * time.Second)
			continue
		}

		log.Println("Подключено")
		err = readLoop(conn, publicAgg, pingAgg)
		conn.Close()

		log.Printf("Соединение потеряно: %v. Переподключение через 5 секунд", err)
		time.Sleep(5 * time.Second)
	}
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal("config error:", err)
	}

	sender := NewTelegramSender(cfg)
	pingAggregator := NewPingAggregator(cfg, sender)
	publicAggregator := NewPublicAggregator(cfg, sender)

	headers := http.Header{}
	headers.Set("User-Agent", "Mozilla/5.0")

	connect(headers, publicAggregator, pingAggregator)
}
