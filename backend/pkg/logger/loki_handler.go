// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logger — loki_handler.go implémente un slog.Handler qui pousse les
// entrées de log vers l'API HTTP de Loki (/loki/api/v1/push) en mode batch.
//
// Activation : définir la variable d'environnement LOKI_URL.
// Exemple    : LOKI_URL=http://loki.monitoring.svc.cluster.local:3100
//
// Le handler est non-bloquant : les entrées sont placées dans un canal bufférisé
// et envoyées par une goroutine dédiée. Si Loki est indisponible, les entrées
// sont simplement abandonnées sans affecter l'application.
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	lokiBatchSize    = 100             // flush si le batch atteint N entrées
	lokiFlushPeriod  = 2 * time.Second // flush toutes les N secondes
	lokiChannelSize  = 4096            // buffer du canal d'envoi
	lokiSendTimeout  = 5 * time.Second // timeout HTTP
)

// lokiEntry représente une entrée de log prête à être envoyée.
type lokiEntry struct {
	ts     time.Time
	labels map[string]string
	line   string
}

// LokiHandler est un slog.Handler qui envoie les logs à Loki.
// ch et client sont partagés entre les copies issues de WithAttrs/WithGroup.
type LokiHandler struct {
	url     string
	client  *http.Client
	ch      chan lokiEntry
	attrs   []slog.Attr
	group   string
	service string
	env     string
}

// newLokiHandler crée un LokiHandler et démarre la goroutine d'envoi.
// Retourne nil si LOKI_URL n'est pas définie.
func newLokiHandler() *LokiHandler {
	url := os.Getenv("LOKI_URL")
	if url == "" {
		return nil
	}
	h := &LokiHandler{
		url:     url + "/loki/api/v1/push",
		client:  &http.Client{Timeout: lokiSendTimeout},
		ch:      make(chan lokiEntry, lokiChannelSize),
		service: getEnv("APP_NAME", "kagibi-backend"),
		env:     getEnv("APP_ENV", "production"),
	}
	go h.run()
	return h
}

// Enabled implémente slog.Handler — tous les niveaux sont acceptés.
func (h *LokiHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

// Handle implémente slog.Handler — met l'entrée en file d'envoi.
func (h *LokiHandler) Handle(_ context.Context, r slog.Record) error {
	// Construire la ligne JSON
	fields := map[string]any{
		"msg":   r.Message,
		"level": r.Level.String(),
		"time":  r.Time.UTC().Format(time.RFC3339Nano),
	}
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}
	line, err := json.Marshal(fields)
	if err != nil {
		return nil
	}

	labels := map[string]string{
		"service": h.service,
		"env":     h.env,
		"level":   r.Level.String(),
	}
	// Extraire component et event_type comme labels Loki si présents
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "component" || a.Key == "event_type" {
			labels[a.Key] = a.Value.String()
		}
		return true
	})

	select {
	case h.ch <- lokiEntry{ts: r.Time, labels: labels, line: string(line)}:
	default:
		// Canal plein : entrée abandonnée silencieusement (Loki lent ou indisponible)
	}
	return nil
}

// WithAttrs implémente slog.Handler.
func (h *LokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	h2.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &h2
}

// WithGroup implémente slog.Handler.
func (h *LokiHandler) WithGroup(name string) slog.Handler {
	h2 := *h
	h2.group = name
	return &h2
}

// run est la goroutine d'envoi qui batch les entrées et les pousse à Loki.
func (h *LokiHandler) run() {
	ticker := time.NewTicker(lokiFlushPeriod)
	defer ticker.Stop()

	var batch []lokiEntry
	for {
		select {
		case entry := <-h.ch:
			batch = append(batch, entry)
			if len(batch) >= lokiBatchSize {
				h.flush(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				h.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

// flush envoie un batch d'entrées à Loki.
func (h *LokiHandler) flush(entries []lokiEntry) {
	// Grouper les entrées par labels identiques
	type streamKey string
	streams := make(map[streamKey]struct {
		labels map[string]string
		values [][2]string
	})

	for _, e := range entries {
		key := streamKey(fmt.Sprintf("%v", e.labels))
		s := streams[key]
		s.labels = e.labels
		s.values = append(s.values, [2]string{
			strconv.FormatInt(e.ts.UnixNano(), 10),
			e.line,
		})
		streams[key] = s
	}

	// Construire le payload Loki
	type lokiStream struct {
		Stream map[string]string `json:"stream"`
		Values [][2]string       `json:"values"`
	}
	type lokiPush struct {
		Streams []lokiStream `json:"streams"`
	}

	push := lokiPush{}
	for _, s := range streams {
		push.Streams = append(push.Streams, lokiStream{
			Stream: s.labels,
			Values: s.values,
		})
	}

	body, err := json.Marshal(push)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Authentification optionnelle (Grafana Cloud)
	if user := os.Getenv("LOKI_USER"); user != "" {
		req.SetBasicAuth(user, os.Getenv("LOKI_TOKEN"))
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return // Loki indisponible — abandon silencieux
	}
	resp.Body.Close()
}
