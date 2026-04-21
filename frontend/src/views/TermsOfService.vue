<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="legal-container">
    <div class="legal-content">
      <h1>Conditions Générales d'Utilisation (CGU)</h1>
      <p class="last-updated">Dernière mise à jour : 20 Avril 2026</p>

      <section>
        <h2>1. Objet</h2>
        <p>Les présentes Conditions Générales d'Utilisation (ci-après "CGU") ont pour objet de définir les modalités de mise à disposition des services du site et de l'application <strong>Kagibi</strong>, ci-après nommé "le Service", et les conditions d'utilisation du Service par l'Utilisateur.</p>
      </section>

      <section>
        <h2>2. Mentions Légales</h2>
        <p>L'édition du site Kagibi est assurée par Buuuntyyy.</p>
        <p>Adresse e-mail de contact : contact@kagibi.cloud</p>
        <p>L'hébergeur du Service est : OVH France.</p>
      </section>

      <section>
        <h2>3. Accès au Service</h2>
        <p>Le Service permet à l'Utilisateur un accès aux services suivants :</p>
        <ul>
          <li>Stockage de fichiers en ligne (Cloud) avec chiffrement de bout en bout ;</li>
          <li>Partage de fichiers et de dossiers via des liens sécurisés ;</li>
          <li>Transferts directs de fichiers entre utilisateurs (fonctionnalité P2P) via WebRTC chiffré.</li>
        </ul>
        <p>Le Service est accessible en tout lieu à tout Utilisateur ayant un accès à Internet. Tous les frais supportés pour l'accès au Service (matériel informatique, logiciels, connexion Internet, etc.) sont à la charge de l'Utilisateur.</p>
      </section>

      <section>
        <h2>4. Responsabilité de l'Utilisateur</h2>
        <p>L'Utilisateur est seul responsable des risques liés à l'utilisation de ses identifiants de connexion et de son mot de passe. En cas de divulgation de son mot de passe, l'Éditeur décline toute responsabilité.</p>
        <p>L'Utilisateur assume l'entière et exclusive responsabilité de l'utilisation qu'il fait du Service, ainsi que du contenu qu'il stocke, envoie ou reçoit via le Service.</p>
        <p>Il est formellement interdit à l'Utilisateur de stocker, partager ou transférer via le Service :</p>
        <ul>
          <li>Tout contenu illicite au regard de la législation française ou européenne ;</li>
          <li>Tout contenu portant atteinte aux droits de propriété intellectuelle de tiers ;</li>
          <li>Tout contenu à caractère pédopornographique, incitant à la haine ou à la violence ;</li>
          <li>Tout logiciel malveillant, virus ou tout autre programme informatique nuisible ;</li>
          <li>Tout contenu contrefait ou portant atteinte aux intérêts de tierces personnes.</li>
        </ul>
        <p>Tout manquement à ces obligations engage la responsabilité exclusive de l'Utilisateur et pourra entraîner la suspension ou la résiliation de son compte, sans préjudice de toute action judiciaire.</p>
        <p>Tout usage du Service par l'Utilisateur ayant directement ou indirectement pour conséquence des dommages causés à l'Éditeur ou à des tiers doit faire l'objet d'une indemnisation au profit de la partie lésée.</p>
      </section>

      <section>
        <h2>5. Responsabilité de l'Éditeur</h2>
        <p>Tout dysfonctionnement du serveur ou du réseau ne peut engager la responsabilité de l'Éditeur. De même, la responsabilité de l'Éditeur ne peut être engagée en cas de force majeure ou du fait imprévisible et insurmontable d'un tiers.</p>
        <p>Le Service Kagibi s'engage à mettre en œuvre tous les moyens nécessaires pour garantir la sécurité et la confidentialité des données. Toutefois, il ne peut garantir une sécurité absolue.</p>
      </section>

      <section>
        <h2>6. Stockage Chiffré — Limitation de Responsabilité</h2>
        <p>Les fichiers déposés sur le Service sont chiffrés de bout en bout, côté client, avant tout transfert vers les serveurs de Kagibi. Les clés de chiffrement sont générées et gérées exclusivement par l'Utilisateur sur son appareil.</p>
        <p><strong>Kagibi ne dispose d'aucune clé de déchiffrement et n'a techniquement aucun accès au contenu des fichiers stockés.</strong> En conséquence :</p>
        <ul>
          <li>L'Éditeur ne peut pas contrôler, modérer, ni vérifier la nature du contenu stocké ;</li>
          <li>L'Éditeur ne peut pas être tenu responsable du contenu stocké par l'Utilisateur ;</li>
          <li>L'Utilisateur est seul responsable de la licéité du contenu qu'il stocke.</li>
        </ul>
        <p>Nonobstant ce qui précède, l'Éditeur se réserve le droit de suspendre ou supprimer tout compte faisant l'objet d'un signalement d'activité illicite par les autorités compétentes, et coopérera pleinement avec toute réquisition judiciaire ou administrative dans le cadre de la législation applicable (notamment la LCEN et le Code de procédure pénale).</p>
        <p>Conformément à l'article 6 de la loi n°2004-575 du 21 juin 2004 pour la confiance dans l'économie numérique (LCEN), l'Éditeur agit en qualité d'hébergeur pour les fichiers stockés et bénéficie du régime de responsabilité limitée prévu par ce texte.</p>
      </section>

      <section>
        <h2>7. Transferts P2P — Responsabilité et Rôle de l'Éditeur</h2>
        <p>La fonctionnalité de transfert de fichiers en pair-à-pair (P2P) repose sur la technologie WebRTC. Dans ce cadre, Kagibi intervient en tant qu'intermédiaire technique selon deux modalités :</p>
        <ul>
          <li><strong>Signalisation :</strong> Kagibi facilite l'établissement de la connexion entre les appareils des utilisateurs (échange d'offres, réponses et candidats ICE). Ce trafic de signalisation ne contient aucun contenu de fichier.</li>
          <li><strong>Relais TURN :</strong> Lorsque la configuration réseau des utilisateurs (NAT, pare-feu) ne permet pas d'établir une connexion directe, le flux de données transite par un serveur TURN opéré par Kagibi. Ce serveur assure une <strong>commutation technique du flux sans aucun stockage</strong> : les données entrent et sortent en temps réel sans être écrites sur disque.</li>
        </ul>
        <p><strong>Dans tous les cas, le contenu des fichiers est chiffré de bout en bout (AES-256-GCM) avant d'être transmis.</strong> Le serveur TURN, comme les serveurs de signalisation, ne dispose d'aucune clé de déchiffrement et ne peut donc pas accéder au contenu des fichiers transférés.</p>
        <p>En conséquence :</p>
        <ul>
          <li>L'Éditeur ne peut pas lire, contrôler, modérer ni stocker le contenu des transferts P2P, qu'ils soient directs ou relayés ;</li>
          <li>L'Éditeur ne peut pas être tenu responsable du contenu transféré entre utilisateurs via la fonctionnalité P2P ;</li>
          <li>L'Éditeur agit en qualité de simple transporteur au sens de l'article 12 de la directive 2000/31/CE et de l'article 6 de la LCEN, et bénéficie du régime d'exonération de responsabilité correspondant.</li>
        </ul>
        <p><strong>L'Utilisateur expéditeur est seul et entièrement responsable du contenu qu'il transfère</strong>, notamment de sa licéité, de l'absence de violation de droits de tiers, et du consentement du destinataire.</p>
        <p>Avant tout transfert P2P, l'Utilisateur doit confirmer explicitement que le fichier qu'il souhaite transférer est légal et conforme aux présentes CGU. Cette confirmation constitue une déclaration sur l'honneur engageant la responsabilité civile et pénale de l'Utilisateur.</p>
        <p>Seules les métadonnées nécessaires à la signalisation (identifiant expéditeur, identifiant ou email destinataire, nom et taille du fichier, jeton d'invitation à durée limitée) sont temporairement enregistrées par Kagibi aux fins de mise en relation technique. Ces métadonnées sont supprimées dans les délais prévus par la Politique de Confidentialité.</p>
      </section>

      <section>
        <h2>8. Propriété Intellectuelle</h2>
        <p>Les marques, logos, signes ainsi que tous les contenus du site (textes, images, son...) font l'objet d'une protection par le Code de la propriété intellectuelle et plus particulièrement par le droit d'auteur.</p>
        <p>L'Utilisateur doit solliciter l'autorisation préalable du site pour toute reproduction, publication, copie des différents contenus. Il s'engage à une utilisation des contenus du site dans un cadre strictement privé, toute utilisation à des fins commerciales et publicitaires est strictement interdite.</p>
      </section>

      <section>
        <h2>9. Données Personnelles</h2>
        <p>L'Utilisateur dispose de droits d'opposition, d'accès, de rectification et de suppression portant sur les données personnelles le concernant, conformément à la Politique de Confidentialité du site.</p>
      </section>

      <section>
        <h2>10. Évolution des CGU</h2>
        <p>Le site Kagibi se réserve le droit de modifier les clauses de ces conditions générales d'utilisation à tout moment. En cas de modification substantielle, l'Utilisateur sera informé par notification ou par email. La poursuite de l'utilisation du Service après modification vaut acceptation des nouvelles CGU.</p>
      </section>

      <section>
        <h2>11. Droit applicable, médiation et juridiction compétente</h2>
        <p>La législation française s'applique au présent contrat. En cas d'absence de résolution amiable d'un litige né entre les parties, les tribunaux français seront seuls compétents pour en connaître.</p>
        <p>Pour tout litige de consommation qui n'aurait pu être résolu directement avec l'Éditeur, le consommateur peut recourir gratuitement à un médiateur de la consommation en vue de la résolution amiable du litige (Code de la consommation art. L.612-1). Les coordonnées du médiateur de référence sont : [NOM ET COORDONNÉES DU MÉDIATEUR].</p>
        <p>La Commission Européenne met à disposition une plateforme de règlement en ligne des litiges (RLL) : <a href="https://ec.europa.eu/consumers/odr" target="_blank">https://ec.europa.eu/consumers/odr</a>.</p>
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
