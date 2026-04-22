// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package mailer

import (
	"fmt"
	"log"
)

// SendWelcome sends a welcome email to a newly registered user.
// Runs asynchronously — caller does not need to wait.
func SendWelcome(toEmail, name string) {
	go func() {
		if err := Send(Message{
			To:      toEmail,
			Subject: "Bienvenue sur Kagibi !",
			Body:    welcomeBody(name),
		}); err != nil {
			log.Printf("[Mailer] Welcome email failed for %s: %v", toEmail, err)
		} else {
			log.Printf("[Mailer] Welcome email sent to %s", toEmail)
		}
	}()
}

// SendP2PInvite notifies a registered user that someone wants to send them a file.
func SendP2PInvite(toEmail, senderName, fileName string, fileSize int64, token string) error {
	return Send(Message{
		To:      toEmail,
		Subject: senderName + " wants to send you a file on Kagibi",
		Body:    p2pInviteBody(toEmail, senderName, fileName, fileSize, token),
	})
}

func p2pInviteBody(toEmail, senderName, fileName string, fileSize int64, token string) string {
	sizeStr := formatBytes(fileSize)
	link := "https://send.kagibi.cloud/?invite=" + token
	return "Hello,\n\n" +
		senderName + " wants to send you a file securely via Kagibi:\n\n" +
		"  File: " + fileName + " (" + sizeStr + ")\n\n" +
		"Click the link below to receive it. You will need to be logged in to your Kagibi account:\n\n" +
		"  " + link + "\n\n" +
		"This link expires in 24 hours. Once you open it, " + senderName + " will need to be online\n" +
		"to start the transfer — the file is sent directly between your browsers (P2P).\n\n" +
		"Kagibi never sees the content of your files.\n\n" +
		"---\n" +
		"The Kagibi team\n" +
		"https://kagibi.cloud\n\n" +
		"---\n" +
		"You received this email because " + senderName + " used your address (" + toEmail + ") to send a file invite.\n" +
		"If you were not expecting this, you can safely ignore it.\n"
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func welcomeBody(name string) string {
	return "Bonjour " + name + ",\n\n" +
		"Bienvenue sur Kagibi — le cloud 100% opensource où personne ne peut lire vos données.\n\n" +
		"Votre compte est prêt. Voici ce que vous pouvez faire dès maintenant :\n\n" +
		"  - Stocker vos fichiers chiffrés (AES-256-GCM, clés uniquement sur votre appareil)\n" +
		"  - Partager des fichiers via liens chiffrés\n" +
		"  - Transférer des fichiers jusqu'à 10 Go en P2P direct, sans stockage sur nos serveurs\n\n" +
		"Nous ne pouvons pas lire vos fichiers — c'est une garantie mathématique, pas une promesse.\n\n" +
		"Accéder à votre Drive : https://kagibi.cloud/dashboard\n\n" +
		"---\n" +
		"Une question ? Rendez-vous sur GitHub : https://github.com/Buuuntyyy/kagibi/issues.\n\n" +
		"L'équipe Kagibi\n" +
		"https://kagibi.cloud\n\n" +
		"---\n" +
		"Vous recevez cet email car vous venez de créer un compte sur Kagibi.\n"
}
