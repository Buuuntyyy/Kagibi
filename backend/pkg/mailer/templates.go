// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package mailer

import "fmt"

// SendWelcome sends a welcome email to a newly registered user.
// Runs asynchronously — caller does not need to wait.
func SendWelcome(toEmail, name string) {
	go func() {
		err := Send(Message{
			To:      toEmail,
			Subject: "Bienvenue sur Kagibi !",
			Body:    welcomeBody(name),
		})
		if err != nil {
			// Non-fatal: registration succeeded, email is best-effort.
			// The error is logged by the caller if needed.
			_ = err
		}
	}()
}

func welcomeBody(name string) string {
	return fmt.Sprintf(`Bonjour %s,

Bienvenue sur Kagibi — le cloud 100% opensource où personne ne peut lire vos données.

Votre compte est prêt. Voici ce que vous pouvez faire dès maintenant :

  • Stocker vos fichiers chiffrés (AES-256-GCM, clés uniquement sur votre appareil)
  • Partager des fichiers via liens chiffrés
  • Transférer des fichiers jusqu'à 10 Go en P2P direct, sans transit sur nos serveurs

Nous ne pouvons pas lire vos fichiers — c'est une garantie mathématique, pas une promesse.

Accéder à votre Drive : https://kagibi.cloud/dashboard

---
Une question ? Rendez-vous sur github : https://github.com/Buuuntyyy/kagibi/issues.

L'équipe Kagibi
https://kagibi.cloud

---
Vous recevez cet email car vous venez de créer un compte sur Kagibi.
`, name)
}
