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
	"github.com/uptrace/bun"
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

// StartDBMonitor lance une goroutine qui met à jour les jauges métier
// depuis PostgreSQL toutes les minutes.
func StartDBMonitor(db *bun.DB) {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		update := func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Nombre total d'utilisateurs
			count, err := db.NewSelect().TableExpr("profiles").Count(ctx)
			if err != nil {
				log.Printf("[Monitoring] Failed to count users: %v", err)
			} else {
				TotalUsersGauge.Set(float64(count))
			}

			// Stockage total utilisé (personnel)
			var total struct{ Sum int64 }
			err = db.NewSelect().
				TableExpr("user_plans").
				ColumnExpr("COALESCE(SUM(storage_used), 0) AS sum").
				Scan(ctx, &total)
			if err != nil {
				log.Printf("[Monitoring] Failed to sum storage: %v", err)
			} else {
				TotalStorageUsedBytes.Set(float64(total.Sum))
			}

			// Nombre total d'organisations actives
			orgCount, err := db.NewSelect().TableExpr("organizations").
				Where("deleted_at IS NULL").Count(ctx)
			if err != nil {
				log.Printf("[Monitoring] Failed to count orgs: %v", err)
			} else {
				OrgsTotal.Set(float64(orgCount))
			}

			// Nombre total d'appartenances org
			memberCount, err := db.NewSelect().TableExpr("org_members").Count(ctx)
			if err != nil {
				log.Printf("[Monitoring] Failed to count org members: %v", err)
			} else {
				OrgMembersTotal.Set(float64(memberCount))
			}

			// Stockage total utilisé par les organisations
			var orgStorage struct{ Sum int64 }
			if err := db.NewSelect().TableExpr("organizations").
				ColumnExpr("COALESCE(SUM(storage_used_bytes), 0) AS sum").
				Where("deleted_at IS NULL").
				Scan(ctx, &orgStorage); err != nil {
				log.Printf("[Monitoring] Failed to sum org storage: %v", err)
			} else {
				OrgStorageUsedBytes.Set(float64(orgStorage.Sum))
			}

			// Nombre de fichiers dans les organisations
			orgFileCount, err := db.NewSelect().TableExpr("org_files").
				Where("deleted_at IS NULL").Count(ctx)
			if err != nil {
				log.Printf("[Monitoring] Failed to count org files: %v", err)
			} else {
				OrgFilesTotal.Set(float64(orgFileCount))
			}

			// Nombre de dossiers dans les organisations
			orgFolderCount, err := db.NewSelect().TableExpr("org_folders").
				Where("deleted_at IS NULL").Count(ctx)
			if err != nil {
				log.Printf("[Monitoring] Failed to count org folders: %v", err)
			} else {
				OrgFoldersTotal.Set(float64(orgFolderCount))
			}

			// Invitations en attente
			invCount, err := db.NewSelect().TableExpr("org_invitations").
				Where("status = 'active'").Count(ctx)
			if err != nil {
				log.Printf("[Monitoring] Failed to count org invitations: %v", err)
			} else {
				OrgInvitationsPendingTotal.Set(float64(invCount))
			}

			// Médiane du nombre de membres par organisation
			var medPerOrg struct{ Median float64 }
			if err := db.NewRaw(`
				SELECT COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY cnt), 0) AS median
				FROM (SELECT org_id, COUNT(*) AS cnt FROM org_members GROUP BY org_id) sub
			`).Scan(ctx, &medPerOrg); err != nil {
				log.Printf("[Monitoring] Failed to compute median members per org: %v", err)
			} else {
				OrgMembersPerOrgMedian.Set(medPerOrg.Median)
			}

			// Médiane du nombre d'organisations par membre
			var medPerMember struct{ Median float64 }
			if err := db.NewRaw(`
				SELECT COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY cnt), 0) AS median
				FROM (SELECT user_id, COUNT(*) AS cnt FROM org_members GROUP BY user_id) sub
			`).Scan(ctx, &medPerMember); err != nil {
				log.Printf("[Monitoring] Failed to compute median orgs per member: %v", err)
			} else {
				OrgOrgsPerMemberMedian.Set(medPerMember.Median)
			}
		}

		update() // première collecte immédiate au démarrage
		for range ticker.C {
			update()
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
