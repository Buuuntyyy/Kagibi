# Configuration du Billing SaferCloud

SaferCloud supporte 3 modes de facturation selon vos besoins :

## 🏠 Mode 1 : Self-Hosted (Sans facturation)

**Pour qui ?** Auto-hébergement personnel ou interne, pas de besoin de gestion d'abonnements.

**Configuration :**
```env
BILLING_ENABLED=false
```

**Caractéristiques :**
- ✅ Stockage illimité
- ✅ Bande passante illimitée
- ✅ Aucune interface d'abonnement affichée
- ✅ Aucune limite de quotas
- ✅ Parfait pour usage personnel ou entreprise

**Utilisation :** Idéal si vous hébergez SaferCloud en interne et ne voulez pas de système d'abonnement.

---

## 🧪 Mode 2 : MockProvider (Développement)

**Pour qui ?** Développement, tests, ou démonstration avec quotas basiques.

**Configuration :**
```env
BILLING_ENABLED=true
# Ne pas configurer BILLING_SERVICE_URL et BILLING_SERVICE_SECRET
```

**Caractéristiques :**
- 📦 Plan gratuit 5 Go de stockage
- 📶 10 Go de bande passante mensuelle
- ✅ Interface d'abonnement visible (mais non fonctionnelle)
- ⚠️ Quotas appliqués localement
- 🔬 Idéal pour tester l'UX du billing sans infrastructure

**Utilisation :** Mode par défaut pour découvrir SaferCloud ou développer sans service de billing externe.

---

## 🚀 Mode 3 : Production (Service externe)

**Pour qui ?** Déploiement en production avec vraie facturation (Lago + Mollie/Stripe).

**Configuration :**
```env
BILLING_ENABLED=true
BILLING_SERVICE_URL=https://billing.your-domain.com
BILLING_SERVICE_SECRET=your-64-char-hex-secret
```

**Caractéristiques :**
- 💳 Vraie facturation via Lago
- 💰 Paiements via Mollie/Stripe
- 📊 Gestion complète des abonnements
- 🧾 Génération de factures
- 📈 Tracking d'usage en temps réel
- 🔐 Communication HMAC-SHA256 sécurisée

**Prérequis :**
1. Déployer le service `SaferCloud-Billing` (repo privé)
2. Configurer Lago et Mollie dans le service billing
3. Générer un secret partagé : `openssl rand -hex 32`

**Utilisation :** Mode production pour monétiser SaferCloud via SaaS.

---

## 📊 Comparaison des modes

| Fonctionnalité | Self-Hosted | MockProvider | Production |
|----------------|-------------|--------------|-----------|
| Stockage | ♾️ Illimité | 5 Go | Selon plan |
| Bande passante | ♾️ Illimité | 10 Go/mois | Selon plan |
| UI Abonnement | ❌ Masquée | ✅ Visible | ✅ Visible |
| Factures | ❌ | ❌ | ✅ |
| Paiements | ❌ | ❌ | ✅ |
| Service externe | ❌ | ❌ | ✅ |

---

## 🔧 Comment choisir ?

### Choisissez **Self-Hosted** si :
- Vous hébergez pour un usage personnel
- Vous êtes une entreprise avec infrastructure interne
- Vous ne voulez pas de système d'abonnement
- Vous avez votre propre stockage

### Choisissez **MockProvider** si :
- Vous testez SaferCloud en local
- Vous développez des fonctionnalités
- Vous voulez voir l'UX du billing sans setup complexe
- Vous faites une démo

### Choisissez **Production** si :
- Vous lancez SaferCloud en SaaS
- Vous voulez monétiser l'application
- Vous avez besoin de gestion d'abonnements
- Vous voulez des paiements automatisés

---

## 🎨 Impact sur l'interface utilisateur

### Mode Self-Hosted (`BILLING_ENABLED=false`)
```
✗ Barre de stockage (sidebar) - MASQUÉE
✗ Section "Mettre à niveau" - MASQUÉE
✗ Page /billing - Inutile (pas d'accès)
✓ Utilisateurs ne voient aucune limite
```

### Mode MockProvider ou Production
```
✓ Barre de stockage (sidebar) - VISIBLE
✓ Section "Mettre à niveau" - VISIBLE
✓ Page /billing - ACCESSIBLE
✓ Quotas affichés
```

---

## 🚦 Démarrage rapide

### Pour développement local (MockProvider)
```bash
cd backend
cp .env.example .env
# Garder BILLING_ENABLED=true (par défaut)
# Ne rien configurer d'autre pour le billing
go run .
```

### Pour self-hosted (pas de billing)
```bash
cd backend
cp .env.example .env
echo "BILLING_ENABLED=false" >> .env
go run .
```

### Pour production
```bash
# 1. Déployer SaferCloud-Billing d'abord
# 2. Configurer SaferCloud Core
cd backend
cp .env.example .env
nano .env
# Ajouter :
# BILLING_ENABLED=true
# BILLING_SERVICE_URL=https://billing.your-domain.com
# BILLING_SERVICE_SECRET=$(openssl rand -hex 32)
go run .
```

---

## 📝 Logs de démarrage

Vous verrez dans les logs quel mode est actif :

```
[Billing] DISABLED - Self-hosted mode (unlimited storage)
[Billing] MockProvider initialized - free plan mode (5GB storage)
[Billing] WebhookProvider initialized - connected to external billing service
```

---

## ❓ FAQ

**Q: Puis-je changer de mode après déploiement ?**
R: Oui, il suffit de modifier la variable `BILLING_ENABLED` et redémarrer.

**Q: Le MockProvider enregistre-t-il vraiment l'usage ?**
R: Oui, mais seulement en mémoire (perdu au redémarrage).

**Q: Puis-je avoir plusieurs instances avec des modes différents ?**
R: Oui ! Production sur le domaine principal, self-hosted en interne.

**Q: Le mode self-hosted est-il vraiment illimité ?**
R: Oui, mais limité par votre stockage S3/disque physique.
