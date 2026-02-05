# Architecture Billing - Clarification RGPD

## 🏗️ Architecture Technique vs Juridique

### Vue Technique
```
SaferCloud Core ──API──> SaferCloud-Billing ──API──> Mollie/Lago
   (public)                   (privé)                (tiers)
```

### Vue Juridique RGPD
```
┌──────────────────────────────────────────────────┐
│  RESPONSABLE DU TRAITEMENT                       │
│  SaferCloud SAS (SIRET: XXX)                     │
│                                                  │
│  ┌─────────────┐   ┌─────────────┐  ┌────────┐  │
│  │ Core (open) │   │ Billing     │  │ Lago   │  │
│  │             │◄─►│ (privé)     │◄►│(auto-  │  │
│  │             │   │             │  │hébergé)│  │
│  └─────────────┘   └─────────────┘  └────────┘  │
│                                                  │
│  = Composants internes (pas de DPA)             │
└──────────────────────┬───────────────────────────┘
                       │
                       ▼
                  ┌────────┐
                  │ Mollie │
                  │ (NL)   │
                  │        │
                  │ SOUS-  │
                  │ TRAIT. │
                  └────────┘
                  DPA requis
```

## 🎯 Clarifications Juridiques

### 1. SaferCloud-Billing N'EST PAS un Sous-Traitant

**Raison** : Même entité juridique (SaferCloud SAS)

**Analogie** :
- Avoir un microservice "auth" n'en fait pas un sous-traitant
- Avoir une base de données séparée n'en fait pas un sous-traitant
- Avoir un service "billing" séparé n'en fait pas un sous-traitant

**Base légale** : RGPD Article 4(8) définit le sous-traitant comme une "personne [...] qui traite des données à caractère personnel pour le compte du responsable du traitement"

→ Vous ne pouvez pas être sous-traitant de vous-même.

### 2. Lago Auto-Hébergé = Composant Interne

**Architecture** : Lago OSS déployé sur votre infrastructure OVH France

**Statut RGPD** :
- ✅ Composant interne (même contrôle que SaferCloud-Billing)
- ✅ Immunité Cloud Act (pas de SaaS US/NL)
- ✅ Souveraineté totale des données de facturation
- ❌ PAS de DPA nécessaire (votre propre infrastructure)

### 3. Mollie EST le Seul Sous-Traitant Externe

**Raison** : Service cloud externe (Mollie B.V., Pays-Bas)

**Obligations** :
- ✅ Signer DPA (Data Processing Agreement)
- ✅ Vérifier conformité RGPD et Clauses Contractuelles Types (SCC)
- ✅ Mentionner dans Politique de Confidentialité
- ✅ Documenter dans Registre des Traitements (Art 30)

## 📋 Checklist de Conformité

### Pour SaferCloud-Billing (Interne)
- [ ] Documenter dans Registre des Traitements
- [ ] Inclure dans AIPD globale
- [ ] Appliquer mêmes mesures de sécurité que le Core
- [ ] ❌ PAS besoin de DPA interne

### Pour Lago Auto-Hébergé (Interne)
- [ ] Déployer Lago OSS sur infrastructure OVH France
- [ ] Configurer Docker Compose avec réseau isolé
- [ ] Sécuriser connexion SaferCloud-Billing ↔ Lago (TLS interne)
- [ ] Documenter dans Registre des Traitements (traitement interne)
- [ ] ❌ PAS de DPA nécessaire (infrastructure contrôlée)

### Pour Mollie (Seul Sous-traitant Externe)
- [ ] ✅ Signer DPA Mollie : https://www.mollie.com/fr/legal/data-processing-agreement
- [ ] ✅ Vérifier Clauses Contractuelles Types (SCC)
- [ ] ✅ Mentionner dans Politique Confidentialité comme unique sous-traitant paiement
- [ ] ✅ Ajouter au Registre des Traitements

## 🔒 Bonne Pratique : Pseudonymisation

Même si Lago est auto-hébergé, **pseudonymiser les données avant transmission à Mollie** reste recommandé :

### Avantages
1. **Minimisation RGPD** (Art 5.1c) : Mollie n'a pas besoin de l'userID réel
2. **Réduction du risque** : En cas de breach chez Mollie (sous-traitant externe), pas de corrélation directe
3. **Facilite audit** : Démontre Privacy-by-Design
4. **Lago interne** : Peut utiliser userID réel sans risque (infrastructure contrôlée)

### Implémentation

```go
// backend/pkg/billing/privacy.go
type PrivacyPreservingBilling struct {
    provider BillingProvider
    pepper   [32]byte // Secret rotation mensuelle
}

func (p *PrivacyPreservingBilling) CreateSubscription(userID string) {
    // Pseudonymiser userID avant envoi à Mollie/Lago
    pseudonym := sha256(pepper + userID)[:16]

    // Stocker mapping en DB local (chiffré)
    storePseudonymMapping(pseudonym, userID)

    // Envoyer pseudonym à Mollie/Lago (pas l'userID réel)
    p.provider.CreateSubscription(pseudonym, planCode)
}
```

## 📄 Template Politique de Confidentialité

```markdown
### Sous-traitants

SaferCloud fait appel aux sous-traitants suivants pour assurer le bon fonctionnement du service :

#### Traitement des paiements
- **Mollie B.V.** (Pays-Bas)
- Données traitées : Email pseudonymisé, montant transaction
- Base légale : Exécution du contrat
- Garanties : DPA signé, SCC UE, ISO 27001

#### Hébergement et infrastructure
- **OVHcloud SAS** (France)
- Données traitées : Fichiers chiffrés, bases de données
- Base légale : Exécution du contrat
- Garanties : DPA automatique, SecNumCloud, souveraineté française

#### Composants internes (pas de sous-traitance)
- **Gestion de la facturation** : Lago OSS auto-hébergé sur OVH France
- **Authentification** : Supabase OSS auto-hébergé sur OVH France
- Ces composants sont sous contrôle direct de SaferCloud SAS et ne constituent pas des sous-traitants.

Tous les sous-traitants externes sont soumis à des accords de protection des données conformes au RGPD.
```

## ❓ FAQ Juridique

**Q: Si SaferCloud-Billing est dans un repo Git séparé, est-ce un sous-traitant ?**
R: Non, le découpage technique n'a pas d'impact juridique. Seule l'entité juridique compte.

**Q: Si SaferCloud-Billing est déployé sur un serveur différent ?**
R: Non, tant que c'est votre infrastructure (même SIRET), c'est interne.

**Q: Si j'engage un développeur freelance pour coder SaferCloud-Billing ?**
R: Le freelance est sous-traitant pour la **prestation de développement**, mais pas pour le **traitement des données** (c'est vous qui opérez le code).

**Q: Dois-je faire un DPA entre SaferCloud Core et SaferCloud-Billing ?**
R: Non, c'est absurde juridiquement. Vous ne signez pas de contrat avec vous-même.

**Q: Pourquoi Lago auto-hébergé n'est pas un sous-traitant alors que c'est du code open-source externe ?**
R: Le critère RGPD est "qui contrôle l'infrastructure". Lago OSS déployé sur vos serveurs OVH = contrôle total = composant interne. Lago SaaS hébergé par Lago Inc. = contrôle externe = sous-traitant.

**Q: Mollie peut-il voir les fichiers des utilisateurs ?**
R: Non, Mollie traite UNIQUEMENT les données de paiement (montant, email pseudonymisé). Les fichiers restent chiffrés sur OVH. Lago (auto-hébergé) gère uniquement les métriques d'usage.
