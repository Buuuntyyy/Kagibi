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

// SendP2PInvite notifies a recipient that someone wants to send them a file.
// lang must be "fr" or "en"; any other value falls back to "fr".
func SendP2PInvite(toEmail, senderName, fileName string, fileSize int64, token, lang string) error {
	var subject, body string
	if lang == "en" {
		subject = senderName + " wants to share a file with you"
		body = p2pInviteBodyEN(toEmail, senderName, fileName, fileSize, token)
	} else {
		subject = senderName + " souhaite vous envoyer un fichier"
		body = p2pInviteBodyFR(toEmail, senderName, fileName, fileSize, token)
	}
	return Send(Message{
		To:      toEmail,
		Subject: subject,
		Body:    body,
	})
}

func p2pInviteBodyFR(toEmail, senderName, fileName string, fileSize int64, token string) string {
	sizeStr := formatBytes(fileSize)
	link := "https://send.kagibi.cloud/?invite=" + token
	return "Bonjour,\n\n" +
		senderName + " souhaite vous transférer un fichier de façon sécurisée via Kagibi.\n\n" +
		"  Fichier : " + fileName + " (" + sizeStr + ")\n\n" +
		"Pour le recevoir, cliquez sur le lien ci-dessous. Aucun compte n'est nécessaire.\n\n" +
		"  " + link + "\n\n" +
		"Ce lien est valable 24 heures. Une fois ouvert, " + senderName + " devra être connecté\n" +
		"pour démarrer le transfert — le fichier est envoyé en P2P chiffré de bout en bout\n" +
		"directement entre vos appareils, ou via un relais Kagibi si votre réseau l'exige.\n" +
		"Dans tous les cas, le contenu reste illisible pour nos serveurs.\n\n" +
		"Si vous ne vous attendiez pas à recevoir ce message, vous pouvez l'ignorer sans\n" +
		"vous inquiéter — aucune action n'est requise de votre part.\n\n" +
		"—\n" +
		"L'équipe Kagibi · https://kagibi.cloud\n\n" +
		"Vous recevez cet e-mail car " + senderName + " a utilisé votre adresse (" + toEmail + ")\n" +
		"pour vous envoyer une invitation de transfert.\n"
}

func p2pInviteBodyEN(toEmail, senderName, fileName string, fileSize int64, token string) string {
	sizeStr := formatBytes(fileSize)
	link := "https://send.kagibi.cloud/?invite=" + token
	return "Hello,\n\n" +
		senderName + " would like to send you a file securely through Kagibi.\n\n" +
		"  File: " + fileName + " (" + sizeStr + ")\n\n" +
		"Click the link below to receive it. No account required.\n\n" +
		"  " + link + "\n\n" +
		"This link expires in 24 hours. Once you open it, " + senderName + " will need to be online\n" +
		"to start the transfer — the file is sent end-to-end encrypted, directly between your\n" +
		"devices or via a Kagibi relay if your network requires it.\n" +
		"Either way, its contents remain unreadable to our servers.\n\n" +
		"If you weren't expecting this message, feel free to ignore it —\n" +
		"no action is required on your part.\n\n" +
		"—\n" +
		"The Kagibi team · https://kagibi.cloud\n\n" +
		"You received this email because " + senderName + " used your address (" + toEmail + ")\n" +
		"to send you a file transfer invitation.\n"
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
