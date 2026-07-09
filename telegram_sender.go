package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TelegramSender struct {
	cfg    *Config
	client *http.Client
}

func NewTelegramSender(cfg *Config) *TelegramSender {
	return &TelegramSender{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

//
// ===== BASE SEND =====
//

func (s *TelegramSender) send(text string, threadID int64) error {
	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		s.cfg.BotToken,
	)

	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", s.cfg.ChatID))
	data.Set("text", text)
	data.Set("parse_mode", "HTML")
	data.Set("disable_web_page_preview", "true")

	// thread (topics in supergroup)
	if threadID != 0 {
		data.Set("message_thread_id", fmt.Sprintf("%d", threadID))
	}

	resp, err := s.client.PostForm(endpoint, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram API error: %s", resp.Status)
	}

	return nil
}

//
// ===== PUBLIC MESSAGE =====
//

func (s *TelegramSender) SendPublic(text string) {
	err := s.send(text, s.cfg.MessageThreadID)
	if err != nil {
		fmt.Printf("Telegram SendPublic error: %v\n", err)
	}
}

//
// ===== PING MESSAGE =====
//

func (s *TelegramSender) SendPing(text string) {
	err := s.send(text, s.cfg.PingThreadID)
	if err != nil {
		fmt.Printf("Telegram SendPing error: %v\n", err)
	}
}

//
// ===== HELPERS =====
//

// htmlEscape экранирует текст для отправки с parse_mode=HTML. Обязателен
// для любых динамических данных (имя отправителя, текст сообщения), иначе
// символы <, >, & могут сломать разметку и Telegram вернёт 400 Bad Request.
func htmlEscape(text string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(text)
}
