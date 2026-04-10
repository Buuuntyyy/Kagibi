// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package billing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// === Hooks pour événements utilisateur ===
// Ces fonctions sont appelées par le code applicatif

// HookUserRegistered est appelé quand un nouvel utilisateur s'inscrit
func HookUserRegistered(ctx context.Context, userID, email string) {
	provider := GetProvider()
	if provider == nil {
		return
	}

	event := UserCreatedEvent{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
		Timestamp: time.Now(),
	}

	go func() {
		if err := provider.OnUserCreated(context.Background(), event); err != nil {
			log.Printf("[Billing] Failed to notify user creation: %v", err)
		}
	}()
}

// HookUserDeleted est appelé quand un utilisateur supprime son compte
func HookUserDeleted(ctx context.Context, userID string) {
	provider := GetProvider()
	if provider == nil {
		return
	}

	go func() {
		if err := provider.OnUserDeleted(context.Background(), userID); err != nil {
			log.Printf("[Billing] Failed to notify user deletion: %v", err)
		}
	}()
}

// === Hooks pour événements de stockage ===

// HookFileUploaded est appelé après un upload réussi
func HookFileUploaded(ctx context.Context, userID string, fileSize int64, fileID string) {
	provider := GetProvider()
	if provider == nil {
		return
	}

	event := UsageEvent{
		UserID:         userID,
		EventType:      "storage_add",
		Bytes:          fileSize,
		Timestamp:      time.Now(),
		IdempotencyKey: fmt.Sprintf("upload_%s", fileID),
		Metadata: map[string]interface{}{
			"file_id": fileID,
		},
	}

	if err := provider.TrackUsage(ctx, event); err != nil {
		log.Printf("[Billing] Error tracking upload: %v", err)
	}
}

// HookFileDeleted est appelé après une suppression de fichier
func HookFileDeleted(ctx context.Context, userID string, fileSize int64, fileID string) {
	provider := GetProvider()
	if provider == nil {
		return
	}

	event := UsageEvent{
		UserID:         userID,
		EventType:      "storage_remove",
		Bytes:          fileSize,
		Timestamp:      time.Now(),
		IdempotencyKey: fmt.Sprintf("delete_%s", fileID),
		Metadata: map[string]interface{}{
			"file_id": fileID,
		},
	}

	if err := provider.TrackUsage(ctx, event); err != nil {
		log.Printf("[Billing] Error tracking deletion: %v", err)
	}
}

// HookFileDownloaded est appelé après un téléchargement
func HookFileDownloaded(ctx context.Context, userID string, fileSize int64, downloadID string) {
	provider := GetProvider()
	if provider == nil {
		return
	}

	// Générer un ID unique si non fourni
	if downloadID == "" {
		downloadID = uuid.New().String()
	}

	event := UsageEvent{
		UserID:         userID,
		EventType:      "bandwidth",
		Bytes:          fileSize,
		Timestamp:      time.Now(),
		IdempotencyKey: fmt.Sprintf("download_%s", downloadID),
	}

	if err := provider.TrackUsage(ctx, event); err != nil {
		log.Printf("[Billing] Error tracking download: %v", err)
	}
}

// HookP2PTransfer est appelé après un transfert P2P
func HookP2PTransfer(ctx context.Context, senderID, receiverID string, bytes int64, transferID string) {
	provider := GetProvider()
	if provider == nil {
		return
	}

	// Tracker pour l'expéditeur
	event := UsageEvent{
		UserID:         senderID,
		EventType:      "p2p_transfer",
		Bytes:          bytes,
		Timestamp:      time.Now(),
		IdempotencyKey: fmt.Sprintf("p2p_%s_sender", transferID),
		Metadata: map[string]interface{}{
			"transfer_id": transferID,
			"direction":   "outbound",
			"peer":        receiverID,
		},
	}

	if err := provider.TrackUsage(ctx, event); err != nil {
		log.Printf("[Billing] Error tracking P2P transfer (sender): %v", err)
	}
}

// === Fonctions de vérification des quotas ===

// CheckUploadAllowed vérifie si l'utilisateur peut uploader un fichier
func CheckUploadAllowed(ctx context.Context, userID string, fileSize int64) (bool, string) {
	provider := GetProvider()
	if provider == nil {
		// Pas de provider = tout est permis
		return true, ""
	}

	result, err := provider.CheckQuota(ctx, userID, fileSize)
	if err != nil {
		log.Printf("[Billing] Error checking quota: %v", err)
		// En cas d'erreur, on autorise (fail-open)
		return true, ""
	}

	return result.Allowed, result.Reason
}

// GetUserStorageLimit retourne la limite de stockage de l'utilisateur en bytes
func GetUserStorageLimit(ctx context.Context, userID string) int64 {
	provider := GetProvider()
	if provider == nil {
		// Limite par défaut: 5 Go
		return 5 * 1024 * 1024 * 1024
	}

	plan, err := provider.GetUserPlan(ctx, userID)
	if err != nil {
		log.Printf("[Billing] Error getting user plan: %v", err)
		return 5 * 1024 * 1024 * 1024
	}

	return plan.StorageLimitGB * 1024 * 1024 * 1024
}

// GetUserStorageUsed retourne l'espace utilisé par l'utilisateur en bytes
func GetUserStorageUsed(ctx context.Context, userID string) int64 {
	provider := GetProvider()
	if provider == nil {
		return 0
	}

	usage, err := provider.GetCurrentUsage(ctx, userID)
	if err != nil {
		log.Printf("[Billing] Error getting usage: %v", err)
		return 0
	}

	return int64(usage.StorageUsedGB * 1024 * 1024 * 1024)
}

// GetUserBandwidthLimit retourne la limite de bande passante mensuelle en bytes
// La bande passante n'est plus limitée par plan — retourne une valeur illimitée.
func GetUserBandwidthLimit(ctx context.Context, userID string) int64 {
	return -1 // illimité
}

// === Helpers pour souscriptions ===

// CreateOrUpdateSubscription crée ou met à jour une souscription avec idempotence
func CreateOrUpdateSubscription(ctx context.Context, userID, planCode, idempotencyKey string) (*Subscription, error) {
	provider := GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("billing provider not initialized")
	}

	// Essayer d'abord de récupérer la souscription existante
	existing, err := provider.GetSubscription(ctx, userID)
	if err == nil && existing != nil {
		// Mise à jour si le plan est différent
		if existing.PlanCode != planCode {
			return provider.UpdateSubscription(ctx, userID, planCode, idempotencyKey)
		}
		return existing, nil
	}

	// Créer une nouvelle souscription
	return provider.CreateSubscription(ctx, userID, planCode, idempotencyKey)
}

// CancelUserSubscription annule la souscription d'un utilisateur
func CancelUserSubscription(ctx context.Context, userID, idempotencyKey string) error {
	provider := GetProvider()
	if provider == nil {
		return fmt.Errorf("billing provider not initialized")
	}

	return provider.CancelSubscription(ctx, userID, idempotencyKey)
}

// === Quota Checks supplémentaires ===

// CheckP2PAllowed vérifie si l'utilisateur peut effectuer un transfert P2P
func CheckP2PAllowed(ctx context.Context, userID string, fileSize int64) (bool, string) {
	provider := GetProvider()
	if provider == nil {
		return true, ""
	}

	plan, err := provider.GetUserPlan(ctx, userID)
	if err != nil {
		return true, "" // fail-open
	}

	p2pLimitGB, ok := plan.Features["p2p_limit_gb"]
	if !ok {
		return true, ""
	}

	limitGB, _ := p2pLimitGB.(float64)
	if limitGB < 0 {
		return true, "" // -1 = illimité
	}

	// Vérification basique sur la taille du fichier individuel
	// La vérification détaillée du cumul mensuel est faite par le service billing privé
	limitBytes := int64(limitGB * 1024 * 1024 * 1024)
	if fileSize > limitBytes {
		return false, fmt.Sprintf("Fichier trop volumineux pour le transfert P2P. Limite: %.0f Go", limitGB)
	}

	return true, ""
}

// CheckFileSizeAllowed vérifie la taille max par fichier selon le plan
func CheckFileSizeAllowed(ctx context.Context, userID string, fileSize int64) (bool, string) {
	provider := GetProvider()
	if provider == nil {
		return true, ""
	}

	plan, err := provider.GetUserPlan(ctx, userID)
	if err != nil {
		return true, "" // fail-open
	}

	maxFileSizeMB, ok := plan.Features["max_file_size_mb"]
	if !ok {
		return true, ""
	}

	limitMB, _ := maxFileSizeMB.(float64)
	if limitMB < 0 {
		return true, "" // -1 = illimité
	}

	limitBytes := int64(limitMB * 1024 * 1024)
	if fileSize > limitBytes {
		return false, fmt.Sprintf("Fichier trop volumineux. Limite: %.0f Mo pour le plan %s", limitMB, plan.Name)
	}

	return true, ""
}
