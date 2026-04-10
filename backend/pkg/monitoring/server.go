// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitoring

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server représente le serveur HTTP de métriques Prometheus
type Server struct {
	httpServer *http.Server
	port       int
}

// StartSessionMonitor lance une goroutine qui met à jour la jauge ActiveUsers
// en comptant les clés Redis correspondant aux utilisateurs actifs
func StartSessionMonitor(redisClient *redis.Client) {
	if redisClient == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			// Compter les clés commençant par "active_user:"
			// Note: KEYS est bloquant sur de très grosses DB, mais ici on suppose une charge raisonnable.
			// Pour la prod à très haute échelle, SCAN est préférable, mais plus complexe à implémenter.
			keys, err := redisClient.Keys(ctx, "active_user:*").Result()
			cancel()

			if err != nil && err != redis.Nil {
				log.Printf("[Monitoring] Error counting active users: %v", err)
				continue
			}

			// Mettre à jour la jauge Prometheus
			// On filtre pour être sûr (même si KEYS avec pattern suffit)
			count := 0
			for _, k := range keys {
				if strings.HasPrefix(k, "active_user:") {
					count++
				}
			}

			ActiveUsers.Set(float64(count))
		}
	}()
}

// NewServer crée une nouvelle instance du serveur de métriques
func NewServer(port int) *Server {
	mux := http.NewServeMux()

	// Endpoint /metrics pour Prometheus
	mux.Handle("/metrics", promhttp.Handler())

	// Endpoint /health pour les health checks
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"kagibi-metrics"}`))
	})

	// Endpoint /ready pour les readiness checks
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		httpServer: srv,
		port:       port,
	}
}

// Start démarre le serveur de métriques de manière non-bloquante
func (s *Server) Start() error {
	go func() {
		fmt.Printf("Serveur de métriques Prometheus démarré sur le port %d\n", s.port)
		fmt.Printf("Métriques disponibles sur http://localhost:%d/metrics\n", s.port)
		fmt.Printf("Health check disponible sur http://localhost:%d/health\n", s.port)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Erreur du serveur de métriques: %v\n", err)
		}
	}()

	return nil
}

// Shutdown arrête gracieusement le serveur de métriques
func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("Arrêt du serveur de métriques...")
	return s.httpServer.Shutdown(ctx)
}

// GetPort retourne le port du serveur
func (s *Server) GetPort() int {
	return s.port
}
