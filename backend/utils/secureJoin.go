// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"errors"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// SanitizeVirtualPath normalise un chemin logique fourni par l'utilisateur.
// Décode itérativement les séquences URL (%2e%2e, %252e%252e…) avant toute
// vérification, puis rejette tout chemin contenant un composant "..".
func SanitizeVirtualPath(inputPath string) (string, error) {
	decoded := inputPath
	for {
		next, err := url.PathUnescape(decoded)
		if err != nil {
			return "", errors.New("encodage de chemin invalide")
		}
		if next == decoded {
			break
		}
		decoded = next
	}

	normalized := strings.ReplaceAll(decoded, "\\", "/")
	if strings.Contains(normalized, "..") {
		return "", errors.New("path traversal detected")
	}

	clean := path.Clean(normalized)
	if clean == "." {
		clean = "/"
	}
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}
	return clean, nil
}

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
