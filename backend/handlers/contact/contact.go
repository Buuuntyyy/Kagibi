// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package contact

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

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

	// Sanitize: strip newlines from header-injectable fields.
	req.Name = sanitize(req.Name)
	req.Subject = sanitize(req.Subject)

	if err := sendMail(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'envoyer le message. Réessayez plus tard."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message envoyé avec succès"})
}

func sendMail(req ContactRequest) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	recipient := os.Getenv("CONTACT_RECIPIENT")

	if host == "" || user == "" || password == "" || recipient == "" {
		return fmt.Errorf("SMTP not configured")
	}
	if port == "" {
		port = "465"
	}

	body := buildBody(req)
	msg := []byte(
		"From: Kagibi Contact <" + user + ">\r\n" +
			"To: " + recipient + "\r\n" +
			"Reply-To: " + req.Email + "\r\n" +
			"Subject: [Contact] " + req.Subject + "\r\n" +
			"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)

	tlsCfg := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", host+":"+port, tlsCfg)
	if err != nil {
		return fmt.Errorf("SMTP dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(smtp.PlainAuth("", user, password, host)); err != nil {
		return fmt.Errorf("SMTP auth: %w", err)
	}
	if err := client.Mail(user); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}
	if _, err := wc.Write(msg); err != nil {
		return fmt.Errorf("SMTP write: %w", err)
	}
	return wc.Close()
}

func buildBody(req ContactRequest) string {
	return fmt.Sprintf(
		"Nouveau message de contact — Kagibi\n"+
			"=====================================\n\n"+
			"Nom    : %s\n"+
			"Email  : %s\n"+
			"Sujet  : %s\n\n"+
			"Message :\n%s\n",
		req.Name, req.Email, req.Subject, req.Message,
	)
}

// sanitize strips characters that could be used for email header injection.
func sanitize(s string) string {
	r := strings.NewReplacer("\r", "", "\n", "", "\t", " ")
	return strings.TrimSpace(r.Replace(s))
}
