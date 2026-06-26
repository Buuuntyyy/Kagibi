// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logger initialises the application-wide structured logger (log/slog).
//
// A single JSON handler writes to stdout. In development mode (GIN_MODE != "release")
// a human-readable text handler is used instead for readability.
//
// Calling Init() early in main() also redirects the stdlib log package so that any
// remaining log.Printf / log.Println call sites automatically emit JSON without
// requiring individual migration.
package logger

import (
	"log"
	"log/slog"
	"net"
	"os"
	"strings"
)

// Init initialises the global slog default logger and redirects the stdlib log
// package to route through it.  Must be called once at the top of main().
func Init() {
	level := parseLevel(os.Getenv("LOG_LEVEL"))

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var handler slog.Handler
	if os.Getenv("GIN_MODE") == "release" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Wrap with service metadata so every log line carries the Loki labels.
	base := slog.New(handler).With(
		"service", "kagibi-backend",
		"env", getEnv("APP_ENV", "development"),
	)

	slog.SetDefault(base)

	// Redirect stdlib log → slog so all legacy log.Printf calls emit JSON.
	log.SetFlags(0)
	log.SetOutput(newBridgeWriter(base))
}

// AnonymiseIP truncates the last octet of an IPv4 address or the last 80 bits of
// an IPv6 address before logging, per CNIL recommendation for application logs.
// Security logs (auth, rate-limit) receive the full IP via a separate field.
func AnonymiseIP(raw string) string {
	ip := net.ParseIP(raw)
	if ip == nil {
		return "invalid"
	}
	if v4 := ip.To4(); v4 != nil {
		return strings.Join([]string{
			strings.Split(raw, ".")[0],
			strings.Split(raw, ".")[1],
			strings.Split(raw, ".")[2],
			"x",
		}, ".")
	}
	// IPv6: keep first 48 bits (3 groups), zero the rest
	parts := strings.Split(raw, ":")
	if len(parts) >= 3 {
		return strings.Join(parts[:3], ":") + "::"
	}
	return "ipv6"
}

// parseLevel maps LOG_LEVEL env var to slog.Level (default: Info).
func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// bridgeWriter redirects stdlib log output to slog.Logger.
type bridgeWriter struct {
	logger *slog.Logger
}

func newBridgeWriter(l *slog.Logger) *bridgeWriter { return &bridgeWriter{logger: l} }

func (w *bridgeWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	w.logger.Info(msg)
	return len(p), nil
}
