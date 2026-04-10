// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package mailer sends transactional emails via SMTP.
// Configuration is read from the MAIL_* environment variables:
//
//	MAIL_HOST         SMTP server hostname
//	MAIL_PORT         SMTP port (587 = STARTTLS, 465 = SSL direct)
//	MAIL_USERNAME     SMTP login
//	MAIL_PASSWORD     SMTP password
//	MAIL_ENCRYPTION   "tls" (STARTTLS/587) | "ssl" (direct TLS/465)
//	MAIL_FROM_ADDRESS Sender address  (e.g. no-reply@kagibi.cloud)
//	MAIL_FROM_NAME    Sender display name (optional, defaults to "Kagibi")
package mailer

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// Message holds the data for a single outbound email.
type Message struct {
	To      string
	Subject string
	Body    string // plain-text
}

// Send delivers msg using the MAIL_* environment variables.
func Send(msg Message) error {
	cfg, err := configFromEnv()
	if err != nil {
		return err
	}
	return cfg.send(msg)
}

type config struct {
	host       string
	port       string
	username   string
	password   string
	encryption string // "tls" (STARTTLS) | "ssl" | ""
	fromAddr   string
	fromName   string
}

func configFromEnv() (*config, error) {
	host := os.Getenv("MAIL_HOST")
	port := os.Getenv("MAIL_PORT")
	user := os.Getenv("MAIL_USERNAME")
	pass := os.Getenv("MAIL_PASSWORD")
	enc := strings.ToLower(os.Getenv("MAIL_ENCRYPTION"))
	from := os.Getenv("MAIL_FROM_ADDRESS")
	name := os.Getenv("MAIL_FROM_NAME")

	if host == "" || user == "" || pass == "" || from == "" {
		return nil, fmt.Errorf("mailer: MAIL_HOST, MAIL_USERNAME, MAIL_PASSWORD and MAIL_FROM_ADDRESS must be set")
	}
	if port == "" {
		if enc == "ssl" {
			port = "465"
		} else {
			port = "587"
		}
	}
	if name == "" {
		name = "Kagibi"
	}
	return &config{
		host: host, port: port, username: user, password: pass,
		encryption: enc, fromAddr: from, fromName: name,
	}, nil
}

func (c *config) send(msg Message) error {
	var client *smtp.Client
	var err error

	addr := c.host + ":" + c.port
	auth := smtp.PlainAuth("", c.username, c.password, c.host)
	tlsCfg := &tls.Config{ServerName: c.host}

	if c.encryption == "ssl" {
		// Port 465 — direct TLS connection.
		conn, dialErr := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, tlsCfg)
		if dialErr != nil {
			return fmt.Errorf("mailer: SSL dial: %w", dialErr)
		}
		client, err = smtp.NewClient(conn, c.host)
	} else {
		// Port 587 — plain connection then STARTTLS upgrade.
		conn, dialErr := net.DialTimeout("tcp", addr, 10*time.Second)
		if dialErr != nil {
			return fmt.Errorf("mailer: dial: %w", dialErr)
		}
		client, err = smtp.NewClient(conn, c.host)
		if err == nil {
			err = client.StartTLS(tlsCfg)
		}
	}
	if err != nil {
		return fmt.Errorf("mailer: SMTP setup: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("mailer: auth: %w", err)
	}
	if err := client.Mail(c.fromAddr); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	if err := client.Rcpt(msg.To); err != nil {
		return fmt.Errorf("mailer: RCPT TO: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err := io.WriteString(wc, c.formatRaw(msg)); err != nil {
		return fmt.Errorf("mailer: write: %w", err)
	}
	return wc.Close()
}

func (c *config) formatRaw(msg Message) string {
	from := fmt.Sprintf("%s <%s>", c.fromName, c.fromAddr)
	return strings.Join([]string{
		"From: " + from,
		"To: " + msg.To,
		"Subject: " + sanitize(msg.Subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		msg.Body,
	}, "\r\n")
}

func sanitize(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}
