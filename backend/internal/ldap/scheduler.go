// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package ldap

import (
	"context"
	"log"
	"sync"
	"time"

	"kagibi/backend/pkg"

	"github.com/uptrace/bun"
)

// Scheduler manages periodic LDAP sync goroutines, one per organisation.
// It is started once at application boot and stopped on graceful shutdown.
type Scheduler struct {
	db      *bun.DB
	mu      sync.Mutex
	tickers map[int64]*time.Ticker
	stops   map[int64]chan struct{}
}

var globalScheduler *Scheduler

// Start initialises the global scheduler, loads all enabled configs, and
// launches their sync tickers. It also starts a watcher that picks up newly
// enabled configs every 5 minutes.
func Start(db *bun.DB) {
	globalScheduler = &Scheduler{
		db:      db,
		tickers: make(map[int64]*time.Ticker),
		stops:   make(map[int64]chan struct{}),
	}
	globalScheduler.loadAll()
	go globalScheduler.watch()
}

// Stop gracefully terminates all running sync goroutines.
func Stop() {
	if globalScheduler != nil {
		globalScheduler.stopAll()
	}
}

// Reload reloads the schedule for a single org (called after config save).
func Reload(orgID int64) {
	if globalScheduler != nil {
		globalScheduler.reload(orgID)
	}
}

func (sc *Scheduler) loadAll() {
	ctx := context.Background()
	var cfgs []pkg.OrgLDAPConfig
	if err := sc.db.NewSelect().Model(&cfgs).
		Where("enabled = true").
		Scan(ctx); err != nil {
		log.Printf("[ldap scheduler] failed to load configs: %v", err)
		return
	}
	for i := range cfgs {
		sc.schedule(&cfgs[i])
	}
}

func (sc *Scheduler) watch() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		sc.loadAll()
	}
}

func (sc *Scheduler) schedule(cfg *pkg.OrgLDAPConfig) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Unregister any existing goroutine for this org.
	if stop, ok := sc.stops[cfg.OrgID]; ok {
		close(stop)
		delete(sc.stops, cfg.OrgID)
	}
	if t, ok := sc.tickers[cfg.OrgID]; ok {
		t.Stop()
		delete(sc.tickers, cfg.OrgID)
	}

	interval := time.Duration(cfg.SyncIntervalMinutes) * time.Minute
	if interval < time.Minute {
		interval = time.Minute
	}

	stop := make(chan struct{})
	ticker := time.NewTicker(interval)
	sc.stops[cfg.OrgID] = stop
	sc.tickers[cfg.OrgID] = ticker

	cfgCopy := *cfg
	go func() {
		log.Printf("[ldap scheduler] org=%d started (interval=%v)", cfgCopy.OrgID, interval)
		for {
			select {
			case <-ticker.C:
				NewSyncer(sc.db, &cfgCopy).Run(context.Background())
			case <-stop:
				log.Printf("[ldap scheduler] org=%d stopped", cfgCopy.OrgID)
				return
			}
		}
	}()
}

func (sc *Scheduler) reload(orgID int64) {
	ctx := context.Background()
	var cfg pkg.OrgLDAPConfig
	if err := sc.db.NewSelect().Model(&cfg).Where("org_id = ?", orgID).Scan(ctx); err != nil {
		sc.unschedule(orgID)
		return
	}
	if !cfg.Enabled {
		sc.unschedule(orgID)
		return
	}
	sc.schedule(&cfg)
}

func (sc *Scheduler) unschedule(orgID int64) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if stop, ok := sc.stops[orgID]; ok {
		close(stop)
		delete(sc.stops, orgID)
	}
	if t, ok := sc.tickers[orgID]; ok {
		t.Stop()
		delete(sc.tickers, orgID)
	}
}

func (sc *Scheduler) stopAll() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	for orgID, stop := range sc.stops {
		close(stop)
		delete(sc.stops, orgID)
	}
	for orgID, t := range sc.tickers {
		t.Stop()
		delete(sc.tickers, orgID)
	}
}
