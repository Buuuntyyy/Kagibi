// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package mailer

import (
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
