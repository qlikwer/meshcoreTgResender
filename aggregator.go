package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

//
// ===== PUBLIC AGGREGATOR =====
//

type PublicEntry struct {
	Msg   Message
	Timer *time.Timer
}

type PublicAggregator struct {
	mu      sync.Mutex
	pending map[string]*PublicEntry

	cfg    *Config
	sender *TelegramSender
}

func NewPublicAggregator(cfg *Config, sender *TelegramSender) *PublicAggregator {
	return &PublicAggregator{
		pending: make(map[string]*PublicEntry),
		cfg:     cfg,
		sender:  sender,
	}
}

func (a *PublicAggregator) Add(msg Message) {
	a.mu.Lock()
	defer a.mu.Unlock()

	h := msg.MessageHash

	p, ok := a.pending[h]
	if !ok {
		p = &PublicEntry{
			Msg: msg,
		}
		a.pending[h] = p
	}

	// обновляем текст (на случай если пришел поздний update)
	p.Msg = msg

	if p.Timer != nil {
		p.Timer.Stop()
	}

	p.Timer = time.AfterFunc(2*time.Second, func() {
		a.flush(h)
	})
}

func (a *PublicAggregator) flush(hash string) {
	a.mu.Lock()
	p, ok := a.pending[hash]
	if !ok {
		a.mu.Unlock()
		return
	}
	delete(a.pending, hash)
	a.mu.Unlock()

	text := fmt.Sprintf(
		"From %s:\n%s",
		p.Msg.SenderName,
		p.Msg.Message,
	)

	a.sender.SendPublic(text)
}

//
// ===== PING AGGREGATOR =====
//

type PingRoute struct {
	Path string
	SNR  float64
}

type PingEntry struct {
	Msg    Message
	Routes map[string]PingRoute
	Timer  *time.Timer
}

type PingAggregator struct {
	mu      sync.Mutex
	pending map[string]*PingEntry

	cfg    *Config
	sender *TelegramSender
}

func NewPingAggregator(cfg *Config, sender *TelegramSender) *PingAggregator {
	return &PingAggregator{
		pending: make(map[string]*PingEntry),
		cfg:     cfg,
		sender:  sender,
	}
}

func (a *PingAggregator) Add(msg Message) {
	a.mu.Lock()
	defer a.mu.Unlock()

	h := msg.MessageHash

	p, ok := a.pending[h]
	if !ok {
		p = &PingEntry{
			Msg:    msg,
			Routes: make(map[string]PingRoute),
		}
		a.pending[h] = p
	}

	// всегда сохраняем последнее состояние сообщения
	p.Msg = msg

	// ключ дедупликации маршрута: по самому пути, а не по SNR
	// (SNR одного и того же физического маршрута слегка дрожит от пакета
	// к пакету, поэтому раньше один маршрут считался "разными").
	key := msg.DisplayCombinedPath

	p.Routes[key] = PingRoute{
		Path: msg.DisplayCombinedPath,
		SNR:  msg.Snr,
	}

	if p.Timer != nil {
		p.Timer.Stop()
	}

	p.Timer = time.AfterFunc(2*time.Second, func() {
		a.flush(h)
	})
}

func (a *PingAggregator) flush(hash string) {
	a.mu.Lock()
	p, ok := a.pending[hash]
	if !ok {
		a.mu.Unlock()
		return
	}
	delete(a.pending, hash)
	a.mu.Unlock()

	var b strings.Builder

	b.WriteString("📡 PING\n\n")
	b.WriteString(fmt.Sprintf("From: %s\n\n", p.Msg.SenderName))
	b.WriteString("Message:\n")
	b.WriteString(p.Msg.Message)
	b.WriteString("\n\nRoutes:\n")

	i := 1
	for _, r := range p.Routes {
		b.WriteString(fmt.Sprintf(
			"%d) SNR: %.1f\n%s\n\n",
			i,
			r.SNR,
			strings.ReplaceAll(r.Path, "NA → ", ""),
		))
		i++
	}

	a.sender.SendPing(b.String())
}
