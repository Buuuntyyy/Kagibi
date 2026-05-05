<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="legal-container">
    <div class="legal-content">
      <h1>Politique de Confidentialité</h1>
      <p class="last-updated">Dernière mise à jour : 21 Avril 2026</p>

      <section>
        <h2>1. Introduction</h2>
        <p>Dans le cadre de son activité, Kagibi est amené à traiter des informations vous concernant. La présente Politique de Confidentialité vous informe de la manière dont nous recueillons et traitons vos données personnelles. Nous vous invitons à la lire attentivement.</p>
        <p>Nous attachons la plus grande importance au respect de la vie privée et des données individuelles de nos utilisateurs. Nous nous engageons à respecter les dispositions de la loi n°78-17 du 6 janvier 1978 relative à l'informatique, aux fichiers et aux libertés modifiée, du Règlement (UE) 2016/679 du 27 avril 2016 (<em>RGPD</em>), et de la loi n°2004-575 du 21 juin 2004 pour la confiance dans l'économie numérique (<em>LCEN</em>).</p>
      </section>

      <section>
        <h2>2. Responsable de traitement</h2>
        <p>Le responsable du traitement des données est Buuuntyyy, joignable par email : contact@kagibi.cloud.</p>
      </section>

      <section>
        <h2>3. Données collectées</h2>
        <p>Nous collectons et traitons les données suivantes :</p>
        <ul>
          <li><strong>Données d'identification :</strong> Nom d'utilisateur, adresse email, identifiant unique de compte.</li>
          <li><strong>Données de connexion :</strong> Adresse IP source, date et heure de connexion, type de navigateur et d'appareil. Ces données sont collectées en application de l'obligation légale prévue par l'article 6 II de la LCEN et le décret n°2021-1363.</li>
          <li><strong>Contenu chiffré :</strong> Métadonnées des fichiers et dossiers (nom, taille, date de modification). <strong>Le contenu des fichiers est chiffré côté client et Kagibi n'y a aucun accès.</strong></li>
          <li><strong>Données de signalisation P2P :</strong> Lors d'un transfert P2P, Kagibi enregistre temporairement les métadonnées nécessaires à la mise en relation technique : identifiant de l'expéditeur, email ou identifiant du destinataire (si renseigné), nom et taille du fichier, jeton d'invitation.</li>
          <li><strong>Flux TURN :</strong> Lorsque la connexion directe entre appareils n'est pas possible, le flux de données chiffré transite par le serveur TURN de Kagibi. Ce flux n'est <strong>jamais stocké</strong> : il est commuté en mémoire en temps réel entre les deux parties. Le contenu des fichiers reste chiffré de bout en bout (AES-256-GCM) et Kagibi ne dispose d'aucune clé permettant de le déchiffrer.</li>
          <li><strong>Données de facturation :</strong> En cas d'abonnement payant, les données de paiement sont traitées par notre prestataire Stripe (Stripe Inc.) et ne sont pas stockées par Kagibi.</li>
          <li><strong>Données de mesure d'audience :</strong> Kagibi utilise l'outil d'analyse Umami, hébergé sur sa propre infrastructure (sans transfert vers des tiers). Umami collecte des données anonymisées : pages visitées, URL de provenance (referrer), type de navigateur, système d'exploitation, résolution d'écran, et pays déduit de l'adresse IP par hachage irréversible. <strong>Aucune adresse IP brute n'est conservée et aucun cookie de traçage n'est déposé.</strong></li>
        </ul>
      </section>

      <section>
        <h2>4. Finalités du traitement</h2>
        <p>Vos données sont collectées pour les finalités suivantes :</p>
        <ul>
          <li>Fourniture et gestion du service Kagibi (création de compte, stockage, partage, transferts P2P).</li>
          <li>Sécurité du service et prévention de la fraude.</li>
          <li>Communication avec les utilisateurs (support technique, notifications importantes).</li>
          <li>Respect des obligations légales et réglementaires, notamment la conservation des données de connexion imposée par la LCEN.</li>
          <li>Mesure d'audience anonyme : compréhension de l'utilisation du service à des fins d'amélioration, sans identification des utilisateurs.</li>
        </ul>
      </section>

      <section>
        <h2>5. Destinataires des données</h2>
        <p>Les données collectées sont destinées aux équipes techniques de Kagibi. Elles peuvent être transmises aux sous-traitants auxquels Kagibi fait appel dans le cadre de l'exécution de ses services (hébergeur OVH France, prestataire de paiement Stripe). L'outil de mesure d'audience Umami est auto-hébergé sur l'infrastructure de Kagibi : aucune donnée d'audience n'est transmise à un tiers.</p>
        <p>Les données de connexion (logs, adresses IP) peuvent être communiquées aux autorités judiciaires ou administratives compétentes sur réquisition légale, conformément à la LCEN et au Code de procédure pénale.</p>
        <p>Kagibi ne vend ni ne loue vos données personnelles à des tiers à des fins de marketing.</p>
      </section>

      <section>
        <h2>6. Durée de conservation</h2>
        <p>Les durées de conservation sont les suivantes :</p>
        <ul>
          <li><strong>Données de connexion (logs, adresses IP) :</strong> 1 an à compter de la date d'enregistrement, conformément à l'obligation légale prévue par l'article 6 II de la LCEN et le décret n°2021-1363 du 24 novembre 2021.</li>
          <li><strong>Données de compte et contenu chiffré :</strong> Conservés pendant toute la durée d'activité du compte, puis supprimés des serveurs actifs dans un délai de 30 jours suivant la suppression du compte.</li>
          <li><strong>Métadonnées P2P (invitations, signaux de signalisation) :</strong> Les invitations P2P expirent automatiquement après 24 heures. Les métadonnées résiduelles (nom de fichier, taille, identifiants) sont supprimées dans un délai de 30 jours suivant la complétion ou l'expiration du transfert. Les signaux WebRTC sont supprimés immédiatement après consommation.</li>
          <li><strong>Données de facturation :</strong> Conservées conformément aux obligations comptables (10 ans) pour les factures, durée réduite pour les autres données de facturation.</li>
        </ul>
      </section>

      <section>
        <h2>7. Ce que Kagibi ne peut pas accéder</h2>
        <p>En raison de l'architecture de chiffrement de bout en bout de Kagibi :</p>
        <ul>
          <li>Kagibi ne peut pas lire le contenu de vos fichiers stockés (chiffrés côté client avec vos propres clés) ;</li>
          <li>Kagibi ne peut pas accéder au contenu des fichiers transférés via P2P, qu'ils transitent directement entre appareils ou par le serveur TURN de relais — le flux est chiffré de bout en bout et n'est jamais stocké ;</li>
          <li>Kagibi ne conserve aucune clé de déchiffrement de vos données.</li>
        </ul>
        <p>Cette architecture implique que la perte de vos clés de chiffrement peut entraîner la perte définitive d'accès à vos données. Kagibi ne peut être tenu responsable de cette perte.</p>
      </section>

      <section>
        <h2>8. Vos droits</h2>
        <p>Conformément à la réglementation applicable, vous disposez des droits suivants :</p>
        <ul>
          <li><strong>Droit d'accès :</strong> Vous pouvez demander l'accès à vos données personnelles.</li>
          <li><strong>Droit de rectification :</strong> Vous pouvez demander la correction de données inexactes ou incomplètes.</li>
          <li><strong>Droit à l'effacement :</strong> Vous pouvez demander la suppression de votre compte et de vos données ("droit à l'oubli"), sous réserve des obligations légales de conservation.</li>
          <li><strong>Droit à la limitation du traitement :</strong> Vous pouvez demander de geler l'utilisation de vos données.</li>
          <li><strong>Droit à la portabilité :</strong> Vous pouvez récupérer vos données dans un format structuré.</li>
          <li><strong>Droit d'opposition :</strong> Vous pouvez vous opposer au traitement de vos données pour des motifs légitimes.</li>
          <li><strong>Sort des données après la mort :</strong> Conformément à la loi "Informatique et Libertés", vous pouvez définir des directives relatives à la conservation, à l'effacement et à la communication de vos données après votre décès.</li>
        </ul>
        <p>Pour exercer ces droits, contactez-nous à : contact@kagibi.cloud.</p>
        <p>Si vous estimez, après nous avoir contactés, que vos droits "Informatique et Libertés" ne sont pas respectés, vous pouvez adresser une réclamation à la CNIL (www.cnil.fr).</p>
      </section>

      <section>
        <h2>9. Sécurité des données</h2>
        <p>Kagibi met en œuvre des mesures techniques et organisationnelles appropriées pour protéger vos données contre l'accès non autorisé, la modification, la divulgation ou la destruction, notamment : chiffrement AES-256-GCM des fichiers, chiffrement RSA-4096 des clés, communication via HTTPS/WSS, et authentification à deux facteurs disponible.</p>
      </section>

      <section>
        <h2>10. Cookies et mesure d'audience</h2>
        <p>Le site utilise des cookies techniques strictement nécessaires au fonctionnement du service (maintien de la session connectée). Ces cookies ne nécessitent pas de consentement préalable au sens de la directive ePrivacy.</p>
        <p>Pour la mesure d'audience, Kagibi utilise <strong>Umami</strong>, un outil d'analyse web respectueux de la vie privée, auto-hébergé sur son propre serveur. Umami ne dépose <strong>aucun cookie</strong> et ne collecte aucune donnée permettant d'identifier personnellement un visiteur. Les données collectées (pages vues, navigateur, pays approximatif) sont entièrement anonymisées et agrégées. Aucun transfert vers un service tiers n'est effectué. Ce traitement est fondé sur l'intérêt légitime de l'éditeur (art. 6.1.f RGPD) et est dispensé de consentement conformément aux recommandations de la CNIL relatives aux outils de mesure d'audience exemptés.</p>
      </section>
      
      <button @click="$router.go(-1)" class="btn-back">Retour</button>
    </div>
  </div>
</template>

<style scoped>
.legal-container {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 40px 20px;
  background-color: var(--background-color, #f5f7fa);
  min-height: 100vh;
  font-family: 'Segoe UI', sans-serif;
  color: #333;
}

.legal-content {
  background: white;
  padding: 40px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.05);
  max-width: 800px;
  width: 100%;
}

h1 {
  color: var(--primary-color, #42b983);
  margin-bottom: 10px;
}

h2 {
  font-size: 1.4rem;
  margin-top: 30px;
  margin-bottom: 15px;
  color: #2c3e50;
  border-bottom: 2px solid #f0f0f0;
  padding-bottom: 10px;
}

p, li {
  line-height: 1.6;
  color: #555;
  margin-bottom: 15px;
}

ul {
  padding-left: 20px;
  margin-bottom: 15px;
}

.last-updated {
  font-style: italic;
  color: #888;
  margin-bottom: 30px;
  border-bottom: 1px solid #eee;
  padding-bottom: 20px;
}

.btn-back {
  margin-top: 30px;
  padding: 10px 20px;
  background-color: var(--primary-color, #42b983);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}

.btn-back:hover {
  opacity: 0.9;
}

@media (max-width: 600px) {
  .legal-container {
    padding: 20px 12px;
  }

  .legal-content {
    padding: 20px 16px;
  }

  h1 {
    font-size: 1.5rem;
  }

  h2 {
    font-size: 1.15rem;
  }
}
</style>
