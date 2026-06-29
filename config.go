package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken        string
	ChatID          int64
	PingThreadID    int64
	MessageThreadID int64
}

func LoadConfig() (*Config, error) {
	// .env не обязателен: переменные могут быть заданы окружением
	// (Docker/systemd и т.п.), поэтому ошибку только логируем.
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env не загружен (%v), использую переменные окружения", err)
	}

	cfg := &Config{}

	cfg.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is not set")
	}

	var err error

	cfg.ChatID, err = strconv.ParseInt(os.Getenv("TELEGRAM_CHANNEL_ID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TELEGRAM_CHANNEL_ID: %w", err)
	}

	cfg.PingThreadID, err = strconv.ParseInt(os.Getenv("MESSAGE_THREAD_ID_PINGS"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid MESSAGE_THREAD_ID_PINGS: %w", err)
	}

	cfg.MessageThreadID, err = strconv.ParseInt(os.Getenv("MESSAGE_THREAD_ID_MESSAGES"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid MESSAGE_THREAD_ID_MESSAGES: %w", err)
	}

	return cfg, nil
}
