// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"errors"
	"path/filepath"
	"strings"
)

// SecureJoin combine un chemin racine et un chemin utilisateur de manière sécurisée.
// Il empêche les attaques de type "Path Traversal" (../).
func SecureJoin(root, unsafePath string) (string, error) {
	// Nettoie le chemin (résout les .. et .)
	cleanPath := filepath.Join(root, unsafePath)

	// Obtient les chemins absolus pour une comparaison fiable
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", err
	}

	// Vérifie que le chemin final commence bien par la racine autorisée
	if !strings.HasPrefix(absPath, absRoot) {
		return "", errors.New("tentative de path traversal détectée")
	}

	return cleanPath, nil
}
