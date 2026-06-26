package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
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

func connect(headers http.Header) {
	for {
		dialer := websocket.DefaultDialer
		conn, _, err := dialer.Dial(
			"wss://www.meshcoretel.ru/ws/packets?region_code=MQF",
			headers,
		)

		if err != nil {
			log.Printf(
				"Ошибка подключения: %v. Повтор через 10 секунд",
				err,
			)

			time.Sleep(10 * time.Second)
			continue
		}

		log.Println("Подключено")
		err = readLoop(conn)
		conn.Close()

		log.Printf(
			"Соединение потеряно: %v. Переподключение через 5 секунд",
			err,
		)
		time.Sleep(5 * time.Second)
	}
}

type Deduplicator struct {
	seen map[string]bool
	fifo []string
	size int
}

func NewDeduplicator(size int) *Deduplicator {
	return &Deduplicator{
		seen: make(map[string]bool),
		fifo: make([]string, 0, size),
		size: size,
	}
}

func (d *Deduplicator) Exists(hash string) bool {
	return d.seen[hash]
}

func (d *Deduplicator) Add(hash string) {
	if d.seen[hash] {
		return
	}

	d.fifo = append(d.fifo, hash)
	d.seen[hash] = true

	if len(d.fifo) > d.size {
		old := d.fifo[0]
		d.fifo = d.fifo[1:]
		delete(d.seen, old)
	}
}

func readLoop(conn *websocket.Conn) error {
	dedup := NewDeduplicator(100)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		log.Printf("Received message: %s", message)

		// Handle incoming messages

		err = godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
		// Read Telegram configuration from environment variables
		botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		if botToken == "" {
			log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
		}

		// Read Telegram configuration from environment variables
		channelIDStr := os.Getenv("TELEGRAM_CHANNEL_ID")
		if channelIDStr == "" {
			log.Fatal("TELEGRAM_CHANNEL_ID environment variable not set")
		}

		chatId, err := strconv.ParseInt(channelIDStr, 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("JSON unmarshal error:", err)
			continue
		}
		log.Printf("ChannelName: %s", msg.ChannelName)
		// Process message based on channel name
		switch msg.ChannelName {
		case "#ping":
			messageThreadId, err := strconv.ParseInt(os.Getenv("MESSAGE_THREAD_ID_PINGS"), 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			sendMessageToTelegram(
				botToken,
				chatId,
				messageThreadId,
				fmt.Sprintf(
					"From %s\nMessage: %s\nRoute: %s\nSNR: %.1f",
					msg.SenderName,
					msg.Message,
					strings.Replace(msg.DisplayCombinedPath, "NA → ", "", -1),
					msg.Snr,
				),
			)
		case "Public", "Test":
			msgHash := msg.MessageHash

			if dedup.Exists(msgHash) {
				continue
			}
			dedup.Add(msgHash)
			messageThreadId, err := strconv.ParseInt(os.Getenv("MESSAGE_THREAD_ID_MESSAGES"), 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			sendMessageToTelegram(botToken, chatId, messageThreadId, fmt.Sprintf("From %s:\n%s", msg.SenderName, msg.Message))
		default:
			sendMessageToTelegram(botToken, chatId, 1, fmt.Sprintf("test: %s:\n%s", msg.SenderName, msg.Message))
			fmt.Printf("Received message on channel %s: %s\n", msg.ChannelName, msg.Message)
		}

		log.Println(string(message))
	}

	return nil
}

func main() {

	headers := http.Header{}

	headers.Set("User-Agent", "Mozilla/5.0")

	connect(headers)

}

func sendMessageToTelegram(botToken string, chatID int64, message_thread_id int64, text string) error {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		botToken,
	)
	log.Printf("url: %s", url)
	log.Printf("botToken: %s", botToken)

	payload := map[string]interface{}{
		"chat_id":           chatID,
		"message_thread_id": message_thread_id,
		"text":              text,
		"parse_mode":        "HTML",
	}

	body, err := json.Marshal(payload)

	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}

	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"telegram error: %s: %s",
			resp.Status,
			string(responseBody),
		)
	}

	fmt.Println("Status:", resp.Status)

	return nil
}
