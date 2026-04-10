// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package contact

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"kagibi/backend/pkg/mailer"

	"github.com/gin-gonic/gin"
)

type ContactRequest struct {
	Name    string `json:"name"    binding:"required,min=2,max=100"`
	Email   string `json:"email"   binding:"required,email"`
	Subject string `json:"subject" binding:"required,min=3,max=150"`
	Message string `json:"message" binding:"required,min=10,max=4000"`
}

// Handler handles POST /api/v1/contact (public, rate-limited by middleware).
func Handler(c *gin.Context) {
	var req ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Champs invalides ou manquants"})
		return
	}

	req.Name = sanitize(req.Name)
	req.Subject = sanitize(req.Subject)

	// CONTACT_RECIPIENT overrides the default From address as delivery target.
	to := os.Getenv("CONTACT_RECIPIENT")
	if to == "" {
		to = os.Getenv("MAIL_FROM_ADDRESS")
	}

	if err := mailer.Send(mailer.Message{
		To:      to,
		Subject: "[Contact] " + req.Subject,
		Body:    buildBody(req),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'envoyer le message. Réessayez plus tard."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message envoyé avec succès"})
}

func buildBody(req ContactRequest) string {
	return fmt.Sprintf(
		"Nouveau message de contact — Kagibi\n"+
			"=====================================\n\n"+
			"Nom     : %s\n"+
			"Email   : %s\n"+
			"Sujet   : %s\n\n"+
			"Message :\n%s\n",
		req.Name, req.Email, req.Subject, req.Message,
	)
}

func sanitize(s string) string {
	return strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "", "\t", " ").Replace(s))
}
